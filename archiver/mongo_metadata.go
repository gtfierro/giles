package archiver

import (
	"encoding/base64"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

// default select clause to ignore internal variables
var ignoreDefault = bson.M{"_id": 0, "_api": 0}

//TODO: copy the session for a transaction -- this is faster
type MongoStore struct {
	session        *mgo.Session
	db             *mgo.Database
	streams        *mgo.Collection
	metadata       *mgo.Collection
	pathmetadata   *mgo.Collection
	apikeys        *mgo.Collection
	apikeylock     sync.Mutex
	maxsid         *uint32
	streamlock     sync.Mutex
	streamtypelock sync.Mutex
	uuidcache      *Cache
	apikcache      *Cache
	uotcache       *Cache
	streamtype     *Cache
	enforceKeys    bool
}

type rdbStreamId struct {
	StreamId uint32
	UUID     string
}

func NewMongoStore(address *net.TCPAddr) *MongoStore {
	log.Notice("Connecting to MongoDB at %v...", address.String())
	session, err := mgo.Dial(address.String())
	if err != nil {
		log.Critical("Could not connect to MongoDB: %v", err)
		return nil
	}
	log.Notice("...connected!")
	//session.SetMode(mgo.Eventual, true)
	db := session.DB("archiver")
	streams := db.C("streams")
	metadata := db.C("metadata")
	pathmetadata := db.C("pathmetadata")
	apikeys := db.C("apikeys")
	// create indexes
	index := mgo.Index{
		Key:        []string{"uuid"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	err = metadata.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Could not create index on metadata.uuid (%v)", err)
	}

	index.Key = []string{"Properties.UnitofTime"}
	index.Unique = false
	err = metadata.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Could not create index on metadata.properties.unitofmeasure (%v)", err)
	}

	index.Key = []string{"Properties.UnitofMeasure"}
	err = metadata.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Could not create index on metadata.properties.unitofmeasure (%v)", err)
	}

	index.Key = []string{"Properties.StreamType"}
	err = metadata.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Could not create index on metadata.properties.streamtype (%v)", err)
	}

	index.Unique = true
	index.Key = []string{"Path"}
	err = pathmetadata.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Could not create index on pathmetadata.Path (%v)", err)
	}

	index.Key = []string{"uuid"}
	err = streams.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Could not create index on streams.uuid (%v)", err)
	}

	index.Key = []string{"name", "email"}
	err = apikeys.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Could not create index on apikeys (%v)", err)
	}

	maxstreamid := &rdbStreamId{}
	streams.Find(bson.M{}).Sort("-streamid").One(&maxstreamid)
	var maxsid uint32 = 1
	if maxstreamid != nil {
		maxsid = maxstreamid.StreamId + 1
	}
	return &MongoStore{session: session,
		db:           db,
		streams:      streams,
		metadata:     metadata,
		pathmetadata: pathmetadata,
		apikeys:      apikeys,
		maxsid:       &maxsid,
		uuidcache:    NewCache(1000),
		apikcache:    NewCache(1000),
		uotcache:     NewCache(1000),
		streamtype:   NewCache(1000),
		enforceKeys:  true}
}

/* MetadataStore interface implementation*/

func (ms *MongoStore) EnforceKeys(enforce bool) {
	ms.enforceKeys = enforce
}

func (ms *MongoStore) CheckKey(apikey string, messages map[string]*SmapMessage) (bool, error) {
	for _, msg := range messages {
		if msg.UUID == "" {
			continue
		} // no API key for path metadata
		ok, err := ms.CanWrite(apikey, msg.UUID)
		if !ok || err != nil {
			return false, err
		}
	}
	return true, nil
}

func (ms *MongoStore) CanWrite(apikey, uuid string) (bool, error) {
	var record bson.M
	if !ms.enforceKeys {
		return true, nil
	}
	foundkey, found := ms.apikcache.Get(uuid)
	if found && foundkey == apikey {
		return true, nil
	} else if found && foundkey != apikey {
		return false, errors.New("API key " + apikey + " is invalid for UUID " + uuid)
	}
	q := ms.metadata.Find(bson.M{"uuid": uuid})
	count, _ := q.Count()
	ms.apikeylock.Lock()
	defer ms.apikeylock.Unlock()
	if count > 0 {
		q.One(&record)
		if record["_api"] != apikey {
			return false, errors.New("API key " + apikey + " is invalid for UUID " + uuid)
		}
		ms.apikcache.Set(uuid, apikey)
	} else {
		// lock?
		exists, err := ms.ApiKeyExists(apikey)
		if !exists || err != nil {
			return false, err
		}
		if uuid == "" {
			return false, err
		}
		err = ms.metadata.Insert(bson.M{"uuid": uuid, "_api": apikey})
		if err != nil {
			return false, err
		}
		ms.apikcache.Set(uuid, apikey)
	}
	return true, nil
}

