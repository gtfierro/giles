package archiver

import (
	"github.com/SoftwareDefinedBuildings/quasar"
	"net"
	"strconv"
)

type QDB struct {
	addr  *net.TCPAddr
	q     *quasar.Quasar
	cm    *ConnectionMap
	store *Store
}

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
	return nil, nil
}

func (q *QDB) LiveConnections() int {
	return 0
}

func (q *QDB) AddStore(s *Store) {
	q.store = s
}
