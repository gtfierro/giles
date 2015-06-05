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
	"errors"
	"fmt"
	"github.com/gtfierro/giles/internal/tree"
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2/bson"
	"io"
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
	qp                   *QueryProcessor
	republisher          *Republisher
	incomingcounter      *counter
	pendingwritescounter *counter
	coalescer            *TransactionCoalescer
	sshscs               *SSHConfigServer
	enforceKeys          bool
}

// Creates a new Archiver instance:
func NewArchiver(c *Config) (a *Archiver) {

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
	a = &Archiver{tsdb: tsdb,
		store:                store,
		objstore:             objstore,
		manager:              manager,
		incomingcounter:      newCounter(),
		pendingwritescounter: newCounter(),
		coalescer:            NewTransactionCoalescer(&tsdb, &store),
		sshscs:               sshscs,
		enforceKeys:          c.Archiver.EnforceKeys}

	// Configure query processor
	qp := NewQueryProcessor(a)
	a.qp = qp

	// Configure republisher
	republisher := NewRepublisher(a)
	a.republisher = republisher
	return
}

// Takes a map of string/SmapMessage (path, sMAP JSON object) and commits them to
// the underlying databases. First, checks that write permission is granted with the accompanied
// apikey (generated with the gilescmd CLI tool), then saves the metadata, pushes the readings
// out to any concerned republish clients, and commits the reading to the timeseries database.
// Returns an error, which is nil if all went well
func (a *Archiver) AddData(readings map[string]*SmapMessage, apikey string) error {
	var (
		//pathMdErr error
		tsMdErr error
	)
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
	//pathMdErr = a.store.SavePathMetadata(readings)
	//if pathMdErr != nil {
	//	return pathMdErr
	//}

	//tsMdErr = a.store.SaveTimeseriesMetadata(readings)
	//if tsMdErr != nil {
	//	return tsMdErr
	//}

	tsMdErr = a.store.SaveTags(readings)
	if tsMdErr != nil {
		return tsMdErr
	}

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
func (a *Archiver) HandleQuery(querystring, apikey string) (interface{}, error) {
	var res interface{}
	var err error
	if apikey != "" {
		log.Info("query with key: %v", apikey)
	}
	log.Info(querystring)
	lex := a.qp.Parse(querystring)
	if lex.error != nil {
		return res, fmt.Errorf("Error (%v) in query \"%v\" (error at %v)\n", lex.error.Error(), querystring, lex.lasttoken)
	}
	log.Debug("query %v", lex.query)
	switch lex.query.qtype {
	case SELECT_TYPE:
		target := lex.query.ContentsBson()
		if lex.query.distinct {
			if len(target) != 1 {
				return res, fmt.Errorf("Distinct query can only use one tag\n")
			}
			res, err = a.store.GetTags(target, true, lex.query.Contents[0], lex.query.WhereBson())
		} else {
			res, err = a.store.GetTags(target, false, "", lex.query.WhereBson())
		}

		if err != nil {
			return res, err
		}
	case DELETE_TYPE:
		var (
			err error
		)
		if len(lex.query.Contents) > 0 { // RemoveTags
			res, err = a.store.RemoveTags(lex.query.ContentsBson(), apikey, lex.query.WhereBson())
		} else { // RemoveDocs
			res, err = a.store.RemoveDocs(apikey, lex.query.WhereBson())
		}
		a.republisher.MetadataChangeKeys(lex.keys)
		log.Info("results %v", res)
		if err != nil {
			return res, err
		}
	case SET_TYPE:
		res, err = a.store.UpdateTags(lex.query.SetBson(), apikey, lex.query.WhereBson())
		if err != nil {
			return res, err
		}
		a.republisher.MetadataChangeKeys(lex.keys)
	case DATA_TYPE:
		// grab reference to the data query
		dq := lex.query.data

		// fetch all possible UUIDs that match the query
		uuids, err := a.GetUUIDs(lex.query.WhereBson())
		if err != nil {
			return res, err
		}

		// limit number of streams
		if dq.limit.streamlimit > 0 && len(uuids) > 0 {
			uuids = uuids[:dq.limit.streamlimit]
		}

		start := uint64(dq.start.UnixNano())
		end := uint64(dq.end.UnixNano())
		switch dq.dtype {
		case IN_TYPE:
			log.Debug("Data in start %v end %v", start, end)
			if start < end {
				res, err = a.GetData(uuids, start, end, UOT_NS, dq.timeconv)
			} else {
				res, err = a.GetData(uuids, end, start, UOT_NS, dq.timeconv)
			}
		case BEFORE_TYPE:
			log.Debug("Data before time %v", start)
			res, err = a.PrevData(uuids, start, int32(dq.limit.limit), UOT_NS, dq.timeconv)
		case AFTER_TYPE:
			log.Debug("Data after time %v", start)
			res, err = a.NextData(uuids, start, int32(dq.limit.limit), UOT_NS, dq.timeconv)
		}
		log.Debug("response %v uuids %v", res, uuids)
	}
	return res, nil
}

func (a *Archiver) Query2(querystring string, apikey string, w io.Writer) error {
	log.Info(querystring)
	lex := a.qp.Parse(querystring)
	if lex.error != nil {
		return fmt.Errorf("Error (%v) in query \"%v\" (error at %v)\n", lex.error.Error(), querystring, lex.lasttoken)
	}
	log.Debug("query %v", lex.query)
	// create root node from WHERE clause of tree
	wn := NewWhereNode(lex.query.WhereBson(), a.store)
	t := tree.NewTree(wn)
	// evalutes where clause
	// TODO: should this go into the tree.Run?
	t.Root.Input()
	// add the selector node to the tree
	sn := NewSelectDataNode(a, lex.query.data)
	t.AddChild(wn, sn)
	// run through the operators and build up the tree
	var (
		last    tree.Node = sn
		newNode tree.Node
	)
	for _, op := range lex.query.operators {
		newNode = a.qp.GetNodeFromOp(op, lex.query)
		if !a.qp.CheckOutToIn(last, newNode) {
			return fmt.Errorf("Node types do not match!")
		}
		t.AddChild(last, newNode)
		last = newNode
	}
	echoClient := NewEchoNode(w)
	t.AddChild(last, echoClient)
	return t.Run()
}

// A major problem is not knowing data types as they flow through. We really need a more efficient transport
// mechanism through giles, without this reliance on marshalling json. It would be nice to just copy data
// in and have the interface on the other side figure out what it is and then marshal/unmarshal appropriately.
// For any of the core archiver methods that right now return []SmapReading, these should instead return interface{},
// and then the interfaces on the other side must do a type switch to determine if it is []SmapReadingObject, []SmapReadingNumber
// or []SmapItem. These types have to be touched up, but because we are no longer directly serializing them,
// the structs can be constructed in a much more straightforward manner, with more type information!
//
// Fixing this should start from the databases up. The timeseries databases should return SmapNumbersResponse objects,
// and the object database should return SmapObjectResponse objects. Those changes will propagate up to the archiver API,
// which should return an interface{} rather than []SmapReading. Lastly, the front interfaces should use a type switch
// to determine what they're dealing with

// For each of the streamids, fetches all data between start and end (where
// start < end). The units for start/end are given by query_uot. We give the units
// so that each time series database can convert the incoming timestamps to whatever
// it needs (most of these will query the metadata store for the unit of time for the
// data stream it is accessing)
func (a *Archiver) GetData(streamids []string, start, end uint64, query_uot, to_uot UnitOfTime) (interface{}, error) {
	var err error
	ret := make([]interface{}, len(streamids))
	for idx, streamid := range streamids {
		stream_uot := a.store.GetUnitOfTime(streamid)
		if a.store.GetStreamType(streamid) == NUMERIC_STREAM {
			res, _ := a.tsdb.GetData(streamids[idx:idx+1], start, end, query_uot)
			ret[idx] = res[0]
			for _, reading := range ret[idx].(SmapNumbersResponse).Readings {
				reading.Time = convertTime(reading.Time, stream_uot, to_uot)
			}
		} else {
			ret[idx], _ = a.objstore.GetObjects(streamid, start, end, query_uot)
			for _, reading := range ret[idx].(SmapObjectResponse).Readings {
				reading.Time = convertTime(reading.Time, stream_uot, to_uot)
			}
		}
	}
	return ret, err
}

// For each of the streamids, fetches data before the start time. If limit is < 0, fetches all data.
// If limit >= 0, fetches only that number of points. See Archiver.GetData for explanation of query_uot
func (a *Archiver) PrevData(streamids []string, start uint64, limit int32, query_uot, to_uot UnitOfTime) (interface{}, error) {
	var err error
	ret := make([]interface{}, len(streamids))
	for idx, streamid := range streamids {
		stream_uot := a.store.GetUnitOfTime(streamid)
		if a.store.GetStreamType(streamid) == NUMERIC_STREAM {
			res, _ := a.tsdb.Prev(streamids[idx:idx+1], start, limit, query_uot)
			ret[idx] = res[0]
			for _, reading := range ret[idx].(SmapNumbersResponse).Readings {
				reading.Time = convertTime(reading.Time, stream_uot, to_uot)
			}
		} else {
			ret[idx], _ = a.objstore.PrevObject(streamid, start, query_uot)
			for _, reading := range ret[idx].(SmapObjectResponse).Readings {
				reading.Time = convertTime(reading.Time, stream_uot, to_uot)
			}
		}
	}
	return ret, err
}

// How do we handle getting data from 2 databases where we don't know which database each UUID is in?
// We can retrieve the stream type for each uuid, and use that to separate the uuids into two lists: timeseries streams
// and object streams. It would be nice to zip these back in order. Unsure if we should dispatch object
// and timeseries database transactions in parallel or serially (will this make a huge difference?), but might
// as well do serially for now.

// For each of the streamids, fetches data after the start time. If limit is < 0, fetches all data.
// If limit >= 0, fetches only that number of points. See Archiver.GetData for explanation of query_uot
func (a *Archiver) NextData(streamids []string, start uint64, limit int32, query_uot, to_uot UnitOfTime) (interface{}, error) {
	var err error
	ret := make([]interface{}, len(streamids))
	for idx, streamid := range streamids {
		stream_uot := a.store.GetUnitOfTime(streamid)
		if a.store.GetStreamType(streamid) == NUMERIC_STREAM {
			res, _ := a.tsdb.Next(streamids[idx:idx+1], start, limit, query_uot)
			ret[idx] = res[0]
			for _, reading := range ret[idx].(SmapNumbersResponse).Readings {
				reading.Time = convertTime(reading.Time, stream_uot, to_uot)
			}
		} else {
			ret[idx], _ = a.objstore.NextObject(streamid, start, query_uot)
			for _, reading := range ret[idx].(SmapObjectResponse).Readings {
				reading.Time = convertTime(reading.Time, stream_uot, to_uot)
			}
		}
	}
	return ret, err
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

func (a *Archiver) HandleUUIDSubscriber(s Subscriber, uuids []string, apikey string) {
	a.republisher.HandleUUIDSubscriber(s, uuids, apikey)
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
