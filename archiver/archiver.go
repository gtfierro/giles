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
package archiver

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2/bson"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

var log = logging.MustGetLogger("archiver")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

// This is the central object for the archiver process and contains most of the requisite
// logic for the core features of the archiver. One of the focuses of Giles is to facilitate
// adapting the sMAP protocol to different interfaces; the handlers packages (HTTP, WS, etc)
// provide handler functions that in turn call the archiver's core functions. Most of these
// core functions use easily usable data formats (such as bson.M), so the handler functions just
// have to deal with translating data formats
//
// For now, because the metadata interface was designed with a MongoDB backend in mind, most of
// the in-transit data types for dealing with metadata use the MongoDB interface defined by
// http://godoc.org/gopkg.in/mgo.v2/bson and http://godoc.org/gopkg.in/mgo.v2. I suggest
// taking a quick look though their documentation and how they talk to Mongo to get a feel
// for what the incoming/outgoing data is going to look like.
type Archiver struct {
	address              *net.TCPAddr
	tsdb                 TSDB
	store                *Store
	republisher          *Republisher
	incomingcounter      *counter
	pendingwritescounter *counter
	R                    *httprouter.Router
}

// Creates a new Archiver instance:
func NewArchiver(c *Config) *Archiver {
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	store := NewStore(c.MongoAddress)
	if store == nil {
		log.Fatal("Error connection to MongoDB instance")
	}

	var tsdb TSDB
	switch c.TSDB {
	case "readingdb":
		/** connect to ReadingDB */
		tsdb = NewReadingDB(c.TSDBAddress, c.Keepalive)
		tsdb.AddStore(store)
		if tsdb == nil {
			log.Fatal("Error connecting to ReadingDB instance")
		}
	case "quasar":
		tsdb = NewQuasar(c.TSDBAddress, c.Keepalive)
		tsdb.AddStore(store)
		if tsdb == nil {
			log.Fatal("Error connecting to Quasar instance")
		}
	default:
		log.Fatal(c.TSDB, " is not a valid timeseries database")
	}
	republisher := NewRepublisher()
	republisher.store = store
	address, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:"+strconv.Itoa(c.Port))
	if err != nil {
		log.Fatal("Error resolving address %v: %v", "0.0.0.0:"+strconv.Itoa(c.Port), err)
	}
	return &Archiver{tsdb: tsdb,
		store:                store,
		republisher:          republisher,
		address:              address,
		R:                    httprouter.New(),
		incomingcounter:      newCounter(),
		pendingwritescounter: newCounter()}

}

// Serves all registered endpoints. Doesn't return, so you might want to call this with 'go archiver.Serve()'
func (a *Archiver) Serve() {
	http.Handle("/", a.R)
	log.Notice("Starting on %v", a.address.String())

	srv := &http.Server{
		Addr: a.address.String(),
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
		a.incomingcounter.Mark()
		if msg.Readings == nil {
			continue
		}
		a.tsdb.Add(msg.Readings)
	}
	return nil
}

// Takes the body of the query and the apikey that accompanies the query. First parses
// the string query into an intermediary form (the abstract syntax tree as the AST type).
// Depending on the action, it will check to see if the provided API key grants sufficient
// permission to return the results. If so, returns those results as []byte (marshaled JSON).
// Most of this method is just switch statements dependent on different components of the
// generated AST. Any actual computation is done as calls to the Archiver API, so if you want
// to use your own query language or handle queries in some external handler, then you shouldn't
// need to use any of this method; just use the Archiver API
//
//TODO: expand the query language to allow the specification of units of time
func (a *Archiver) HandleQuery(querystring, apikey string) ([]byte, error) {
	if apikey != "" {
		log.Info("query with key: %v", apikey)
	}
	log.Info(querystring)
	var data []byte
	ast := parse(querystring)
	where := ast.Where.ToBson()
	switch ast.TargetType {
	// if we are fetching tags
	case TAGS_TARGET:
		bson_target := ast.Target.(*tagsTarget).ToBson()
		distinct_key := ast.Target.(*tagsTarget).Contents[0]
		is_distinct := ast.Target.(*tagsTarget).Distinct
		res, err := a.store.GetTags(bson_target, is_distinct, distinct_key, where)
		if err != nil {
			return data, err
		}
		data, _ = json.Marshal(res)
		// if we are setting tags
	case SET_TARGET:
		res, err := a.store.SetTags(ast.Target.(*setTarget).Updates, apikey, where)
		if err != nil {
			return data, err
		}
		data, _ = json.Marshal(res)
	// if we are fetching data
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
			response, err = a.GetData(uuids, start, end, UOT_MS)
		case AFTER:
			ref := uint64(target.Ref.Unix())
			log.Debug("after %v", ref)
			response, err = a.NextData(uuids, ref, target.Limit, UOT_MS)
		case BEFORE:
			ref := uint64(target.Ref.Unix())
			log.Debug("before %v", ref)
			response, err = a.PrevData(uuids, ref, target.Limit, UOT_MS)
		}
		data, _ = json.Marshal(response)
	}
	return data, nil
}

