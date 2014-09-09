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

func (rdb *RDB) GetConnection() (net.Conn, error) {
	conn, err := net.DialTCP("tcp", nil, rdb.addr)
	if err == nil {
		conn.SetKeepAlive(true)
	}
	return conn, err
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

/**
 * For all the ReadingDB methods, we need to remember that this should really try to act like a
   standalone package (ish). Given this constraint, we should not require using methods from the
   metadata store. These methods will return SmapResponse structs
**/

/**
 * What's the common functionality for all the methods? Sending and receiving
**/
func (rdb *RDB) sendAndReceive(payload []byte, msgtype MessageType, conn *net.Conn) (SmapResponse, error) {
	var sr SmapResponse
	var err error
	m := &Message{}
	h := &Header{Type: msgtype, Length: uint32(len(payload))}
	m.header = h
	m.data = payload
	_, err = (*conn).Write(m.ToBytes())
	if err != nil {
		log.Println("Error writing data to ReadingDB", err)
		return sr, err
	}
	sr, err = rdb.ReceiveData(conn)
	return sr, err
}

/*
  Retrieves the last [limit] readings before (and including)
  [ref] for all streams that match query [w]

  [limit] defaults to 1
*/
func (rdb *RDB) Prev(uuids []string, ref uint64, limit uint32) ([]SmapResponse, error) {
	var err error
	var retdata = []SmapResponse{}
	var data []byte
	var substream uint32 = 0
	var direction = Nearest_PREV
	var sr SmapResponse

	for _, uuid := range uuids {
		conn, err := rdb.GetConnection()
		if err != nil {
			return retdata, err
		}
		sid := store.GetStreamId(uuid)
		query := &Nearest{Streamid: &sid, Substream: &substream,
			Reference: &ref, Direction: &direction, N: &limit}
		data, err = proto.Marshal(query)
		sr, err = rdb.sendAndReceive(data, MessageType_NEAREST, &conn)
		sr.UUID = uuid
		retdata = append(retdata, sr)
	}
	return retdata, err
}

/*
  Retrieves the last [limit] readings after (and including)
  [ref] for all streams that match query [w]

  [limit] defaults to 1
*/
func (rdb *RDB) Next(uuids []string, ref uint64, limit uint32) ([]SmapResponse, error) {
	var err error
	var retdata = []SmapResponse{}
	var data []byte
	var substream uint32 = 0
	var direction = Nearest_NEXT
	var sr SmapResponse

	for _, uuid := range uuids {
		conn, err := rdb.GetConnection()
		if err != nil {
			return retdata, err
		}
		sid := store.GetStreamId(uuid)
		query := &Nearest{Streamid: &sid, Substream: &substream,
			Reference: &ref, Direction: &direction, N: &limit}
		data, err = proto.Marshal(query)
		sr, err = rdb.sendAndReceive(data, MessageType_NEAREST, &conn)
		sr.UUID = uuid
		retdata = append(retdata, sr)
	}
	return retdata, err
}

/*
  Retrieves all data between (and including) [start] and [end]
  for all streams matching query [w]
*/
func (rdb *RDB) GetData(uuids []string, start, end uint64) ([]SmapResponse, error) {
	if start > end {
		start, end = end, start
	}
	var err error
	var retdata = []SmapResponse{}
	var data []byte
	var substream uint32 = 0
	var action uint32 = 1
	var sr SmapResponse
	for _, uuid := range uuids {
		conn, err := rdb.GetConnection()
		if err != nil {
			return retdata, err
		}
		sid := store.GetStreamId(uuid)
		query := &Query{Streamid: &sid, Substream: &substream,
			Starttime: &start, Endtime: &end, Action: &action}
		data, err = proto.Marshal(query)
		sr, err = rdb.sendAndReceive(data, MessageType_QUERY, &conn)
		sr.UUID = uuid
		retdata = append(retdata, sr)
	}
	return retdata, err
}

/*
 * Listens for data coming from ReadingDB
**/
func (rdb *RDB) ReceiveData(conn *net.Conn) (SmapResponse, error) {
	// buffer for received bytes
	var sr = SmapResponse{}
	var err error
	recv := make([]byte, 2048)
	n, _ := (*conn).Read(recv)
	recv = recv[:n] // truncate to the length of known valid data
	// message type is first 4 bytes TODO: use it?
	msglen := binary.BigEndian.Uint32(recv[4:8])
	// for now, assume the message is a ReadingDB Response protobuf
	response := &Response{}
	// remaining length is the expected length of message - how much we've already received
	var remaining_length = uint32(0)
	if uint32(n-8) != msglen {
		remaining_length = msglen - uint32(n)
	}
	for {
		// base case
		if remaining_length <= 0 {
			break
		}
		buffer_length := min(2048, remaining_length)
		newrecv := make([]byte, buffer_length)
		bytes_read, _ := (*conn).Read(newrecv)
		recv = append(recv, newrecv[:bytes_read]...)
		remaining_length = remaining_length - uint32(bytes_read)
	}
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
	return sr, err
}
