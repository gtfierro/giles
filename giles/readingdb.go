package giles

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"net"
	"strconv"
)

var streamids = make(map[string]uint32)
var maxstreamid uint32 = 0

type Header struct {
	Type   MessageType
	Length uint32
}

type Message struct {
	header *Header
	data   []byte
}

type SmapResponse struct {
	Readings [][]float64
	UUID     string `json:"uuid"`
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
		log.Error("error committing streamid")
		return nil
	}
	var substream uint32 = 0

	// create ReadingSet
	readingset := &ReadingSet{Streamid: &streamid,
		Substream: &substream,
		Data:      make([](*Reading), len(sr.Readings))}
	// populate readings
	for i, reading := range sr.Readings {
		timestamp = reading[0].(uint64)
		value = reading[1].(float64)
		seqno = uint64(i)
		(*readingset).Data[i] = &Reading{Timestamp: &timestamp, Seqno: &seqno, Value: &value}
	}

	// marshal for sending over wire
	data, err := proto.Marshal(readingset)
	if err != nil {
		log.Panic("Error marshaling ReadingSet:", err)
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
	addr *net.TCPAddr
	conn net.Conn
	In   chan *[]byte
	cm   *ConnectionMap
}

func NewReadingDB(ip string, port int, connectionkeepalive int) *RDB {
	log.Notice("Connecting to ReadingDB at %v:%v...", ip, port)
	address := ip + ":" + strconv.Itoa(port)
	tcpaddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Panic("Error resolving TCP address", address, err)
		return nil
	}
	log.Notice("...connected!")
	rdb := &RDB{addr: tcpaddr,
		In: make(chan *[]byte),
		cm: &ConnectionMap{streams: map[string]*Connection{}, keepalive: connectionkeepalive}}
	return rdb
}

func (rdb *RDB) GetConnection() (net.Conn, error) {
	conn, err := net.DialTCP("tcp", nil, rdb.addr)
	if err == nil {
		conn.SetKeepAlive(true)
	}
	return conn, err
}

// Add deposits incoming readings in order for them to be sent to the database.
// When Add returns, the client should be guaranteed that the writes will be
// committed to the underlying store. Returns True if there were readings to be committed,
// and False if there were no readings found in the incoming message
func (rdb *RDB) Add(sr *SmapReading) bool {
	if len(sr.Readings) == 0 {
		return false
	}

	m := NewMessage(sr)

	data := m.ToBytes()
	rdb.cm.Add(sr.UUID, &data)

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
		log.Error("Error writing data to ReadingDB", err)
		return sr, err
	}
	sr, err = rdb.receiveData(conn)
	return sr, err
}

/*
  Retrieves the last [limit] readings before (and including)
  [ref] for all streams that match query [w]

  [limit] defaults to 1
*/
func (rdb *RDB) Prev(uuids []string, ref uint64, limit int32) ([]SmapResponse, error) {
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
		u_limit := uint32(limit)
		query := &Nearest{Streamid: &sid, Substream: &substream,
			Reference: &ref, Direction: &direction, N: &u_limit}
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
func (rdb *RDB) Next(uuids []string, ref uint64, limit int32) ([]SmapResponse, error) {
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
		u_limit := uint32(limit)
		query := &Nearest{Streamid: &sid, Substream: &substream,
			Reference: &ref, Direction: &direction, N: &u_limit}
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
func (rdb *RDB) receiveData(conn *net.Conn) (SmapResponse, error) {
	// buffer for received bytes
	var sr = SmapResponse{}
	var err error
	recv := make([]byte, 2048)
	n, _ := (*conn).Read(recv)
	recv = recv[:n] // truncate to the length of known valid data
	// message type is first 4 bytes TODO: use it?
	_ = binary.BigEndian.Uint32(recv[:4])
	msglen := binary.BigEndian.Uint32(recv[4:8])
	// for now, assume the message is a ReadingDB Response protobuf
	response := &Response{}
	// remaining length is the expected length of message - how much we've already received
	var remaining_length = uint32(0)
	if uint32(n-8) != msglen {
		remaining_length = msglen - uint32(n) + 8
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
		log.Error("Error receiving data from Readingdb:", err)
		return sr, err
	}
	data := response.GetData()
	if data == nil {
		log.Error("No data returned from Readingdb")
		return sr, err
	}
	//sr.UUID = uuid
	sr.Readings = [][]float64{}
	for _, rdg := range data.GetData() {
		//TODO: this *1000 should probably generalized for the unitoftime
		sr.Readings = append(sr.Readings, []float64{float64(*rdg.Timestamp) * 1000, *rdg.Value})
	}
	return sr, err
}

func (rdb *RDB) LiveConnections() int {
	return rdb.cm.LiveConnections()
}
