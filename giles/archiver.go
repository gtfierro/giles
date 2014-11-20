// License stuff

// Package giles implements an archiver that follows the sMAP protocol
//
// Overview
//
// Part of the motivation for the creation of Giles was to emphasize the
// distinction between sMAP the software (originally written in Python) and
// sMAP the profile. The Giles archiver is an implementation of the latter,
// and is intended to be fully compatible with existing sMAP tools.
//
// One of the "innovations" that Giles brings to the sMAP ecosystem is the
// notion that what is typically thought of as the sMAP "archiver" is really
// a collection of components: the message bus/frontend, the timeseries store,
// the metadata store, and the query language. All of these are closely linked,
// of course, but treating them as separate entities means that we can use
// different timeseries or metadata databases or even different implementations
// of the query language (perhaps over Apache Spark/Mlib?)
package giles

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"os"
	"time"
)

var log = logging.MustGetLogger("archiver")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

// This is the central object for the archiver process and contains most of the requisite
// logic for the core features of the archiver. One of the focuses of Giles is to facilitate
// adapting the sMAP protocol to different interfaces; the X_handlers.go files (TCP, WS, etc)
// provide handler functions that in turn call the archiver's core functions. Most of these
// core functions use easily usable data formats (such as bson.M), so the handler functions just
// have to deal with translating data formats
type Archiver struct {
	address              string
	tsdb                 TSDB
	store                *Store
	republisher          *Republisher
	incomingcounter      *Counter
	pendingwritescounter *Counter
	R                    *mux.Router
}

// Creates a new Archiver instance:
//   - archiverport: HTTP port on which to serve the archiver (default is 8079)
//   - tsdbstring: which timeseries database we are using (default is 'readingdb')
//   - tsdbip, tsdbport: address for instance of timeseries database (default is 'localhost:4242')
//   - tsdbkeepalive: number of seconds to maintain a connection open to the timeseries database
//     for a given unique identifier (see information on Pool)
//   - mongoip, mongoport: address for MongoDB instance, used for metadata, API keys, etc
func NewArchiver(archiverport int, tsdbip string, tsdbport int, mongoip string,
	mongoport int, tsdbstring string, tsdbkeepalive int, address string) *Archiver {
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	store := NewStore(mongoip, mongoport)
	if store == nil {
		log.Fatal("Error connection to MongoDB instance")
	}

	var tsdb TSDB
	switch tsdbstring {
	case "readingdb":
		/** connect to ReadingDB */
		tsdb = NewReadingDB(tsdbip, tsdbport, tsdbkeepalive)
		tsdb.AddStore(store)
		if tsdb == nil {
			log.Fatal("Error connecting to ReadingDB instance")
		}
	case "quasar":
		log.Fatal("quasar")
	default:
		log.Fatal(tsdbstring, " is not a valid timeseries database")
	}
	republisher := NewRepublisher()
	republisher.store = store
	return &Archiver{tsdb: tsdb,
		store:                store,
		republisher:          republisher,
		address:              address,
		R:                    mux.NewRouter(),
		incomingcounter:      NewCounter(),
		pendingwritescounter: NewCounter()}

}

// Creates routes for WebSocket endpoints. These are the same as the normal HTTP/TCP endpoints, but are
// preceeded with '/ws/`. Not served until Archiver.Serve() is called.
//func (a *Archiver) HandleWebSocket() {
//	log.Debug("Hanadling WebSockets")
//	a.R.HandleFunc("/ws/api/query", curryhandler(a, WsQueryHandler)).Methods("POST")
//	a.R.HandleFunc("/ws/tags/uuid", curryhandler(a, WsTagsHandler)).Methods("GET")
//	a.R.HandleFunc("/ws/tags/uuid/{uuid}", curryhandler(a, WsTagsHandler)).Methods("GET")
//}

// Serves all registered endpoints. Doesn't return, so you might want to call this with 'go archiver.Serve()'
func (a *Archiver) Serve() {
	http.Handle("/", a.R)
	log.Notice("Starting on %v", a.address)

	srv := &http.Server{
		Addr: a.address,
	}
	srv.ListenAndServe()
}

