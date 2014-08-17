package main

import (
	"fmt"
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

func (s *Store) Query(stringquery []byte) *[]bson.M {
	var res []bson.M
	query := parse(string(stringquery))
	fmt.Println(*query)
	err := s.metadata.Find(*query).All(&res)
	if err != nil {
		log.Panic(err)
	}
	return &res
}
