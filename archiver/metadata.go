package archiver

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net"
	"sync"
	"sync/atomic"
)

type rdbStreamId struct {
	StreamId uint32
	UUID     string
}

type Store struct {
	session      *mgo.Session
	db           *mgo.Database
	streams      *mgo.Collection
	metadata     *mgo.Collection
	pathmetadata *mgo.Collection
	apikeys      *mgo.Collection
	apikeylock   sync.Mutex
	maxsid       *uint32
	streamlock   sync.Mutex
	uuidcache    *LRU
	pmdcache     *LRU
	pathcache    *LRU
	apikcache    *LRU
}

func NewStore(address *net.TCPAddr) *Store {
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
		Background: true,
		Sparse:     true,
	}
	err = metadata.EnsureIndex(index)
	if err != nil {
		log.Fatal("Could not create index on metadata.uuid")
	}

	err = streams.EnsureIndex(index)
	if err != nil {
		log.Fatal("Could not create index on streams.uuid")
	}

	index.Key = []string{"Path", "uuid"}
	err = pathmetadata.EnsureIndex(index)
	if err != nil {
		log.Fatal("Could not create index on pathmetadata.Path")
	}

	maxstreamid := &rdbStreamId{}
	streams.Find(bson.M{}).Sort("-streamid").One(&maxstreamid)
	var maxsid uint32 = 1
	if maxstreamid != nil {
		maxsid = maxstreamid.StreamId + 1
	}
	return &Store{session: session, db: db, streams: streams, metadata: metadata, pathmetadata: pathmetadata, apikeys: apikeys, maxsid: &maxsid, uuidcache: NewLRU(1000), pmdcache: NewLRU(1000), pathcache: NewLRU(1000), apikcache: NewLRU(1000)}
}

func (s *Store) getStreamId(uuid string) uint32 {
	s.streamlock.Lock()
	defer s.streamlock.Unlock()
	if v, found := s.uuidcache.Get(uuid); found {
		return v.(uint32)
	}
	streamid := &rdbStreamId{}
	err := s.streams.Find(bson.M{"uuid": uuid}).One(&streamid)
	if err != nil {
		// not found, so create
		streamid.StreamId = (*s.maxsid)
		streamid.UUID = uuid
		inserterr := s.streams.Insert(streamid)
		if inserterr != nil {
			log.Error("Error inserting streamid", inserterr)
			return 0
		}
		atomic.AddUint32(s.maxsid, 1)
		log.Notice("Creating StreamId %v for uuid %v", streamid.StreamId, uuid)
	}
	s.uuidcache.Set(uuid, streamid.StreamId)
	return streamid.StreamId
}

