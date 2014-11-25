package archiver

import (
	"github.com/SoftwareDefinedBuildings/quasar/cpinterface"
	capn "github.com/glycerine/go-capnproto"
	"net"
	"strconv"
)

// This is a translator interface for Quasar
// (https://github.com/SoftwareDefinedBuildings/quasar) that implements the
// TSDB interface (look at interfaces.go).
type QDB struct {
	addr  *net.TCPAddr
	cm    *ConnectionMap
	store *Store
}

// Create a new reference to a Quasar instance running at ip:port.  Connections
// for a unique stream identifier will be kept alive for `connectionkeepalive`
// seconds. All communicaton with Quasar is done over a TCP connection that
// speaks Capn Proto (http://kentonv.github.io/capnproto/). Quasar can also
// provide a direct HTTP interface, but we choose to implement only the Capn
// Proto interface for more efficient transport.
func NewQuasar(ip string, port int, connectionkeepalive int) *QDB {
	log.Notice("Conneting to Quasar at %v:%v...", ip, port)
	address := ip + ":" + strconv.Itoa(port)
	tcpaddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Panic("Error resolving TCP address", address, err)
		return nil
	}
	log.Notice("...connected!")
	return &QDB{addr: tcpaddr}
}

func (q *QDB) Add(sr *SmapReading) bool {
	if len(sr.Readings) == 0 {
		return false
	}
	//buf := bytes.Buffer{}
	seg := capn.NewBuffer(nil)
	req := cpinterface.NewRootRequest(seg)
	ins := req.InsertValues()
	ins.SetUuid([]byte(sr.UUID))
	rl := cpinterface.NewRecordList(seg, len(sr.Readings))
	rla := rl.ToArray()
	for i, val := range sr.Readings {
		rla[i].SetTime(int64(val[0].(uint64)))
		rla[i].SetValue(val[1].(float64))
	}
	ins.SetValues(rl)
	conn, _ := q.GetConnection()
	seg.WriteTo(conn)
	return true
}

func (q *QDB) Prev(uuids []string, start uint64, limit int32) ([]SmapResponse, error) {
	return []SmapResponse{}, nil
}

func (q *QDB) Next(uuids []string, start uint64, limit int32) ([]SmapResponse, error) {
	return []SmapResponse{}, nil
}

func (q *QDB) GetData(uuids []string, start uint64, end uint64) ([]SmapResponse, error) {
	return []SmapResponse{}, nil
}

func (q *QDB) GetConnection() (net.Conn, error) {
	conn, err := net.DialTCP("tcp", nil, q.addr)
	if err == nil {
		conn.SetKeepAlive(true)
	}
	return conn, err
}

func (q *QDB) LiveConnections() int {
	return 0
}

func (q *QDB) AddStore(s *Store) {
	q.store = s
}
