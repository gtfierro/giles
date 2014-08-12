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
	var maxsid uint32 = 0
	return &Store{session: session, db: db, streams: streams, maxsid: &maxsid}
}

func (s *Store) GetStreamId(uuid string) {
	streamid := &RDBStreamId{}
	err := s.streams.Find(bson.M{"uuid": uuid}).One(&streamid)
	if err != nil {
		// not found, so create
		log.Println(err)
		streamlock.Lock()
		streamid.StreamId = (*s.maxsid)
		streamid.UUID = uuid
		inserterr := s.streams.Insert(streamid)
		if inserterr != nil {
			log.Panic(inserterr)
		}
		atomic.AddUint32(s.maxsid, 1)
		streamlock.Unlock()
	}
	log.Println(uuid, streamid.UUID, streamid.StreamId)
}
