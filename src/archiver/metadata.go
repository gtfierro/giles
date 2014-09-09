package main

import (
	"encoding/json"
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

func (s *Store) Query(stringquery []byte) ([]byte, error) {
	var res []bson.M
	var d []byte
	var err error
	ast := parse(string(stringquery))
	where := ast.Where.ToBson()
	switch ast.TargetType {
	case TAGS_TARGET:
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
		d, err = json.Marshal(res)
	case SET_TARGET:
		target := ast.Target.(*SetTarget).Updates
		info, err2 := s.metadata.UpdateAll(where, bson.M{"$set": target})
		if err2 != nil {
			return d, err2
		}
		log.Println("Updated", info.Updated, "records")
		d, err = json.Marshal(bson.M{"Updated": info.Updated})
	case DATA_TARGET:
		target := ast.Target.(*DataTarget)
		uuids := store.GetUUIDs(ast.Where.ToBson())
        log.Println("these uuids",uuids)
		if target.Streamlimit > -1 {
			uuids = uuids[:target.Streamlimit] // limit number of streams
		}
		conn, err := rdb.GetConnection()
		if err != nil {
			return d, err
		}
		var response []SmapResponse
		switch target.Type {
		case IN:
			start := uint64(target.Start.Unix())
			end := uint64(target.End.Unix())
			log.Println("start", start, "end", end)
			response, err = rdb.GetData(uuids, start, end, &conn)
		case AFTER:
			ref := uint64(target.Ref.Unix())
			log.Println("after", ref)
			response, err = rdb.Next(uuids, ref, target.Limit, &conn)
		case BEFORE:
			ref := uint64(target.Ref.Unix())
			log.Println("before", ref)
			response, err = rdb.Prev(uuids, ref, target.Limit, &conn)
		}
		if err != nil {
			return d, err
		}
		d, err = json.Marshal(response)
	}
	if err != nil {
		return d, err
	}
	return d, err
}

/*
  Resolve a query to a slice of UUIDs
*/
func (s *Store) GetUUIDs(where bson.M) []string {
	var tmp []bson.M
	err := s.metadata.Find(where).Select(bson.M{"uuid": 1}).All(&tmp)
	if err != nil {
		log.Panic(err)
	}
	var res = []string{}
	for _, uuid := range tmp {
		res = append(res, uuid["uuid"].(string))
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