func (s *Store) apikeyexists(apikey string) (bool, error) {
	query := s.apikeys.Find(bson.M{"key": apikey})
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

func (s *Store) CanWrite(apikey, uuid string) (bool, error) {
	var record bson.M
	foundkey, found := s.apikcache.Get(uuid)
	if found && foundkey == apikey {
		return true, nil
	} else if found && foundkey != apikey {
		return false, errors.New("API key " + apikey + " is invalid for UUID " + uuid)
	}
	q := s.metadata.Find(bson.M{"uuid": uuid})
	count, _ := q.Count()
	s.apikeylock.Lock()
	defer s.apikeylock.Unlock()
	if count > 0 {
		q.One(&record)
		if record["_api"] != apikey {
			return false, errors.New("API key " + apikey + " is invalid for UUID " + uuid)
		}
		s.apikcache.Set(uuid, apikey)
	} else {
		// lock?
		exists, err := s.apikeyexists(apikey)
		if !exists || err != nil {
			return false, err
		}
		if uuid == "" {
			return false, err
		}
		//log.Debug("inserting uuid %v with api %v", uuid, apikey)
		err = s.metadata.Insert(bson.M{"uuid": uuid, "_api": apikey})
		if err != nil {
			return false, err
		}
		s.apikcache.Set(uuid, apikey)
	}
	return true, nil
}

func (s *Store) CheckKey(apikey string, messages map[string]*SmapMessage) (bool, error) {
	for _, msg := range messages {
		if msg.UUID == "" {
			continue
		} // no API key for path metadata
		ok, err := s.CanWrite(apikey, msg.UUID)
		if !ok || err != nil {
			return false, err
		}
	}
	return true, nil
}

/**
 * We use a pointer to the map so that we can edit it in-place
**/
func (s *Store) SavePathMetadata(messages *map[string]*SmapMessage) {
	/**
	 * We add the root metadata to everything in Contents
	**/
	var rootuuid string
	if (*messages)["/"] != nil && (*messages)["/"].Metadata != nil {
		rootuuid = (*messages)["/"].UUID
		for _, path := range (*messages)["/"].Contents {
			(*messages)["/"].Metadata["uuid"] = rootuuid
			_, err := s.pathmetadata.Upsert(bson.M{"Path": "/" + path, "uuid": rootuuid}, bson.M{"$set": (*messages)["/"].Metadata})
			if err != nil {
				log.Error("Error saving metadata for %v", "/"+path)
				log.Panic(err)
			}
		}
		delete((*messages), "/")
		s.pmdcache.Set(rootuuid, true)
	}
	/**
	 * For the rest of the keys, check if Contents is nonempty. If it is, we iterate through and update
	 * the metadata for that path
	**/
	for path, msg := range *messages {
		if msg.Metadata == nil {
			continue
		}
		if len(msg.Contents) > 0 {
			msg.Metadata["uuid"] = rootuuid
			_, err := s.pathmetadata.Upsert(bson.M{"Path": path, "uuid": rootuuid}, bson.M{"$set": msg.Metadata})
			if err != nil {
				log.Error("Error saving metadata for %v", path)
				log.Panic(err)
			}
			delete((*messages), path)
			s.pmdcache.Set(rootuuid, true)
		}
	}

}

func (s *Store) SaveMetadata(msg *SmapMessage) {
	/* check if we have any metadata or properties.
	   This should get hit once per stream unless the stream's
	   metadata changes
	*/
	var toWrite, prefixMetadata bson.M
	var uuidM = bson.M{"uuid": msg.UUID}
	changed, found := s.pmdcache.Get(msg.UUID)
	if !found {
		s.pmdcache.Set(msg.UUID, true)
	}
	if (changed != nil && changed.(bool)) || !found {
		toWrite = bson.M{}
		for _, prefix := range getPrefixes(msg.Path) {
			s.pathmetadata.Find(bson.M{"Path": prefix, "uuid": msg.UUID}).Select(bson.M{"_id": 0, "Path": 0, "_api": 0}).One(&prefixMetadata)
			for k, v := range prefixMetadata {
				toWrite["Metadata."+k] = v
			}
			s.pmdcache.Set(msg.UUID, false)
		}
		if len(toWrite) > 0 {
			_, err := s.metadata.Upsert(bson.M{"uuid": msg.UUID}, bson.M{"$set": toWrite})
			if err != nil {
				log.Critical("Error saving metadata for %v: %v", msg.UUID, err)
			}
		}
	}
	path, found := s.pathcache.Get(msg.UUID)
	// if not found,
	if !found {
		s.pathcache.Set(msg.UUID, msg.Path)
	}
	if (path != nil && path.(string) != msg.Path) || !found {
		_, err := s.metadata.Upsert(uuidM, bson.M{"$set": bson.M{"Path": msg.Path}})
		if err != nil {
			log.Critical("Error saving path for %v: %v", msg.UUID, err)
		}
	}
	if msg.Metadata == nil && msg.Properties == nil && msg.Actuator == nil {
		return
	}
	// check if we already have this path for this uuid
	if msg.Metadata != nil {
		for k, v := range msg.Metadata {
			_, err := s.metadata.Upsert(uuidM, bson.M{"$set": bson.M{"Metadata." + k: v}})
			if err != nil {
				log.Critical("Error saving metadata for %v: %v", msg.UUID, err)
			}
		}
	}
	if msg.Properties != nil {
		for k, v := range msg.Properties {
			_, err := s.metadata.Upsert(uuidM, bson.M{"$set": bson.M{"Properties." + k: v}})
			if err != nil {
				log.Critical("Error saving properties for %v: %v", msg.UUID, err)
			}
		}
	}
	if msg.Actuator != nil {
		_, err := s.metadata.Upsert(uuidM, bson.M{"$set": bson.M{"Actuator": msg.Actuator}})
		if err != nil {
			log.Critical("Error saving actuator for %v: %v", msg.UUID, err)
		}
	}
}

// Retrieves the tags indicated by `target` for documents that match the `where` clause. If `is_distinct` is true,
// then it will return a list of distinct values for the tag `distinct_key`
func (s *Store) GetTags(target bson.M, is_distinct bool, distinct_key string, where bson.M) ([]interface{}, error) {
	var res []interface{}
	var err error
	var staged *mgo.Query
	if len(target) == 0 {
		staged = s.metadata.Find(where).Select(bson.M{"_id": 0, "_api": 0})
	} else {
		target["_id"] = 0
		target["_api"] = 0
		staged = s.metadata.Find(where).Select(target)
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

func (s *Store) SetTags(updates bson.M, apikey string, where bson.M) (bson.M, error) {
	var res bson.M
	uuids, err := s.GetUUIDs(where)
	if err != nil {
		return res, err
	}
	for _, uuid := range uuids {
		ok, err := s.CanWrite(apikey, uuid)
		if !ok || err != nil {
			return res, err
		}
	}
	info, err2 := s.metadata.UpdateAll(where, bson.M{"$set": updates})
	if err2 != nil {
		return res, err2
	}
	log.Info("Updated %v records", info.Updated)
	return bson.M{"Updated": info.Updated}, nil
}

// Return all metadata for a certain UUID
func (s *Store) TagsUUID(uuid string) ([]bson.M, error) {
	staged := s.metadata.Find(bson.M{"uuid": uuid}).Select(bson.M{"_id": 0, "_api": 0})
	res := []bson.M{}
	err := staged.All(&res)
	return res, err
}

// Resolve a query to a slice of UUIDs
func (s *Store) GetUUIDs(where bson.M) ([]string, error) {
	var tmp []bson.M
	var res = []string{}
	err := s.metadata.Find(where).Select(bson.M{"uuid": 1}).All(&tmp)
	if err != nil {
		return res, err
	}
	for _, uuid := range tmp {
		res = append(res, uuid["uuid"].(string))
	}
	return res, nil
}

// Resolve a query to a slice of StreamIds
func (s *Store) getStreamIds(where bson.M) []uint32 {
	var tmp []bson.M
	var res []uint32
	err := s.metadata.Find(where).Select(bson.M{"uuid": 1}).All(&tmp)
	if err != nil {
		log.Panic(err)
	}
	for _, uuid := range tmp {
		res = append(res, s.getStreamId(uuid["uuid"].(string)))
	}
	return res
}

// retrieve the unit of time for the stream identified by the given UUID.
// Should return one of ns, us, ms, s; defaults to ms
func (s *Store) GetUnitofTime(uuid string) string {
	var res bson.M
	err := s.metadata.Find(bson.M{"uuid": uuid}).Select(bson.M{"Properties.UnitofTime": 1}).One(&res)
	if err != nil {
		return "ms"
	}
	return res["Properties"].(bson.M)["UnitofTime"].(string)
}
