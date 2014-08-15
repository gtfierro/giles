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
	session *mgo.Session
	db      *mgo.Database
	streams *mgo.Collection
	maxsid  *uint32
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
	maxstreamid := &RDBStreamId{}
	streams.Find(bson.M{}).Sort("-streamid").One(&maxstreamid)
	var maxsid uint32 = 1
	if maxstreamid != nil {
		maxsid = maxstreamid.StreamId + 1
	}
	return &Store{session: session, db: db, streams: streams, maxsid: &maxsid}
}

//TODO: restructure to return value
func (s *Store) GetStreamId(uuid string) uint32 {
	streamlock.Lock()
	defer streamlock.Unlock()

	streamid := &RDBStreamId{}
	err := s.streams.Find(bson.M{"uuid": uuid}).One(&streamid)
	if err != nil {
		// not found, so create
		log.Println(err)
		streamid.StreamId = (*s.maxsid)
		streamid.UUID = uuid
		inserterr := s.streams.Insert(streamid)
		if inserterr != nil {
			log.Println(inserterr)
			return 0
		}
		atomic.AddUint32(s.maxsid, 1)
	}
	log.Println(uuid, streamid.UUID, streamid.StreamId)
	return streamid.StreamId
}
