package main

import (
	_ "code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
	"log"
	"net"
	"sync"
	"sync/atomic"
    "encoding/binary"
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

type Header struct {
	Type   MessageType
	Length uint32
}

type Message struct {
    header *Header
    data []byte
}

/* 
    for now, assume all Smap Readings have same uuid. In the future
    We will probably want to queue up the serialization of a bunch
    and then write in bulk.
*/
func NewMessage(sr *SmapReading) *Message {
  m := &Message{}
  var timestamp uint64
  var value float64
  var seqno uint64
  //TODO: get streamid from smap readings
  var streamid uint32 = getStreamid(sr.UUID)
  var substream uint32 = 0

  // create ReadingSet
  readingset := &ReadingSet{Streamid: &streamid,
                            Substream: &substream,
                            Data: make([](*Reading), len(sr.Readings))}
  // populate readings
  for i, reading := range sr.Readings {
    timestamp = uint64(reading[0])
    value = float64(reading[1])
    seqno = uint64(i)
    (*readingset).Data[i] = &Reading{Timestamp: &timestamp, Seqno: &seqno, Value: &value}
  }

  // marshal for sending over wire
  data, err := proto.Marshal(readingset)
  if err != nil {
    log.Panic("Error marshaling ReadingSet", err)
    return nil
  }

  // create header
  h := &Header{Type: MessageType_READINGSET, Length: uint32(len(data))}
  m.header = h
  m.data = data
  return m
}

func (m *Message) ToBytes() []byte {
  onthewire := make([]byte, 8)
  binary.BigEndian.PutUint32(onthewire, uint32(m.header.Type))
  binary.BigEndian.PutUint32(onthewire[4:8], m.header.Length)
  onthewire = append(onthewire, m.data...)
  return onthewire
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
        // test
        test := &ReadingSet{}
        err := proto.Unmarshal((*b), test)
        if err != nil {
          println("got error unmarshaling", err)
        }
        println(test.GetStreamid(), test.GetSubstream())
        data := test.GetData()
        for _, d := range data {
          println(d.GetTimestamp(), d.GetValue(), d.GetSeqno())
        }

		n, err := rdb.conn.Write((*b))
		if err != nil {
			log.Println("Error writing data to ReadingDB", err, len((*b)), n)
			rdb.Connect()
		}
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

    m := NewMessage(sr)

    data := m.ToBytes()
	rdb.In <- &data

	return true
}