// SavePathMetadata takes a map of paths to sMAP messages and saves them to the
// path metadata store.  The Path keys can be terminal paths, e.g. the full
// path of a timeseries, which are identified by the linked sMAP message having
// a UUID (and optionally a Readings) field. Path keys can also be non-terminal
// paths, which do not have an associated UUID.
//
// This method stores the metadata associated with paths, indexed by each path.
// The pathmetadata collection should be queried internally by SaveTimeseriesMetadata
// to build up the full document for each individual timeseries
func (ms *MongoStore) SavePathMetadata(messages map[string]*SmapMessage) error {
	var (
		toWrite  bson.M
		setBson  = bson.M{"$set": ""}
		pathBson = bson.M{"Path": ""}
		retErr   error
	)
	for path, msg := range messages {
		// check if we have anything to do
		if msg.Path == "" || (msg.Metadata == nil && msg.Properties == nil && msg.Actuator == nil) {
			continue
		}

		// construct base object
		toWrite = bson.M{"Path": path}

		// populate the toWrite object with the metadata
		if msg.Metadata != nil && len(msg.Metadata) > 0 {
			for k, v := range msg.Metadata {
				toWrite["Metadata."+k] = v
			}
		}
		if msg.Properties != nil && len(msg.Properties) > 0 {
			for k, v := range msg.Properties {
				toWrite["Properties."+k] = v
			}
		}
		if msg.Actuator != nil && len(msg.Actuator) > 0 {
			for k, v := range msg.Actuator {
				toWrite["Actuator."+k] = v
			}
		}

		pathBson["Path"] = path
		setBson["$set"] = toWrite
		_, retErr = ms.pathmetadata.Upsert(pathBson, setBson)

	}
	return retErr
}

// Here, we are handed a chunk of incoming sMAP messages, which can include both non-terminal
// and terminal (timeseries) paths. Timeseries (terminal) paths are identified by having a UUID
// key in their SmapMessage struct. For each of these, we decompose the full timeseries Path
// into its components -- e.g. /a/b/c -> /, /a, /a/b -- and inherit from the PathMetadata
// collection into the metadata for this source. Timeseries-specific metadata is then upserted
// into this document, and the result is saved in the Metadata collection
// TODO: this is super slow: Find and Upsert take the most time. Avoid this method if we already
// have all the new metadata!
// Important to note: sMAP 2.0 archiver does *not* do path inheritance of metadata within the archiver.
// Inheritance only happens w/n a source, when multiple messages are sent
func (ms *MongoStore) SaveTimeseriesMetadata(messages map[string]*SmapMessage) error {
	var (
		one      bson.M
		queryErr error
		retErr   error
		ignore   = bson.M{"_id": 0, "_api": 0, "Path": 0}
		setBson  = bson.M{"$set": ""}
		uuidBson = bson.M{"uuid": ""}
	)
	session := ms.session.Copy()
	pm := session.DB("archiver").C("pathmetadata")
	md := session.DB("archiver").C("metadata")
	defer session.Close()
	querypath := bson.M{"Path": ""}
	for path, msg := range messages {
		if msg.UUID == "" { // not a timeseries path
			continue
		}
		// here, we know we have a timeseries path, so we initialize the eventual document
		toWrite := bson.M{"Path": path, "uuid": msg.UUID}

		// for each of its prefixes, fetch the document from the PathMetadata collection
		// and merge it into the toWrite doc
		for _, prefix := range getPrefixes(path) {
			if queryErr == nil {
				querypath["Path"] = prefix

				// ignore "not found" path errors
				queryErr = pm.Find(querypath).Select(ignore).One(&one)
				if queryErr != nil && queryErr.Error() == "not found" {
					queryErr = nil
				}

				// do the merge
				for k, v := range one {
					toWrite[k] = v
				}
			}
		}

		// finally, merge in the timeseries-specific metadata
		if msg.Metadata != nil && len(msg.Metadata) > 0 {
			for k, v := range msg.Metadata {
				toWrite["Metadata."+k] = v
			}
		}
		if msg.Properties != nil && len(msg.Properties) > 0 {
			for k, v := range msg.Properties {
				toWrite["Properties."+k] = v
			}
		}
		if msg.Actuator != nil && len(msg.Actuator) > 0 {
			for k, v := range msg.Actuator {
				toWrite["Actuator."+k] = v
			}
		}

		setBson["$set"] = toWrite
		uuidBson["uuid"] = msg.UUID
		if len(toWrite) > 2 {
			_, retErr = md.Upsert(uuidBson, setBson)
			if retErr != nil {
				log.Critical("error saving md %v", retErr)
				return retErr
			}
		}
	}
	return queryErr
}

