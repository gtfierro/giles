package archiver

import (
	"bytes"
	uuidlib "code.google.com/p/go-uuid/uuid"
	"errors"
	capn "github.com/glycerine/go-capnproto"
	qsr "github.com/gtfierro/giles/internal/quasarcapnp"
	"net"
	"sync"
)

// This is a translator interface for Quasar
// (https://github.com/SoftwareDefinedBuildings/quasar) that implements the
// TSDB interface (look at interfaces.go).
type QDB struct {
	addr       *net.TCPAddr
	cm         *ConnectionMap
	store      MetadataStore
	packetpool sync.Pool
	bufferpool sync.Pool
}

type QuasarReading struct {
	seg *capn.Segment
	req *qsr.Request
	ins *qsr.CmdInsertValues
}

// Create a new reference to a Quasar instance running at ip:port.  Connections
// for a unique stream identifier will be kept alive for `connectionkeepalive`
// seconds. All communicaton with Quasar is done over a TCP connection that
// speaks Capn Proto (http://kentonv.github.io/capnproto/). Quasar can also
// provide a direct HTTP interface, but we choose to implement only the Capn
// Proto interface for more efficient transport.
func NewQuasar(address *net.TCPAddr, connectionkeepalive int) *QDB {
	log.Notice("Conneting to Quasar at %v...", address.String())
	return &QDB{addr: address,
		cm: NewConnectionMap(connectionkeepalive),
		packetpool: sync.Pool{
			New: func() interface{} {
				seg := capn.NewBuffer(nil)
				req := qsr.NewRootRequest(seg)
				req.SetEchoTag(0)
				ins := qsr.NewCmdInsertValues(seg)
				ins.SetSync(false)
				return QuasarReading{
					seg: seg,
					req: &req,
					ins: &ins,
				}
			},
		},
		bufferpool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 200)) // 200 byte buffer
			},
		},
	}
}

func (q *QDB) receive(conn *net.Conn, limit int32) (SmapResponse, error) {
	var sr = SmapResponse{}
	seg, err := capn.ReadFromStream(*conn, nil)
	if err != nil {
		log.Error("Error receiving data from Quasar %v", err)
		return sr, err
	}
	resp := qsr.ReadRootResponse(seg)

	switch resp.Which() {
	case qsr.RESPONSE_VOID:
		if resp.StatusCode() != qsr.STATUSCODE_OK {
			log.Error("Received error status code when writing: %v", resp.StatusCode())
		}
	case qsr.RESPONSE_RECORDS:
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

func (q *QDB) Add(sb *StreamBuf) bool {
	if len(sb.readings) == 0 {
		return false
	}
	uuid := uuidlib.Parse(sb.uuid)
	qr := q.packetpool.Get().(QuasarReading)
	qr.ins.SetUuid([]byte(uuid))
	rl := qsr.NewRecordList(qr.seg, len(sb.readings))
	rla := rl.ToArray()
	for i, val := range sb.readings {
		time := convertTime(val[0].(uint64), sb.unitOfTime, UOT_NS)
		rla[i].SetTime(int64(time))
		rla[i].SetValue(val[1].(float64))
	}
	qr.ins.SetValues(rl)
	qr.req.SetInsertValues(*qr.ins)
	buf := q.bufferpool.Get().(*bytes.Buffer)
	_, err := qr.seg.WriteTo(buf)
	if err != nil {
		log.Error("Error writing %v", err)
		return false
	}
	data := buf.Bytes()
	q.cm.Add(sb.uuid, &data, q)
	q.packetpool.Put(qr)
	return true
}

func (q *QDB) queryNearestValue(uuids []string, start uint64, limit int32, backwards bool) ([]SmapResponse, error) {
	var ret = make([]SmapResponse, len(uuids))
	for i, uu := range uuids {
		seg := capn.NewBuffer(nil)
		req := qsr.NewRootRequest(seg)
		qnv := qsr.NewCmdQueryNearestValue(seg)
		qnv.SetBackward(backwards)
		uuid := uuidlib.Parse(uu)
		qnv.SetUuid([]byte(uuid))
		time := convertTime(start, UOT_S, UOT_NS)
		qnv.SetTime(int64(time))
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
	start = convertTime(start, uot, UOT_MS)
	return q.queryNearestValue(uuids, start, limit, true)
}

func (q *QDB) Next(uuids []string, start uint64, limit int32, uot UnitOfTime) ([]SmapResponse, error) {
	start = convertTime(start, uot, UOT_MS)
	return q.queryNearestValue(uuids, start, limit, false)
}

func (q *QDB) GetData(uuids []string, start uint64, end uint64, uot UnitOfTime) ([]SmapResponse, error) {
	var ret = make([]SmapResponse, len(uuids))
	start = convertTime(start, uot, UOT_NS)
	end = convertTime(end, uot, UOT_NS)
	for i, uu := range uuids {
		seg := capn.NewBuffer(nil)
		req := qsr.NewRootRequest(seg)
		qnv := qsr.NewCmdQueryStandardValues(seg)
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

func (q *QDB) AddStore(s MetadataStore) {
	q.store = s
}
