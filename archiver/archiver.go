//go:generate go tool yacc -o query.go -p SQ query.y
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
	"fmt"
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2/bson"
	"net"
	"os"
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
	tsdb                 TSDB
	store                MetadataStore
	manager              APIKeyManager
	objstore             ObjectStore
	republisher          *Republisher
	incomingcounter      *counter
	pendingwritescounter *counter
	coalescer            *TransactionCoalescer
	sshscs               *SSHConfigServer
	enforceKeys          bool
}

// Creates a new Archiver instance:
func NewArchiver(c *Config) *Archiver {

	logBackendLeveled := logging.AddModuleLevel(logBackend)
	// handle log level
	switch *c.Archiver.LogLevel {
	case "CRITICAL":
		logBackendLeveled.SetLevel(logging.CRITICAL, "")
	case "ERROR":
		logBackendLeveled.SetLevel(logging.ERROR, "")
	case "WARNING":
		logBackendLeveled.SetLevel(logging.WARNING, "")
	case "NOTICE":
		logBackendLeveled.SetLevel(logging.NOTICE, "")
	case "INFO":
		logBackendLeveled.SetLevel(logging.INFO, "")
	case "DEBUG":
		logBackendLeveled.SetLevel(logging.DEBUG, "")
	}

	logging.SetBackend(logBackendLeveled)
	logging.SetFormatter(logging.MustStringFormatter(format))

	// Configure Metadata store (+ object store)
	var store MetadataStore
	var manager APIKeyManager

	switch *c.Archiver.Metadata {
	case "mongo":
		// Mongo connection
		mongoaddr, err := net.ResolveTCPAddr("tcp4", *c.Mongo.Address+":"+*c.Mongo.Port)
		if err != nil {
			log.Fatal("Error parsing Mongo address: %v", err)
		}
		mongostore := NewMongoStore(mongoaddr)
		if mongostore == nil {
			log.Fatal("Error connection to MongoDB instance")
		}
		store = mongostore
		manager = mongostore
	case "venkman":
		log.Fatal("No support for venkman yet")
	default:
		log.Fatal(*c.Archiver.Metadata, " is not a recognized metadata store")
	}

	// Configure API key enforcement
	store.EnforceKeys(c.Archiver.EnforceKeys)

	// Configure Timeseries database
	var tsdb TSDB
	switch *c.Archiver.TSDB {
	/** connect to ReadingDB */
	case "readingdb":
		log.Fatal("No current support for ReadingDB")
		/** connect to Quasar */
	case "quasar":
		qsraddr, err := net.ResolveTCPAddr("tcp4", *c.Quasar.Address+":"+*c.Quasar.Port)
		if err != nil {
			log.Fatal("Error parsing Quasar address: %v", err)
		}
		tsdb = NewQuasarDB(qsraddr, *c.Archiver.MaxConnections)
		tsdb.AddStore(store)
		if tsdb == nil {
			log.Fatal("Error connecting to Quasar instance")
		}
	default:
		log.Fatal(c.Archiver.TSDB, " is not a valid timeseries database")
	}

	// Configure republisher
	republisher := NewRepublisher()
	republisher.store = store

	// Configure Object store
	var objstore ObjectStore
	switch *c.Archiver.Objects {
	case "mongo":
		// Mongo connection
		mongoaddr, err := net.ResolveTCPAddr("tcp4", *c.Mongo.Address+":"+*c.Mongo.Port)
		if err != nil {
			log.Fatal("Error parsing Mongo address: %v", err)
		}
		mongostore := NewMongoObjectStore(mongoaddr)
		if mongostore == nil {
			log.Fatal("Error connection to MongoDB instance")
		}
		objstore = mongostore
		objstore.AddStore(store)
	default:
		log.Fatal(*c.Archiver.Objects, " is not a recognized object store")
	}

	// Configure SSH server
	var sshscs *SSHConfigServer
	if c.SSH.Enabled {
		sshscs = NewSSHConfigServer(manager, *c.SSH.Port, *c.SSH.PrivateKey,
			*c.SSH.AuthorizedKeysFile,
			*c.SSH.User, *c.SSH.Pass,
			c.SSH.PasswordEnabled, c.SSH.KeyAuthEnabled)
		go sshscs.Listen()
	}
	return &Archiver{tsdb: tsdb,
		store:                store,
		objstore:             objstore,
		manager:              manager,
		republisher:          republisher,
		incomingcounter:      newCounter(),
		pendingwritescounter: newCounter(),
		coalescer:            NewTransactionCoalescer(&tsdb, &store),
		sshscs:               sshscs,
		enforceKeys:          c.Archiver.EnforceKeys}

}

