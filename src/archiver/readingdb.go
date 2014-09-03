package main

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"sync"
)

var streamids = make(map[string]uint32)
var maxstreamid uint32 = 0
var streamlock sync.Mutex

type Header struct {
	Type   MessageType
	Length uint32
}

type Message struct {
	header *Header
	data   []byte
}

type SmapResponse struct {
	Readings [][]uint64
	UUID     string
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
	var streamid uint32 = store.GetStreamId(sr.UUID)
	if streamid == 0 {
		log.Println("error committing streamid")
		return nil
	}
	var substream uint32 = 0

	// create ReadingSet
	readingset := &ReadingSet{Streamid: &streamid,
		Substream: &substream,
		Data:      make([](*Reading), len(sr.Readings))}
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

func NewReadingDB(ip string, port int) *RDB {
	address := ip + ":" + strconv.Itoa(port)
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
	if err != nil {
		log.Panic("Error connecting to ReadingDB: ", rdb.addr, err)
		return
	}
	conn.SetKeepAlive(true)
	rdb.conn = conn
}

//TODO: explore having a different channel for each UUID.
// too many connections? Keep pool of last N UUIDs and have
// those keep channels open for writing. Likely we are not
// saturating the link.
//TODO: net/http benchmarking
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
		return false
	}

	m := NewMessage(sr)

	data := m.ToBytes()
	rdb.In <- &data

	return true
}

//TODO: figure out return values here
/**
 * For all the ReadingDB methods, we need to remember that this should really try to act like a
   standalone package (ish). Given this constraint, we should not require using methods from the
   metadata store. These methods will return SmapResponse structs
**/

/*
  Retrieves the most recent [limit] readings from
  all streams that match query [w]

  [limit] defaults to 1
  TODO: just have this accept a list of streamids
*/
func (rdb *RDB) Latest(sq *SmapQuery, limit uint64) {

}

/*
  Retrieves the last [limit] readings before (and including)
  [ref] for all streams that match query [w]

  [limit] defaults to 1
  TODO: just have this accept a list of streamids
*/
func (rdb *RDB) Prev(sq *SmapQuery, ref, limit uint64) {
}

/*
  Retrieves the last [limit] readings after (and including)
  [ref] for all streams that match query [w]

  [limit] defaults to 1
  TODO: just have this accept a list of streamids
*/
func (rdb *RDB) Next(sq *SmapQuery, ref, limit uint64) {
}

/*
  Retrieves all data between (and including) [start] and [end]
  for all streams matching query [w]
  TODO: just have this accept a list of streamids
*/
func (rdb *RDB) GetData(uuids []string, start, end uint64) ([]SmapResponse, error) {
	var err error
	var retdata = []SmapResponse{}
	var substream uint32 = 0
	var action uint32 = 1
	for _, uuid := range uuids {
		sid := store.GetStreamId(uuid)
		sid = 32
		m := &Message{}
		query := &Query{Streamid: &sid, Substream: &substream,
			Starttime: &start, Endtime: &end, Action: &action}
		data, err := proto.Marshal(query)
		h := &Header{Type: MessageType_QUERY, Length: uint32(len(data))}
		m.header = h
		m.data = data

		n, err := rdb.conn.Write(m.ToBytes())
		if err != nil {
			log.Println("Error writing data to ReadingDB", err, len((data)), n)
			rdb.Connect()
		}
		sr, err := rdb.ReceiveData()
		if err != nil {
			return retdata, err
		}
		retdata = append(retdata, sr)
	}
	return retdata, err
}

/*
 * Listens for data coming from ReadingDB
**/
func (rdb *RDB) ReceiveData() (SmapResponse, error) {
	// buffer for received bytes
	var sr = SmapResponse{}
	var err error
	recv := make([]byte, 1024)
	n, _ := rdb.conn.Read(recv)
	// message type is first 4 bytes TODO: use it?
	msglen := binary.BigEndian.Uint32(recv[4:8])
	// for now, assume the message is a ReadingDB Response protobuf
	response := &Response{}
	if msglen <= uint32(n) {
		err = proto.Unmarshal(recv[8:msglen+8], response)
		if err != nil {
			log.Println("Error receiving data from Readingdb", err)
			return sr, err
		}
		data := response.GetData()
		if data == nil {
			log.Println("No data returned from Readingdb")
			return sr, err
		}
		//sr.UUID = uuid
		sr.Readings = [][]uint64{}
		for _, rdg := range data.GetData() {
			sr.Readings = append(sr.Readings, []uint64{*rdg.Timestamp, uint64(*rdg.Value)})
		}
	} else {
		//TODO read more bytes
	}
	return sr, err
}

/*
  Retrieves all data between (and including) [start] and [end]
  for all streams with a uuid in [uuids]
*/
func (rdb *RDB) DataUUID(uuids []string, start, end uint64) {
}