// Takes a map of string/SmapMessage (path, sMAP JSON object) and commits them to
// the underlying databases. First, checks that write permission is granted with the accompanied
// apikey (generated with the gilescmd CLI tool), then saves the metadata, pushes the readings
// out to any concerned republish clients, and commits the reading to the timeseries database.
// Returns an error, which is nil if all went well
func (a *Archiver) AddData(readings map[string]*SmapMessage, apikey string) error {
	ok, err := a.store.CheckKey(apikey, readings)
	if err != nil {
		log.Error("Error checking API key %v: %v", apikey, err)
		return err
	}
	if !ok {
		return errors.New("Unauthorized api key " + apikey)
	}
	a.store.SavePathMetadata(&readings)
	for _, msg := range readings {
		go a.store.SaveMetadata(msg)
		go a.republisher.Republish(msg)
		a.tsdb.Add(msg.Readings)
		a.incomingcounter.Mark()
	}
	return nil
}

// Takes the body of the query and the apikey that accompanies the query.
func (a *Archiver) HandleQuery(querystring, apikey string) ([]byte, error) {
	if apikey != "" {
		log.Info("query with key: %v", apikey)
	}
	log.Info(querystring)
	var data []byte
	ast := parse(querystring)
	where := ast.Where.ToBson()
	switch ast.TargetType {
	case TAGS_TARGET:
		bson_target := ast.Target.(*tagsTarget).ToBson()
		distinct_key := ast.Target.(*tagsTarget).Contents[0]
		is_distinct := ast.Target.(*tagsTarget).Distinct
		res, err := a.store.GetTags(bson_target, is_distinct, distinct_key, where)
		if err != nil {
			return data, err
		}
		data, _ = json.Marshal(res)
	case SET_TARGET:
		res, err := a.store.SetTags(ast.Target.(*setTarget).Updates, apikey, where)
		if err != nil {
			return data, err
		}
		data, _ = json.Marshal(res)
	case DATA_TARGET:
		target := ast.Target.(*dataTarget)
		uuids, err := a.GetUUIDs(ast.Where.ToBson())
		if err != nil {
			return data, err
		}
		if target.Streamlimit > -1 {
			uuids = uuids[:target.Streamlimit] // limit number of streams
		}
		var response []SmapResponse
		switch target.Type {
		case IN:
			start := uint64(target.Start.Unix())
			end := uint64(target.End.Unix())
			log.Debug("start %v end %v", start, end)
			response, err = a.GetData(uuids, start, end)
		case AFTER:
			ref := uint64(target.Ref.Unix())
			log.Debug("after %v", ref)
			response, err = a.NextData(uuids, ref, target.Limit)
		case BEFORE:
			ref := uint64(target.Ref.Unix())
			log.Debug("before %v", ref)
			response, err = a.PrevData(uuids, ref, target.Limit)
		}
		data, _ = json.Marshal(response)
	}
	return data, nil
}

func (a *Archiver) GetData(streamids []string, start, end uint64) ([]SmapResponse, error) {
	return a.tsdb.GetData(streamids, start, end)
}

func (a *Archiver) PrevData(streamids []string, start uint64, limit int32) ([]SmapResponse, error) {
	return a.tsdb.Prev(streamids, start, limit)
}

func (a *Archiver) NextData(streamids []string, start uint64, limit int32) ([]SmapResponse, error) {
	return a.tsdb.Next(streamids, start, limit)
}

func (a *Archiver) GetTags(select_tags, where_tags map[string]interface{}) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (a *Archiver) GetUUIDs(where_tags bson.M) ([]string, error) {
	return []string{}, nil
}

func (a *Archiver) TagsUUID(uuid string) ([]bson.M, error) {
	return []bson.M{}, nil
}

func (a *Archiver) HandleSubscriber(rw http.ResponseWriter, query string) {
	a.republisher.HandleSubscriber(rw, string(query))
}

func (a *Archiver) SetTags(update_tags, where_tags map[string]interface{}) (int, error) {
	return 0, nil
}

func (a *Archiver) PrintStatus() {
	go periodicCall(1*time.Second, a.status) // status from stats.go
}
