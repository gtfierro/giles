package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
	"sync/atomic"
)

type RDBStreamId struct {
	StreamId uint32
	UUID     string
}

type Store struct {
	session  *mgo.Session
	db       *mgo.Database
	streams  *mgo.Collection
	metadata *mgo.Collection
	maxsid   *uint32
}

func NewStore(ip string, port int) *Store {
	address := ip + ":" + strconv.Itoa(port)
	session, err := mgo.Dial(address)
	if err != nil {
		log.Panic(err)
		return nil
	}
	db := session.DB("archiver")
	streams := db.C("streams")
	metadata := db.C("metadata")
	maxstreamid := &RDBStreamId{}
	streams.Find(bson.M{}).Sort("-streamid").One(&maxstreamid)
	var maxsid uint32 = 1
	if maxstreamid != nil {
		maxsid = maxstreamid.StreamId + 1
	}
	return &Store{session: session, db: db, streams: streams, metadata: metadata, maxsid: &maxsid}
}

func (s *Store) GetStreamId(uuid string) uint32 {
	for k, v := range UUIDCache {
		if k == uuid {
			return v
		}
	}
	streamid := &RDBStreamId{}
	err := s.streams.Find(bson.M{"uuid": uuid}).One(&streamid)
	if err != nil {
		streamlock.Lock()
		// not found, so create
		streamid.StreamId = (*s.maxsid)
		streamid.UUID = uuid
		inserterr := s.streams.Insert(streamid)
		if inserterr != nil {
			log.Println(inserterr)
			return 0
		}
		atomic.AddUint32(s.maxsid, 1)
		log.Println("Creating StreamId", streamid.StreamId, "for uuid", uuid)
		streamlock.Unlock()
	}
	UUIDCache[uuid] = streamid.StreamId
	return streamid.StreamId
}

func (s *Store) SaveMetadata(msg *SmapMessage) {
	/* check if we have any metadata or properties.
	   This should get hit once per stream unless the stream's
	   metadata changes
	*/
	if msg.Metadata == nil && msg.Properties == nil {
		return
	}

	if msg.Metadata != nil {
		_, err := s.metadata.Upsert(bson.M{"uuid": msg.UUID}, bson.M{"$set": bson.M{"Metadata": msg.Metadata}})
		if err != nil {
			log.Println("Error saving metadata for", msg.UUID)
			log.Panic(err)
		}
	}
	if msg.Properties != nil {
		_, err := s.metadata.Upsert(bson.M{"uuid": msg.UUID}, bson.M{"$set": bson.M{"Properties": msg.Properties}})
		if err != nil {
			log.Println("Error saving properties for", msg.UUID)
			log.Panic(err)
		}
	}
}

//TODO: use the rest of the AST! this only does tags queries right now
func (s *Store) Query(stringquery []byte) *[]bson.M {
	var res []bson.M
	ast := parse(string(stringquery))
	where := ast.Where.ToBson()
	switch ast.TargetType {
	case TAGS_TARGET:
		var err error
		var staged *mgo.Query
		target := ast.Target.(*TagsTarget).ToBson()
		if len(target) == 0 {
			staged = s.metadata.Find(where).Select(bson.M{"_id": 0})
		} else {
			target["_id"] = 0
			staged = s.metadata.Find(where).Select(target)
		}
		if ast.Target.(*TagsTarget).Distinct {
			log.Panic("Distinct not currently working")
			err = staged.Distinct(ast.Target.(*TagsTarget).Contents[0], &res)
		} else {
			err = staged.All(&res)
		}
		if err != nil {
			log.Panic(err)
		}
	case DATA_TARGET:
		log.Println("Data operations not supported yet")
		return &res
	}
	return &res
}

/*
  Resolve a query to a slice of UUIDs
*/
func (s *Store) GetUUIDs(where bson.M) []string {
	var res []string
	err := s.metadata.Find(where).Select(bson.M{"uuid": 1}).All(&res)
	if err != nil {
		log.Panic(err)
	}
	return res
}

/*
  Resolve a query to a slice of StreamIds
*/
func (s *Store) GetStreamIds(where bson.M) []uint32 {
	var tmp []bson.M
	var res []uint32
	err := s.metadata.Find(where).Select(bson.M{"uuid": 1}).All(&tmp)
	if err != nil {
		log.Panic(err)
	}
	for _, uuid := range tmp {
		res = append(res, s.GetStreamId(uuid["uuid"].(string)))
	}
	return res
}
