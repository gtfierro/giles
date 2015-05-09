package archiver

import (
	"github.com/golang/protobuf/proto"
	"encoding/binary"
	rdbp "github.com/gtfierro/giles/internal/readingdbproto"
	"net"
)

var streamids = make(map[string]uint32)
var maxstreamid uint32 = 0

// Because we can send different types of protobuf messages,
// we include this prefixed header to all outgoing packets
// to ReadingDB to identify what kind of action we are doing
type header struct {
	Type   rdbp.MessageType
	Length uint32
}

type Message struct {
	header *header
	data   []byte
}

/*
   for now, assume all Smap Readings have same uuid. In the future
   We will probably want to queue up the serialization of a bunch
   and then write in bulk.
*/
func NewMessage(sb *StreamBuf, store MetadataStore) *Message {
	m := &Message{}
	//var timestamp uint64
	//var value float64
	//var seqno uint64
	//var streamid uint32 = store.GetStreamId(sb.uuid)
	//if streamid == 0 {
	//	log.Error("error committing streamid")
	//	return nil
	//}
	//var substream uint32 = 0

	//// create ReadingSet
	//readingset := &rdbp.ReadingSet{Streamid: &streamid,
	//	Substream: &substream,
	//	Data:      make([](*rdbp.Reading), len(sb.readings), len(sb.readings))}
	//// populate readings
	//for i, reading := range sb.readings {
	//	timestamp = reading[0].(uint64)
	//	value = reading[1].(float64)
	//	seqno = uint64(i)
	//	(*readingset).Data[i] = &rdbp.Reading{Timestamp: &timestamp, Seqno: &seqno, Value: &value}
	//}

	//// marshal for sending over wire
	//data, err := proto.Marshal(readingset)
	//if err != nil {
	//	log.Panic("Error marshaling ReadingSet:", err)
	//	return nil
	//}

	//// create header
	//h := &header{Type: rdbp.MessageType_READINGSET, Length: uint32(len(data))}
	//m.header = h
	//m.data = data
	return m
}

func (m *Message) ToBytes() []byte {
	onthewire := make([]byte, 8)
	binary.BigEndian.PutUint32(onthewire, uint32(m.header.Type))
	binary.BigEndian.PutUint32(onthewire[4:8], m.header.Length)
	onthewire = append(onthewire, m.data...)
	return onthewire
}

// This is a translator interface for ReadingDB (adaptive branch --
// https://github.com/SoftwareDefinedBuildings/readingdb/tree/adaptive) that
// implements the TSDB interface (look at interfaces.go)
type RDB struct {
	addr  *net.TCPAddr
	In    chan *[]byte
	cm    *ConnectionMap
	store MetadataStore
}

// Create a new reference to a ReadingDB instance running at ip:port.
// Connections for a unique stream identifier will be kept alive for
// `connectionkeepalive` seconds. All communicaton with ReadingDB is done over
// a TCP connection that speaks protobuf
// (https://developers.google.com/protocol-buffers/). For a description and
// implementation of ReadingDB protobuf, please see
// https://github.com/gtfierro/giles/archiver/internal/readingdbproto
func NewReadingDB(address *net.TCPAddr, connectionkeepalive int) *RDB {
	log.Notice("Connecting to ReadingDB at %v...", address.String())
	log.Notice("...connected!")
	rdb := &RDB{addr: address,
		In: make(chan *[]byte),
		cm: NewConnectionMap(connectionkeepalive)}
	return rdb
}

func (rdb *RDB) GetConnection() (net.Conn, error) {
	conn, err := net.DialTCP("tcp", nil, rdb.addr)
	if err == nil {
		conn.SetKeepAlive(true)
	}
	return conn, err
}

func (rdb *RDB) AddStore(store MetadataStore) {
	rdb.store = store
}

// Add deposits incoming readings in order for them to be sent to the database.
// When Add returns, the client should be guaranteed that the writes will be
// committed to the underlying store. Returns True if there were readings to be committed,
// and False if there were no readings found in the incoming message
func (rdb *RDB) Add(sb *StreamBuf) bool {
	//if sb.readings == nil || len(sb.readings) == 0 {
	//	return false
	//}
	m := NewMessage(sb, rdb.store)

	data := m.ToBytes()
	rdb.cm.Add(sb.uuid, &data, rdb)

	return true
}