/*
The incoming messages will be in the form of {pathname: metadata/properties/etc}.
Only the timeseries will have UUIDs attached. When we receive a message like this, we need
to compress all of the prefix-path kv pairs into each of the timeseries, and then save those
timeseries to the metadata collection
*/
//TODO: when we perform the inheritance, we should return the normalized map to the archiver,
//      either by copying it back or by creating a new one.
func (ms *MongoStore) SaveTags(messages *map[string]*SmapMessage) error {
	var err error
	for path, msg := range *messages {
		if msg.UUID == "" || (msg.Metadata == nil && msg.Properties == nil && msg.Actuator == nil) {
			delete(*messages, path)
			continue
		}
		toWrite := bson.M{"Path": path, "uuid": msg.UUID}
		if msg.Metadata != nil && len(msg.Metadata) > 0 {
			for k, v := range msg.Metadata {
				toWrite["Metadata."+k] = v
			}
		}
		if msg.Properties != nil && len(msg.Properties) > 0 {
			for k, v := range msg.Properties {
				toWrite["Properties."+k] = v
			}
		}
		if msg.Actuator != nil && len(msg.Actuator) > 0 {
			for k, v := range msg.Actuator {
				toWrite["Actuator."+k] = v
			}
		}
		for _, prefix := range getPrefixes(path) { // accumulate all metadata for this timeseries
			if (*messages)[prefix] == nil {
				continue
			}
			if (*messages)[prefix].Metadata != nil {
				for k, v := range (*messages)[prefix].Metadata {
					toWrite["Metadata."+k] = v
					if msg.Metadata == nil {
						msg.Metadata = bson.M{}
					}
					msg.Metadata[k] = v
				}
			}
			if (*messages)[prefix].Properties != nil {
				for k, v := range (*messages)[prefix].Properties {
					toWrite["Properties."+k] = v
					if msg.Properties == nil {
						msg.Properties = bson.M{}
					}
					msg.Properties[k] = v
				}
			}
		}
		if len(toWrite) > 0 {
			_, err = ms.metadata.Upsert(bson.M{"uuid": msg.UUID}, bson.M{"$set": toWrite})
		}
	}
	return err
}

// Retrieves the tags indicated by `target` for documents that match the `where` clause. If `is_distinct` is true,
// then it will return a list of distinct values for the tag `distinct_key`
func (ms *MongoStore) GetTags(target bson.M, is_distinct bool, distinct_key string, where bson.M) ([]interface{}, error) {
	var res []interface{}
	var err error
	var staged *mgo.Query
	uuids, uuidErr := ms.GetUUIDs(where)
	if uuidErr != nil {
		return res, uuidErr
	}
	staged = ms.metadata.Find(bson.M{"uuid": bson.M{"$in": uuids}}).Select(bson.M{"_id": 0, "_api": 0})
	err = staged.All(&res)

	if len(target) == 0 {
		staged = ms.metadata.Find(where).Select(bson.M{"_id": 0, "_api": 0})
	} else {
		// because we can't have both inclusion and exclusion, we check if the
		// target is including anything. If we get a "1", then we don't add
		// our exclusions here
		doExclude := true
		for _, include := range target {
			if include == 1 {
				doExclude = false
				break
			}
		}
		target["_id"] = 0 // always exclude this
		if doExclude {
			target["_api"] = 0
		}
		staged = ms.metadata.Find(where).Select(target)
	}
	if is_distinct {
		var res2 []interface{}
		err = staged.Distinct(distinct_key, &res2)
		return res2, err
	} else {
		err = staged.All(&res)
		return res, err
	}
}

