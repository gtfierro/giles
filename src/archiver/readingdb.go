package main

import (
	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
	"log"
	"net"
)

type RDB struct {
	addr *net.TCPAddr
	Conn net.Conn
}

func NewReadingDB(address string) *RDB {
	tcpaddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Panic("Error resolving TCP address", address, err)
		return nil
	}
	rdb := &RDB{addr: tcpaddr}
	return rdb
}

func (rdb *RDB) Connect() {
	conn, err := net.DialTCP("tcp", nil, rdb.addr)
	if err != nil {
		log.Panic("Error connecting to ReadingDB", rdb.addr, err)
		return
	}
	rdb.Conn = conn
}

func (rdb *RDB) Add(sr *SmapReading) bool {
	if rdb.Conn == nil {
		log.Panic("RDB is not connected")
		return false
	}
	var seqno uint32 = 0
	var timestamp uint32
	var value float64
	streamid, _ := uuid.Parse(sr.UUID).Id()
	readingset := &ReadingSet{Streamid: &streamid, Substream: &seqno, Data: make([](*Reading), len(sr.Readings))}
	for i, reading := range sr.Readings {
		timestamp = uint32(reading[0])
		value = float64(reading[1])
		(*readingset).Data[i] = &Reading{Timestamp: &timestamp, Seqno: &seqno, Value: &value}
	}

	data, err := proto.Marshal(readingset)
	if err != nil {
		log.Panic("Error marshaling ReadingSet", err)
		return false
	}
	_, err = rdb.Conn.Write(data)
	if err != nil {
		log.Panic("Error writing data to ReadingDB", err)
		return false
	}

	//println(readingset.Streamid)
	//println((*readingset.Data[0].Timestamp))
	//println((*readingset.Data[0].Seqno))
	//println((*readingset.Data[0].Value))

	return true
}