// Takes a map of string/SmapMessage (path, sMAP JSON object) and commits them to
// the underlying databases. First, checks that write permission is granted with the accompanied
// apikey (generated with the gilescmd CLI tool), then saves the metadata, pushes the readings
// out to any concerned republish clients, and commits the reading to the timeseries database.
// Returns an error, which is nil if all went well
func (a *Archiver) AddData(readings map[string]*SmapMessage, apikey string) error {
	if a.enforceKeys {
		ok, err := a.store.CheckKey(apikey, readings)
		if err != nil {
			log.Error("Error checking API key %v: %v", apikey, err)
			return err
		}
		if !ok {
			return errors.New("Unauthorized api key " + apikey)
		}
	}
	// save metadata
	a.store.SaveTags(readings)
	// if any of these are NOT nil, then we signal the republisher
	// that some metadata may have changed
	for _, rdg := range readings {
		if rdg.Metadata != nil ||
			rdg.Properties != nil ||
			rdg.Actuator != nil {
			a.republisher.MetadataChange(rdg)
		}
	}
	for _, msg := range readings {
		a.republisher.Republish(msg)
		a.incomingcounter.Mark()
		if msg.Readings == nil {
			continue
		}
		if a.store.GetStreamType(msg.UUID) == OBJECT_STREAM {
			_, err := a.objstore.AddObject(msg)
			if err != nil {
				return err
			}
		} else {
			a.coalescer.AddSmapMessage(msg)
		}
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
func (a *Archiver) HandleQuery(querystring, apikey string) ([]byte, error) {
	var data []byte
	var res []interface{}
	var err error
	if apikey != "" {
		log.Info("query with key: %v", apikey)
	}
	log.Info(querystring)
	lex := Parse(querystring)
	if lex.error != nil {
		return data, fmt.Errorf("Error (%v) in query \"%v\" (error at %v)\n", lex.error.Error(), querystring, lex.lasttoken)
	}
	log.Debug("query %v", lex.query)
	switch lex.query.qtype {
	case SELECT_TYPE:
		target := lex.query.ContentsBson()
		if lex.query.distinct {
			if len(target) != 1 {
				return data, fmt.Errorf("Distinct query can only use one tag\n")
			}
			res, err = a.store.GetTags(target, true, lex.query.Contents[0], lex.query.WhereBson())
		} else {
			res, err = a.store.GetTags(target, false, "", lex.query.WhereBson())
		}

		if err != nil {
			return data, err
		}
		data, _ = json.Marshal(res)
	case DELETE_TYPE:
		var (
			err error
			res bson.M
		)
		if len(lex.query.Contents) > 0 { // RemoveTags
			res, err = a.store.RemoveTags(lex.query.ContentsBson(), apikey, lex.query.WhereBson())
		} else { // RemoveDocs
			res, err = a.store.RemoveDocs(apikey, lex.query.WhereBson())
		}
		a.republisher.MetadataChangeKeys(lex.keys)
		log.Info("results %v", res)
		if err != nil {
			return data, err
		}
		data, _ = json.Marshal(res)
	case SET_TYPE:
		res, err := a.store.UpdateTags(lex.query.SetBson(), apikey, lex.query.WhereBson())
		if err != nil {
			return data, err
		}
		a.republisher.MetadataChangeKeys(lex.keys)
		data, _ = json.Marshal(res)
	case DATA_TYPE:
		// grab reference to the data query
		dq := lex.query.data

		// fetch all possible UUIDs that match the query
		uuids, err := a.GetUUIDs(lex.query.WhereBson())
		if err != nil {
			return data, err
		}

		// limit number of streams
		if dq.limit.streamlimit > 0 && len(uuids) > 0 {
			uuids = uuids[:dq.limit.streamlimit]
		}

		var response []SmapReading
		start := uint64(dq.start.UnixNano())
		end := uint64(dq.end.UnixNano())
		switch dq.dtype {
		case IN_TYPE:
			log.Debug("Data in start %v end %v", start, end)
			if start < end {
				response, err = a.GetData(uuids, start, end, UOT_NS, dq.timeconv)
			} else {
				response, err = a.GetData(uuids, end, start, UOT_NS, dq.timeconv)
			}
		case BEFORE_TYPE:
			log.Debug("Data before time %v", start)
			response, err = a.PrevData(uuids, start, int32(dq.limit.limit), UOT_NS, dq.timeconv)
		case AFTER_TYPE:
			log.Debug("Data after time %v", start)
			response, err = a.NextData(uuids, start, int32(dq.limit.limit), UOT_NS, dq.timeconv)
		}
		log.Debug("response %v uuids %v", response, uuids)
		data, _ = json.Marshal(response)
	}
	return data, nil
}

// For each of the streamids, fetches all data between start and end (where
// start < end). The units for start/end are given by query_uot. We give the units
// so that each time series database can convert the incoming timestamps to whatever
// it needs (most of these will query the metadata store for the unit of time for the
// data stream it is accessing)
func (a *Archiver) GetData(streamids []string, start, end uint64, query_uot, to_uot UnitOfTime) ([]SmapReading, error) {
	resp, err := a.tsdb.GetData(streamids, start, end, query_uot)
	if err == nil { // if no error, adjust timeseries
		for i, sr := range resp {
			stream_uot := a.store.GetUnitOfTime(sr.UUID)
			if len(sr.Readings) == 0 && a.store.GetStreamType(sr.UUID) == OBJECT_STREAM {
				newrdg, err := a.objstore.GetObjects(sr.UUID, start, end, query_uot)
				if err != nil {
					return resp, err
				}
				sr = newrdg
			}
			for j, reading := range sr.Readings {
				reading[0] = float64(convertTime(uint64(reading[0].(float64)), stream_uot, to_uot))
				sr.Readings[j] = reading
			}
			resp[i] = sr
		}
	}
	return resp, err
}

// For each of the streamids, fetches data before the start time. If limit is < 0, fetches all data.
// If limit >= 0, fetches only that number of points. See Archiver.GetData for explanation of query_uot
func (a *Archiver) PrevData(streamids []string, start uint64, limit int32, query_uot, to_uot UnitOfTime) ([]SmapReading, error) {
	resp, err := a.tsdb.Prev(streamids, start, limit, query_uot)
	if err == nil { // if no error, adjust timeseries
		for i, sr := range resp {
			stream_uot := a.store.GetUnitOfTime(sr.UUID)
			// if no readings from timeseries database, it might be an object stream
			if len(sr.Readings) == 0 && a.store.GetStreamType(sr.UUID) == OBJECT_STREAM {
				newrdg, err := a.objstore.PrevObject(sr.UUID, start, query_uot)
				if err != nil {
					return resp, err
				}
				sr = newrdg
			}
			for j, reading := range sr.Readings {
				reading[0] = float64(convertTime(uint64(reading[0].(float64)), stream_uot, to_uot))
				sr.Readings[j] = reading
			}
			resp[i] = sr
		}
	}
	return resp, err
}

// For each of the streamids, fetches data after the start time. If limit is < 0, fetches all data.
// If limit >= 0, fetches only that number of points. See Archiver.GetData for explanation of query_uot
func (a *Archiver) NextData(streamids []string, start uint64, limit int32, query_uot, to_uot UnitOfTime) ([]SmapReading, error) {
	resp, err := a.tsdb.Next(streamids, start, limit, query_uot)
	if err == nil { // if no error, adjust timeseries
		for i, sr := range resp {
			stream_uot := a.store.GetUnitOfTime(sr.UUID)
			// if no readings from timeseries database, it might be an object stream
			if len(sr.Readings) == 0 && a.store.GetStreamType(sr.UUID) == OBJECT_STREAM {
				newrdg, err := a.objstore.NextObject(sr.UUID, start, query_uot)
				if err != nil {
					return resp, err
				}
				sr = newrdg
			}
			for j, reading := range sr.Readings {
				reading[0] = float64(convertTime(uint64(reading[0].(float64)), stream_uot, to_uot))
				sr.Readings[j] = reading
			}
			resp[i] = sr
		}
	}
	return resp, err
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
func (a *Archiver) TagsUUID(uuid string) (bson.M, error) {
	return a.store.UUIDTags(uuid)
}

// For all streams that match the WHERE clause in the provided query string,
// will push all subsequent incoming information (data and tags) on those streams
// to the client associated with the provided http.ResponseWriter.
func (a *Archiver) HandleSubscriber(s Subscriber, query, apikey string) {
	a.republisher.HandleSubscriber(s, query, apikey)
}

// For all streams that match the provided where clause in where_tags, sets the key-value
// pairs specified in update_tags.
func (a *Archiver) SetTags(update_tags, where_tags map[string]interface{}, apikey string) (int, error) {
	res, err := a.store.UpdateTags(update_tags, apikey, where_tags)
	return res["Updated"].(int), err
}

func (a *Archiver) PrintStatus() {
	go periodicCall(1*time.Second, a.status) // status from stats.go
}