func (ms *MongoStore) UpdateTags(updates bson.M, apikey string, where bson.M) (bson.M, error) {
	var res bson.M
	uuids, err := ms.GetUUIDs(where)
	if err != nil {
		return res, err
	}
	for _, uuid := range uuids {
		ok, err := ms.CanWrite(apikey, uuid)
		if !ok || err != nil {
			return res, err
		}
	}
	info, err2 := ms.metadata.UpdateAll(where, bson.M{"$set": updates})
	if err2 != nil {
		return res, err2
	}
	log.Info("Updated %v records", info.Updated)
	return bson.M{"Updated": info.Updated}, nil
}

func (ms *MongoStore) RemoveDocs(apikey string, where bson.M) (bson.M, error) {
	var res bson.M
	uuids, err := ms.GetUUIDs(where)
	if err != nil {
		return res, err
	}
	for _, uuid := range uuids {
		ok, canWriteErr := ms.CanWrite(apikey, uuid)
		if !ok || canWriteErr != nil {
			return res, canWriteErr
		}
	}
	ci, removeErr := ms.metadata.RemoveAll(where)
	log.Info("Removed %v records", ci.Removed)
	return bson.M{"Removed": ci.Removed}, removeErr
}

func (ms *MongoStore) RemoveTags(updates bson.M, apikey string, where bson.M) (bson.M, error) {
	var res bson.M
	uuids, err := ms.GetUUIDs(where)
	if err != nil {
		return res, err
	}
	for _, uuid := range uuids {
		ok, canWriteErr := ms.CanWrite(apikey, uuid)
		if !ok || canWriteErr != nil {
			return res, canWriteErr
		}
	}
	info, updateErr := ms.metadata.UpdateAll(where, bson.M{"$unset": updates})
	log.Info("Updated %v records", info.Updated)
	return bson.M{"Updated": info.Updated}, updateErr
}

// Return all metadata for a certain UUID
func (ms *MongoStore) UUIDTags(uuid string) (bson.M, error) {
	staged := ms.metadata.Find(bson.M{"uuid": uuid}).Select(bson.M{"_id": 0, "_api": 0})
	res := bson.M{}
	err := staged.One(&res)
	return res, err
}

// Resolve a query to a slice of UUIDs
func (ms *MongoStore) GetUUIDs(where bson.M) ([]string, error) {
	var tmp []bson.M
	var res = []string{}
	err := ms.metadata.Find(where).Select(bson.M{"uuid": 1}).All(&tmp)
	if err != nil {
		return res, err
	}
	for _, uuid := range tmp {
		res = append(res, uuid["uuid"].(string))
	}
	return res, nil
}

// retrieve the unit of time for the stream identified by the given UUID.
// Should return one of ns, us, ms, s; defaults to ms
func (ms *MongoStore) GetUnitOfTime(uuid string) UnitOfTime {
	var res bson.M
	if uot, found := ms.uotcache.Get(uuid); found {
		return uot.(UnitOfTime)
	}
	err := ms.metadata.Find(bson.M{"uuid": uuid}).Select(bson.M{"Properties.UnitofTime": 1}).One(&res)
	if err != nil {
		ms.uotcache.Set(uuid, UOT_MS)
		return UOT_MS
	}
	if prop, found := res["Properties"]; found {
		if uot, found := prop.(bson.M)["UnitofTime"]; found {
			switch uot.(string) {
			case "ns":
				ms.uotcache.Set(uuid, UOT_NS)
				return UOT_NS
			case "us":
				ms.uotcache.Set(uuid, UOT_US)
				return UOT_US
			case "ms":
				ms.uotcache.Set(uuid, UOT_MS)
				return UOT_MS
			case "s":
				ms.uotcache.Set(uuid, UOT_S)
				return UOT_S
			}
		}
	}
	ms.uotcache.Set(uuid, UOT_MS)
	return UOT_MS
}

func (ms *MongoStore) GetStreamType(uuid string) StreamType {
	ms.streamtypelock.Lock()
	if st, found := ms.streamtype.Get(uuid); found {
		ms.streamtypelock.Unlock()
		return st.(StreamType)
	}
	var res bson.M
	err := ms.metadata.Find(bson.M{"uuid": uuid}).Select(bson.M{"Properties.StreamType": 1}).One(&res)
	if err != nil {
		ms.streamtype.Set(uuid, NUMERIC_STREAM)
	}
	if prop, found := res["Properties"]; found {
		if st, found := prop.(bson.M)["StreamType"]; found {
			if st.(string) == "object" {
				ms.streamtype.Set(uuid, OBJECT_STREAM)
				ms.streamtypelock.Unlock()
				return OBJECT_STREAM
			} else {
				ms.streamtype.Set(uuid, NUMERIC_STREAM)
				ms.streamtypelock.Unlock()
				return NUMERIC_STREAM
			}
		}
	}
	ms.streamtype.Set(uuid, NUMERIC_STREAM)
	ms.streamtypelock.Unlock()
	return NUMERIC_STREAM
}

