package archiver

import (
	"errors"
	"fmt"
	"github.com/gtfierro/msgpack"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net"
	"sync"
)

// The object store interface into Mongo uses a collection named 'objects'. Each document in this
// collection contains 3 keys:
//      uuid: the stream identifier
//      object: a binary blob (byte array) of data (MsgPack encoded)
//      timestamp: the timestamp associated with this record IN NANOSECONDS
// This is obviously a very primitive interface to the object store, and doesn't do nice things like
// transaction coalescence.
type MongoObjectStore struct {
	session     *mgo.Session
	db          *mgo.Database
	objects     *mgo.Collection
	enforceKeys bool
	store       MetadataStore
	bufpool     sync.Pool
}

func NewMongoObjectStore(address *net.TCPAddr) *MongoObjectStore {
	log.Notice("ObjectStore: Connecting to MongoDB at %v...", address.String())
	session, err := mgo.Dial(address.String())
	if err != nil {
		log.Critical("Could not connect to MongoDB: %v", err)
		return nil
	}
	log.Notice("...connected!")
	db := session.DB("archiver")
	objects := db.C("objects")
	// create indexes
	index := mgo.Index{
		Key:        []string{"uuid", "timestamp"},
		Unique:     true,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = objects.EnsureIndex(index)
	if err != nil {
		log.Fatal("Could not create index on objects")
	}
	bufpool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 1000)
		},
	}
	return &MongoObjectStore{session: session,
		db:          db,
		objects:     objects,
		enforceKeys: true,
		bufpool:     bufpool}
}

func (ms *MongoObjectStore) AddObject(msg *SmapMessage) (bool, error) {
	var time_ns uint64
	if msg.Readings == nil || len(msg.Readings) == 0 {
		return false, errors.New("No readings in sMAP message")
	} // return early
	if len(msg.UUID) == 0 {
		return false, errors.New("Reading has no UUID")
	}

	uot := ms.store.GetUnitOfTime(msg.UUID)

	bytes := ms.bufpool.Get().([]byte)
	defer ms.bufpool.Put(bytes)
	for _, rdg := range msg.Readings {
		length := msgpack.Encode(rdg[1], &bytes)
		if length == 1000 {
			return false, errors.New("Encoded value was larger than 1000 bytes!")
		}
		if time, ok := rdg[0].(uint64); ok {
			time_ns = convertTime(time, uot, UOT_NS)
		} else {
			return false, fmt.Errorf("Invalid timestamp %v", rdg[0])
		}
		err := ms.objects.Insert(bson.M{"uuid": msg.UUID, "timestamp": time_ns, "object": bytes[:length]})
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (ms *MongoObjectStore) PrevObject(uuid string, time uint64, uot UnitOfTime) (SmapReading, error) {
	var res bson.M
	var ret SmapReading
	time_ns := convertTime(time, uot, UOT_NS)
	err := ms.objects.Find(bson.M{"uuid": uuid, "timestamp": bson.M{"$lte": time_ns}}).Sort("-timestamp").One(&res)
	if err != nil {
		log.Error("got an err %v", err)
		return ret, err
	}
	bytes := res["object"].([]byte)
	_, decoded := msgpack.Decode(&bytes, 0)
	uot = ms.store.GetUnitOfTime(uuid)
	timestamp := convertTime(uint64(res["timestamp"].(int64)), UOT_NS, uot)
	ret = SmapReading{Readings: [][]interface{}{[]interface{}{float64(timestamp), decoded}}, UUID: res["uuid"].(string)}
	return ret, nil
}

func (ms *MongoObjectStore) NextObject(uuid string, time uint64, uot UnitOfTime) (SmapReading, error) {
	var res bson.M
	var ret SmapReading
	time_ns := convertTime(time, uot, UOT_NS)
	err := ms.objects.Find(bson.M{"uuid": uuid, "timestamp": bson.M{"$gte": time_ns}}).Sort("+timestamp").One(&res)
	if err != nil {
		log.Error("got an err %v", err)
		return ret, err
	}
	bytes := res["object"].([]byte)
	_, decoded := msgpack.Decode(&bytes, 0)
	uot = ms.store.GetUnitOfTime(uuid)
	timestamp := convertTime(uint64(res["timestamp"].(int64)), UOT_NS, uot)
	ret = SmapReading{Readings: [][]interface{}{[]interface{}{float64(timestamp), decoded}}, UUID: res["uuid"].(string)}
	return ret, nil
}

func (ms *MongoObjectStore) GetObjects(uuid string, start uint64, end uint64, uot UnitOfTime) (SmapReading, error) {
	var res []bson.M
	var ret SmapReading
	start_time_ns := convertTime(start, uot, UOT_NS)
	end_time_ns := convertTime(end, uot, UOT_NS)
	err := ms.objects.Find(bson.M{"uuid": uuid,
		"$and": []bson.M{bson.M{"timestamp": bson.M{"$gte": start_time_ns}}, bson.M{"timestamp": bson.M{"$lte": end_time_ns}}},
	}).Sort("+timestamp").All(&res)
	if err != nil {
		log.Error("got an err %v", err)
		return ret, err
	}
	log.Debug("res %v", res)
	uot = ms.store.GetUnitOfTime(uuid)
	ret = SmapReading{Readings: make([][]interface{}, len(res)), UUID: uuid}
	for idx, chunk := range res {
		bytes := chunk["object"].([]byte)
		_, decoded := msgpack.Decode(&bytes, 0)
		timestamp := convertTime(uint64(chunk["timestamp"].(int64)), UOT_NS, uot)
		ret.Readings[idx] = []interface{}{float64(timestamp), decoded}
	}
	return ret, nil
}

func (ms *MongoObjectStore) AddStore(store MetadataStore) {
	ms.store = store
}