// Sends a packet, constructs header, and then listens on that connection and returns the response
func (rdb *RDB) sendAndReceive(payload []byte, msgtype rdbp.MessageType, conn *net.Conn) (SmapReading, error) {
	var sr SmapReading
	var err error
	m := &Message{}
	h := &header{Type: msgtype, Length: uint32(len(payload))}
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

// Retrieves the last [limit] readings before (and including)
// [ref] for all streams that match query [w]
// [limit] defaults to 1
func (rdb *RDB) Prev(uuids []string, ref uint64, limit int32, query_uot UnitOfTime) ([]SmapReading, error) {
	var err error
	var retdata = []SmapReading{}
	var data []byte
	var substream uint32 = 0
	var direction = rdbp.Nearest_PREV
	var sr SmapReading
	ref = convertTime(ref, query_uot, UOT_MS)

	for _, uuid := range uuids {
		conn, err := rdb.GetConnection()
		if err != nil {
			return retdata, err
		}
		sid := rdb.store.GetStreamId(uuid)
		u_limit := uint32(limit)
		query := &rdbp.Nearest{Streamid: &sid, Substream: &substream,
			Reference: &ref, Direction: &direction, N: &u_limit}
		data, err = proto.Marshal(query)
		sr, err = rdb.sendAndReceive(data, rdbp.MessageType_NEAREST, &conn)
		sr.UUID = uuid
		retdata = append(retdata, sr)
	}
	return retdata, err
}

// Retrieves the last [limit] readings after (and including)
// [ref] for all streams that match query [w]
// [limit] defaults to 1
func (rdb *RDB) Next(uuids []string, ref uint64, limit int32, query_uot UnitOfTime) ([]SmapReading, error) {
	var err error
	var retdata = []SmapReading{}
	var data []byte
	var substream uint32 = 0
	var direction = rdbp.Nearest_NEXT
	var sr SmapReading
	ref = convertTime(ref, query_uot, UOT_MS)

	for _, uuid := range uuids {
		conn, err := rdb.GetConnection()
		if err != nil {
			return retdata, err
		}
		sid := rdb.store.GetStreamId(uuid)
		u_limit := uint32(limit)
		query := &rdbp.Nearest{Streamid: &sid, Substream: &substream,
			Reference: &ref, Direction: &direction, N: &u_limit}
		data, err = proto.Marshal(query)
		sr, err = rdb.sendAndReceive(data, rdbp.MessageType_NEAREST, &conn)
		sr.UUID = uuid
		retdata = append(retdata, sr)
	}
	return retdata, err
}

// Retrieves all data between (and including) [start] and [end]
// for all streams matching query [w]
func (rdb *RDB) GetData(uuids []string, start, end uint64, query_uot UnitOfTime) ([]SmapReading, error) {
	if start > end {
		start, end = end, start
	}
	start = convertTime(start, query_uot, UOT_MS)
	end = convertTime(end, query_uot, UOT_MS)
	var err error
	var retdata = []SmapReading{}
	var data []byte
	var substream uint32 = 0
	var action uint32 = 1
	var sr SmapReading
	for _, uuid := range uuids {
		conn, err := rdb.GetConnection()
		if err != nil {
			return retdata, err
		}
		sid := rdb.store.GetStreamId(uuid)
		query := &rdbp.Query{Streamid: &sid, Substream: &substream,
			Starttime: &start, Endtime: &end, Action: &action}
		data, err = proto.Marshal(query)
		sr, err = rdb.sendAndReceive(data, rdbp.MessageType_QUERY, &conn)
		sr.UUID = uuid
		retdata = append(retdata, sr)
	}
	return retdata, err
}

/*
 * Listens for data coming from ReadingDB
**/
func (rdb *RDB) receiveData(conn *net.Conn) (SmapReading, error) {
	// buffer for received bytes
	var sr = SmapReading{}
	var err error
	recv := make([]byte, 2048)
	n, _ := (*conn).Read(recv)
	recv = recv[:n] // truncate to the length of known valid data
	// message type is first 4 bytes TODO: use it?
	_ = binary.BigEndian.Uint32(recv[:4])
	msglen := binary.BigEndian.Uint32(recv[4:8])
	// for now, assume the message is a ReadingDB Response protobuf
	response := &rdbp.Response{}
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
	sr.Readings = [][]interface{}{}
	for _, rdg := range data.GetData() {
		//TODO: this *1000 should probably generalized for the unitoftime
		sr.Readings = append(sr.Readings, []interface{}{float64(*rdg.Timestamp), *rdg.Value})
	}
	return sr, err
}

func (rdb *RDB) LiveConnections() int {
	return rdb.cm.LiveConnections()
}
