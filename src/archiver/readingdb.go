package main

import (
	_ "code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

var streamids = make(map[string]uint32)
var maxstreamid uint32 = 0
var streamlock sync.Mutex

func getStreamid(uuid string) uint32 {
	streamlock.Lock()
	defer streamlock.Unlock()
	if streamids[uuid] == 0 {
		atomic.AddUint32(&maxstreamid, 1)
		streamids[uuid] = maxstreamid
	}
	return streamids[uuid]
}

type RDB struct {
	sync.Mutex
	addr *net.TCPAddr
	conn net.Conn
	In   chan *[]byte
}

func NewReadingDB(address string) *RDB {
	tcpaddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Panic("Error resolving TCP address", address, err)
		return nil
	}
	rdb := &RDB{addr: tcpaddr, In: make(chan *[]byte)}
	return rdb
}

func (rdb *RDB) Connect() {
	if rdb.conn != nil {
		rdb.conn.Close()
	}
	conn, err := net.DialTCP("tcp", nil, rdb.addr)
	conn.SetKeepAlive(true)
	if err != nil {
		log.Panic("Error connecting to ReadingDB", rdb.addr, err)
		return
	}
	rdb.conn = conn
}

func (rdb *RDB) DoWrites() {
	for b := range rdb.In {
		if len((*b)) == 0 {
			continue
		}
		n, err := rdb.conn.Write((*b))
		if err != nil {
			log.Println("Error writing data to ReadingDB", err, len((*b)), n)
			rdb.Connect()
		}
		rdb.conn.Write([]byte{'\n'})
		var recv []byte
		n, _ = rdb.conn.Read(recv)
		if n > 0 {
			log.Println("got back", recv)
		}
	}
}

func (rdb *RDB) Add(sr *SmapReading) bool {
	if rdb.conn == nil {
		log.Panic("RDB is not connected")
		return false
	}
	if len(sr.Readings) == 0 {
		log.Println("No readings")
		return false
	}
	var seqno uint32 = 0
	var timestamp uint32
	var value float64
	//streamid := uuid.Parse(sr.UUID)
	streamid := getStreamid(sr.UUID)
	readingset := &ReadingSet{Streamid: &streamid, Substream: &seqno, Data: make([](*Reading), len(sr.Readings))}
	for i, reading := range sr.Readings {
		timestamp = uint32(reading[0])
		value = float64(reading[1])
		(*readingset).Data[i] = &Reading{Timestamp: &timestamp, Seqno: &seqno, Value: &value}
		//log.Println(timestamp, value, streamid, sr.UUID)
	}

	data, err := proto.Marshal(readingset)
	if err != nil {
		log.Panic("Error marshaling ReadingSet", err)
		return false
	}

	x := new(ReadingSet)
	proto.Unmarshal(data, x)
	println(x.Streamid, x.Data)

	rdb.In <- &data
	//_, err = rdb.conn.Write(data)
	//if err != nil {
	//    log.Panic("Error writing data to ReadingDB", err)
	//    rdb.Connect()
	//}

	//println(readingset.Streamid)
	//println((*readingset.Data[0].Timestamp))
	//println((*readingset.Data[0].Seqno))
	//println((*readingset.Data[0].Value))

	return true
}
