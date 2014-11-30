package archiver

import (
	uuidlib "code.google.com/p/go-uuid/uuid"
	"errors"
	"github.com/SoftwareDefinedBuildings/quasar/cpinterface"
	capn "github.com/glycerine/go-capnproto"
	"net"
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
func NewQuasar(address net.TCPAddr, connectionkeepalive int) *QDB {
	log.Notice("Conneting to Quasar at %v...", address.String())
	return &QDB{addr: &address,
		cm: NewConnectionMap(connectionkeepalive)}
}

func (q *QDB) receive(conn *net.Conn, limit int32) (SmapResponse, error) {
	var sr = SmapResponse{}
	seg, err := capn.ReadFromStream(*conn, nil)
	if err != nil {
		log.Error("Error receiving data from Quasar %v", err)
		return sr, err
	}
	resp := cpinterface.ReadRootResponse(seg)

	switch resp.Which() {
	case cpinterface.RESPONSE_VOID:
		if resp.StatusCode() != cpinterface.STATUSCODE_OK {
			log.Error("Received error status code when writing: %v", resp.StatusCode())
		}
	case cpinterface.RESPONSE_RECORDS:
		if resp.StatusCode() != 0 {
			return sr, errors.New("Error when reading from Quasar:" + resp.StatusCode().String())
		}
		sr.Readings = [][]float64{}
		log.Debug("limit %v, num values %v", limit, len(resp.Records().Values().ToArray()))
		for i, rec := range resp.Records().Values().ToArray() {
			if limit > -1 && int32(i) >= limit {
				break
			}
			sr.Readings = append(sr.Readings, []float64{float64(rec.Time() * 1000), rec.Value()})
		}
		return sr, nil
	}
	return sr, nil

}

func (q *QDB) Add(sr *SmapReading) bool {
	if len(sr.Readings) == 0 {
		return false
	}
	uuid := uuidlib.Parse(sr.UUID)
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
	q.receive(&conn, -1)
	return true
}

func (q *QDB) queryNearestValue(uuids []string, start uint64, limit int32, backwards bool) ([]SmapResponse, error) {
	var ret = make([]SmapResponse, len(uuids))
	for i, uu := range uuids {
		seg := capn.NewBuffer(nil)
		req := cpinterface.NewRootRequest(seg)
		qnv := cpinterface.NewCmdQueryNearestValue(seg)
		qnv.SetBackward(backwards)
		uuid := uuidlib.Parse(uu)
		qnv.SetUuid([]byte(uuid))
		qnv.SetTime(int64(start))
		req.SetQueryNearestValue(qnv)
		conn, err := q.GetConnection()
		if err != nil {
			log.Error("Error getting connection %v", err)
			return ret, err
		}
		_, err = seg.WriteTo(conn) // here, ignoring # bytes written
		if err != nil {
			return ret, err
		}
		sr, err := q.receive(&conn, limit)
		sr.UUID = uu
		ret[i] = sr
	}
	return ret, nil
}

// Currently, I haven't figured out the beset way to get Quasar to get me responses to
// queries such as "the last 10 values before now". Currently, Prev and Next will
// just return the single closest value
func (q *QDB) Prev(uuids []string, start uint64, limit int32, uot UnitOfTime) ([]SmapResponse, error) {
	return q.queryNearestValue(uuids, start, limit, true)
}

func (q *QDB) Next(uuids []string, start uint64, limit int32, uot UnitOfTime) ([]SmapResponse, error) {
	start = convertTime(start, uot, UOT_MS)
	return q.queryNearestValue(uuids, start, limit, false)
}

func (q *QDB) GetData(uuids []string, start uint64, end uint64, uot UnitOfTime) ([]SmapResponse, error) {
	var ret = make([]SmapResponse, len(uuids))
	start = convertTime(start, uot, UOT_MS)
	end = convertTime(end, uot, UOT_MS)
	for i, uu := range uuids {
		seg := capn.NewBuffer(nil)
		req := cpinterface.NewRootRequest(seg)
		qnv := cpinterface.NewCmdQueryStandardValues(seg)
		uuid := uuidlib.Parse(uu)
		qnv.SetUuid([]byte(uuid))
		qnv.SetStartTime(int64(start))
		qnv.SetEndTime(int64(end))
		req.SetQueryStandardValues(qnv)
		conn, err := q.GetConnection()
		if err != nil {
			log.Error("Error getting connection %v", err)
			return ret, err
		}
		_, err = seg.WriteTo(conn) // here, ignoring # bytes written
		if err != nil {
			return ret, err
		}
		sr, err := q.receive(&conn, -1)
		sr.UUID = uu
		ret[i] = sr
	}
	return ret, nil
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
