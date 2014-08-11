package main

import (
	"gopkg.in/mgo.v2"
	"log"
	"strconv"
)

type RDBStreamId struct {
	StreamId uint32
	UUID     string
}

type Store struct {
	session *mgo.Session
	db      *mgo.Database
	streams *mgo.Collection
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
	return &Store{session: session, db: db, streams: streams}
}
