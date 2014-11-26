package archiver

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
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

func (q *QDB) receive(conn *net.Conn) {
	seg, err := capn.ReadFromStream(*conn, nil)
	if err != nil {
		log.Error("Error receiving data from Quasar %v", err)
		return
	}
	req := cpinterface.ReadRootResponse(seg)

	switch req.Which() {
	case cpinterface.RESPONSE_VOID:
		if req.StatusCode() != cpinterface.STATUSCODE_OK {
			log.Error("Received error status code when writing %v", req.StatusCode())
		}
	case cpinterface.RESPONSE_RECORDS:
		parseResponseRecords(&req)
	}

}

func (q *QDB) Add(sr *SmapReading) bool {
	if len(sr.Readings) == 0 {
		return false
	}
	uuid := uuid.Parse(sr.UUID)
	seg := capn.NewBuffer(nil)
	req := cpinterface.NewRootRequest(seg)
	req.SetEchoTag(0)
	ins := cpinterface.NewCmdInsertValues(seg)
	ins.SetUuid([]byte(uuid))
	rl := cpinterface.NewRecordList(seg, len(sr.Readings))
	rla := rl.ToArray()
	for i, val := range sr.Readings {
		rla[i].SetTime(int64(val[0].(uint64)))
		rla[i].SetValue(val[1].(float64))
	}
	ins.SetValues(rl)
	ins.SetSync(false)
	req.SetInsertValues(ins)
	conn, err := q.GetConnection()
	if err != nil {
		log.Error("Error getting connection %v", err)
		return false
	}
	_, err = seg.WriteTo(conn)
	if err != nil {
		log.Error("Error writing %v", err)
		return false
	}
	q.receive(&conn)
	return true
}

func (q *QDB) Prev(uuids []string, start uint64, limit int32) ([]SmapResponse, error) {
	seg := capn.NewBuffer(nil)
	req := cpinterface.NewRootRequest(seg)
	qnv := req.QueryNearestValue()
	qnv.SetBackward(true) // set to query previous values
	uuid := uuid.Parse(uuids[0])
	qnv.SetUuid([]byte(uuid))
	log.Debug("prev UUID: %v", uuid)
	qnv.SetTime(int64(start))
	req.SetQueryNearestValue(qnv)
	conn, err := q.GetConnection()
	log.Debug("writing %v echo tag %v", req.Which(), req.EchoTag())
	if err != nil {
		log.Error("Error getting connection %v", err)
		return []SmapResponse{}, err
	}
	_, err = seg.WriteTo(conn)
	if err != nil {
		return []SmapResponse{}, err
	}
	q.receive(&conn)
	return []SmapResponse{}, nil
}

func (q *QDB) Next(uuids []string, start uint64, limit int32) ([]SmapResponse, error) {
	seg := capn.NewBuffer(nil)
	req := cpinterface.NewRootRequest(seg)
	qnv := req.QueryNearestValue()
	qnv.SetBackward(false) // set to query next values
	uuid := uuid.Parse(uuids[0])
	qnv.SetUuid([]byte(uuid))
	log.Debug("next UUID: %v", uuid)
	qnv.SetTime(int64(start))
	req.SetQueryNearestValue(qnv)
	conn, err := q.GetConnection()
	log.Debug("writing %v echo tag %v", req.Which(), req.EchoTag())
	if err != nil {
		log.Error("Error getting connection %v", err)
		return []SmapResponse{}, err
	}
	_, err = seg.WriteTo(conn)
	if err != nil {
		return []SmapResponse{}, err
	}
	q.receive(&conn)
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

func parseResponseRecords(resp *cpinterface.Response) ([]SmapResponse, error) {
	var ret []SmapResponse
	if resp.StatusCode() != 0 {
		return ret, errors.New("Error when reading from Quasar:" + resp.StatusCode().String())
	}
	for i, rec := range resp.Records().Values().ToArray() {
		log.Debug("idx %v value %v", i, rec)
	}
	return ret, nil
}
