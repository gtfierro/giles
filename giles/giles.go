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

//TODO: probably name this one 'archiver' and rename 'archiver.go' to 'giles.go'

type Archiver struct {
	address              string
	tsdb                 TSDB
	store                *Store
	republisher          *Republisher
	incomingcounter      *Counter
	pendingwritescounter *Counter
}

func NewArchiver(archiverport int, readingdbip string, readingdbport int, mongoip string,
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
		tsdb = NewReadingDB(readingdbip, readingdbport, tsdbkeepalive)
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
		incomingcounter:      NewCounter(),
		pendingwritescounter: NewCounter()}

}

// Serves HTTP endpoints
func (a *Archiver) ServeHTTP() {
	log.Debug("serving http")
	r := mux.NewRouter()
	r.HandleFunc("/add/{key}", curryhandler(a, AddReadingHandler)).Methods("POST")
	r.HandleFunc("/republish", curryhandler(a, RepublishHandler)).Methods("POST")
	r.HandleFunc("/api/query", curryhandler(a, QueryHandler)).Queries("key", "{key:[A-Za-z0-9-_=%]+}").Methods("POST")
	r.HandleFunc("/api/query", curryhandler(a, QueryHandler)).Methods("POST")
	r.HandleFunc("/api/tags/uuid/{uuid}", curryhandler(a, TagsHandler)).Methods("GET")
	r.HandleFunc("/api/{method}/uuid/{uuid}", curryhandler(a, DataHandler)).Methods("GET")

	//r.HandleFunc("/ws/api/query", WsQueryHandler).Methods("POST")
	//r.HandleFunc("/ws/tags/uuid", WsTagsHandler).Methods("GET")
	//r.HandleFunc("/ws/tags/uuid/{uuid}", WsTagsHandler).Methods("GET")
	http.Handle("/", r)
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
		bson_target := ast.Target.(*TagsTarget).ToBson()
		distinct_key := ast.Target.(*TagsTarget).Contents[0]
		is_distinct := ast.Target.(*TagsTarget).Distinct
		res, err := a.store.GetTags(bson_target, is_distinct, distinct_key, where)
		if err != nil {
			return data, err
		}
		data, _ = json.Marshal(res)
	case SET_TARGET:
		res, err := a.store.SetTags(ast.Target.(*SetTarget).Updates, apikey, where)
		if err != nil {
			return data, err
		}
		data, _ = json.Marshal(res)
	case DATA_TARGET:
		target := ast.Target.(*DataTarget)
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

func (a *Archiver) SetTags(update_tags, where_tags map[string]interface{}) (int, error) {
	return 0, nil
}

func (a *Archiver) PrintStatus() {
	go periodicCall(1*time.Second, a.status) // status from stats.go
}