// For each of the streamids, fetches all data between start and end (where
// start < end). The units for start/end are given by query_uot. We give the units
// so that each time series database can convert the incoming timestamps to whatever
// it needs (most of these will query the metadata store for the unit of time for the
// data stream it is accessing)
func (a *Archiver) GetData(streamids []string, start, end uint64, query_uot UnitOfTime) ([]SmapResponse, error) {
	return a.tsdb.GetData(streamids, start, end, query_uot)
}

// For each of the streamids, fetches data before the start time. If limit is < 0, fetches all data.
// If limit >= 0, fetches only that number of points. See Archiver.GetData for explanation of query_uot
func (a *Archiver) PrevData(streamids []string, start uint64, limit int32, query_uot UnitOfTime) ([]SmapResponse, error) {
	return a.tsdb.Prev(streamids, start, limit, query_uot)
}

// For each of the streamids, fetches data after the start time. If limit is < 0, fetches all data.
// If limit >= 0, fetches only that number of points. See Archiver.GetData for explanation of query_uot
func (a *Archiver) NextData(streamids []string, start uint64, limit int32, query_uot UnitOfTime) ([]SmapResponse, error) {
	return a.tsdb.Next(streamids, start, limit, query_uot)
}

// For all streams that match the provided where clause in where_tags, returns the values of the requested
// tags. where_tags is a bson.M object that follows the same syntax as a MongoDB query. select_tags is
// a map[string]int corresponding to which tags we wish returned. A value of 1 means the tag will be
// returned (and ignores all other tags), and a value of 0 means the tag will NOT be returned (and all
// other tags will be).
func (a *Archiver) GetTags(select_tags, where_tags bson.M) ([]interface{}, error) {
	return a.store.GetTags(select_tags, false, "", where_tags)
}

// Returns a list of UUIDs for all streams that match the provided 'where' clause. where_tags is a bson.M
// object that follows the same syntax as a MongoDB query. This query is executed against the underlying
// metadata store. As we move into supporting multiple possible metadata storage solutions, this interface
// may change.
func (a *Archiver) GetUUIDs(where_tags bson.M) ([]string, error) {
	return a.store.GetUUIDs(where_tags)
}

// Returns all tags for the stream with the provided UUID
func (a *Archiver) TagsUUID(uuid string) ([]bson.M, error) {
	return a.store.TagsUUID(uuid)
}

// For all streams that match the WHERE clause in the provided query string,
// will push all subsequent incoming information (data and tags) on those streams
// to the client associated with the provided http.ResponseWriter.
//
// For now, this query is evaluated only once at the time of subscription.
func (a *Archiver) HandleSubscriber(rw http.ResponseWriter, query string) {
	a.republisher.HandleSubscriber(rw, string(query))
}

// For all streams that match the provided where clause in where_tags, sets the key-value
// pairs specified in update_tags.
func (a *Archiver) SetTags(update_tags, where_tags map[string]interface{}, apikey string) (int, error) {
	res, err := a.store.SetTags(update_tags, apikey, where_tags)
	return res["Updated"].(int), err
}

func (a *Archiver) PrintStatus() {
	go periodicCall(5*time.Second, a.status) // status from stats.go
}