/** Implementing the APIKeyManager interface **/

func (ms *MongoStore) GetStreamId(uuid string) uint32 {
	ms.streamlock.Lock()
	defer ms.streamlock.Unlock()
	if v, found := ms.uuidcache.Get(uuid); found {
		return v.(uint32)
	}
	streamid := &rdbStreamId{}
	err := ms.streams.Find(bson.M{"uuid": uuid}).One(&streamid)
	if err != nil {
		// not found, so create
		streamid.StreamId = (*ms.maxsid)
		streamid.UUID = uuid
		inserterr := ms.streams.Insert(streamid)
		if inserterr != nil {
			log.Error("Error inserting streamid", inserterr)
			return 0
		}
		atomic.AddUint32(ms.maxsid, 1)
		log.Notice("Creating StreamId %v for uuid %v", streamid.StreamId, uuid)
	}
	ms.uuidcache.Set(uuid, streamid.StreamId)
	return streamid.StreamId
}

func (ms *MongoStore) ApiKeyExists(apikey string) (bool, error) {
	query := ms.apikeys.Find(bson.M{"key": apikey})
	count, err := query.Count()
	if err != nil {
		return false, err
	}
	if count > 1 {
		return false, errors.New("More than 1 API key with value " + apikey)
	}
	if count < 1 {
		return false, errors.New("No API key with value " + apikey)
	}
	return true, nil
}

func (ms *MongoStore) NewKey(name, email string, public bool) (string, error) {
	var apikey string
	var err error
	urandom, err := os.Open("/dev/urandom")
	if err != nil {
		return apikey, errors.New(fmt.Sprintf("Could not open /dev/urandom (%v)", err))
	}
	randbytes := make([]byte, 64)
	n, err := urandom.Read(randbytes)
	if err != nil {
		return apikey, errors.New(fmt.Sprintf("Could not read /dev/urandom (%v)", err))
	}
	if n != 64 {
		return apikey, errors.New(fmt.Sprintf("Could not read 64 bytes from /dev/urandom %v", n))
	}
	apikey = base64.URLEncoding.EncodeToString(randbytes)
	err = ms.apikeys.Insert(bson.M{"key": apikey, "name": name, "email": email, "public": public})
	if err != nil {
		return apikey, errors.New(fmt.Sprintf("could not save new apikey for %v %v to Mongo (%v)", name, email, err))
	}
	return apikey, nil
}

func (ms *MongoStore) GetKey(name, email string) (string, error) {
	var res interface{}
	err := ms.apikeys.Find(bson.M{"name": name, "email": email}).Select(bson.M{"key": 1}).One(&res)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not retrieve apikey for %v %v (%v)", name, email, err))
	}
	return res.(bson.M)["key"].(string), nil
}

func (ms *MongoStore) ListKeys(email string) ([]map[string]interface{}, error) {
	var res []map[string]interface{}
	err := ms.apikeys.Find(bson.M{"email": email}).All(&res)
	if err != nil {
		return res, errors.New(fmt.Sprintf("Could not get apikeys for %v (%v)", email, err))
	}
	return res, nil
}

func (ms *MongoStore) DeleteKeyByName(name, email string) (string, error) {
	ci, err := ms.apikeys.RemoveAll(bson.M{"name": name, "email": email})
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not delete keys for %v %v (%v)", name, email, err))
	}
	return fmt.Sprintf("Removed %v records", ci.Removed), nil
}

func (ms *MongoStore) DeleteKeyByValue(key string) (string, error) {
	err := ms.apikeys.Remove(bson.M{"key": key})
	return fmt.Sprintf("Removed key"), err
}

func (ms *MongoStore) Owner(key string) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := ms.apikeys.Find(bson.M{"key": key}).Select(bson.M{"name": 1, "email": 1}).One(&res)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not find owners for key %v (%v)", key, err))
	}
	return res, nil
}
