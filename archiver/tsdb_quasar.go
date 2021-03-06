package archiver

import (
	"bytes"
	"errors"
	"fmt"
	capn "github.com/glycerine/go-capnproto"
	qsr "github.com/gtfierro/giles/internal/quasarcapnp"
	uuidlib "github.com/pborman/uuid"
	"net"
	"sync"
)

type QuasarDB struct {
	addr       *net.TCPAddr
	store      MetadataStore
	packetpool sync.Pool
	bufferpool sync.Pool
	connpool   *ConnectionPool
}

type QuasarReading struct {
	seg *capn.Segment
	req *qsr.Request
	ins *qsr.CmdInsertValues
}

func NewQuasarDB(address *net.TCPAddr, maxConnections int) *QuasarDB {
	log.Notice("Connecting to Quasar at %v...", address.String())
	quasar := &QuasarDB{addr: address,
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

	quasar.connpool = NewConnectionPool(quasar.getConnection, maxConnections)
	return quasar
}

func (quasar *QuasarDB) getConnection() *TSDBConn {
	conn, err := net.DialTCP("tcp", nil, quasar.addr)
	if err != nil {
		log.Error("Error getting connection to Quasar (%v)", err)
		return nil
	}
	conn.SetKeepAlive(true)
	return &TSDBConn{conn, false}
}

func (quasar *QuasarDB) GetConnection() (net.Conn, error) {
	return nil, nil
}

func (quasar *QuasarDB) AddStore(s MetadataStore) {
	quasar.store = s
}

func (quasar *QuasarDB) LiveConnections() int {
	return 0
}

func (quasar *QuasarDB) Add(sb *StreamBuf) bool {
	if len(sb.readings) == 0 {
		return false
	}
	conn := quasar.connpool.Get()
	defer quasar.connpool.Put(conn)
	uuid := uuidlib.Parse(sb.uuid)
	qr := quasar.packetpool.Get().(QuasarReading)
	qr.ins.SetUuid([]byte(uuid))
	rl := qsr.NewRecordList(qr.seg, sb.idx)
	rla := rl.ToArray()
	for i, val := range sb.readings[:sb.idx] {
		time := convertTime(val.GetTime(), sb.unitOfTime, UOT_NS)
		rla[i].SetTime(int64(time))
		if num, ok := val.GetValue().(float64); ok {
			rla[i].SetValue(num)
		} else {
			log.Error("Bad number in buffer %v %v", sb.uuid, val)
			return false
		}
	}
	qr.ins.SetValues(rl)
	qr.req.SetInsertValues(*qr.ins)
	qr.seg.WriteTo(conn)
	_, err := quasar.receive(conn, -1)
	if err != nil {
		log.Error("Error writing to quasar %v", err)
		return false
	}
	quasar.packetpool.Put(qr)
	return true
}

func (quasar *QuasarDB) receive(conn *TSDBConn, limit int32) (SmapNumbersResponse, error) {
	var sr = SmapNumbersResponse{}
	seg, err := capn.ReadFromStream(conn, nil)
	if err != nil {
		conn.Close()
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
		sr.Readings = []*SmapNumberReading{}
		log.Debug("limit %v, num values %v", limit, len(resp.Records().Values().ToArray()))
		for i, rec := range resp.Records().Values().ToArray() {
			if limit > -1 && int32(i) >= limit {
				break
			}
			sr.Readings = append(sr.Readings, &SmapNumberReading{Time: uint64(rec.Time()), Value: rec.Value()})
		}
		return sr, nil
	default:
		return sr, fmt.Errorf("Got unexpected Quasar Error code (%v)", resp.StatusCode().String())
	}
	return sr, nil

}

func (quasar *QuasarDB) queryNearestValue(uuids []string, start uint64, limit int32, backwards bool) ([]SmapNumbersResponse, error) {
	var ret = make([]SmapNumbersResponse, len(uuids))
	conn := quasar.connpool.Get()
	defer quasar.connpool.Put(conn)
	for i, uu := range uuids {
		stream_uot := quasar.store.GetUnitOfTime(uu)
		seg := capn.NewBuffer(nil)
		req := qsr.NewRootRequest(seg)
		qnv := qsr.NewCmdQueryNearestValue(seg)
		qnv.SetBackward(backwards)
		uuid := uuidlib.Parse(uu)
		qnv.SetUuid([]byte(uuid))
		qnv.SetTime(int64(start))
		req.SetQueryNearestValue(qnv)
		_, err := seg.WriteTo(conn) // here, ignoring # bytes written
		if err != nil {
			return ret, err
		}
		sr, err := quasar.receive(conn, limit)
		if err != nil {
			return ret, err
		}
		sr.UUID = uu
		for _, reading := range sr.Readings {
			reading.Time = convertTime(reading.Time, UOT_NS, stream_uot)
		}
		ret[i] = sr
	}
	return ret, nil
}

func (quasar *QuasarDB) Prev(uuids []string, start uint64, limit int32, uot UnitOfTime) ([]SmapNumbersResponse, error) {
	start = convertTime(start, uot, UOT_NS)
	return quasar.queryNearestValue(uuids, start, limit, true)
}

func (quasar *QuasarDB) Next(uuids []string, start uint64, limit int32, uot UnitOfTime) ([]SmapNumbersResponse, error) {
	start = convertTime(start, uot, UOT_NS)
	return quasar.queryNearestValue(uuids, start, limit, false)
}

func (quasar *QuasarDB) GetData(uuids []string, start uint64, end uint64, uot UnitOfTime) ([]SmapNumbersResponse, error) {
	var ret = make([]SmapNumbersResponse, len(uuids))
	start = convertTime(start, uot, UOT_NS)
	end = convertTime(end, uot, UOT_NS)
	conn := quasar.connpool.Get()
	defer quasar.connpool.Put(conn)
	for i, uu := range uuids {
		stream_uot := quasar.store.GetUnitOfTime(uu)
		seg := capn.NewBuffer(nil)
		req := qsr.NewRootRequest(seg)
		qnv := qsr.NewCmdQueryStandardValues(seg)
		uuid := uuidlib.Parse(uu)
		qnv.SetUuid([]byte(uuid))
		qnv.SetStartTime(int64(start))
		qnv.SetEndTime(int64(end))
		req.SetQueryStandardValues(qnv)
		_, err := seg.WriteTo(conn) // here, ignoring # bytes written
		if err != nil {
			return ret, err
		}
		sr, err := quasar.receive(conn, -1)
		if err != nil {
			return ret, err
		}
		sr.UUID = uu
		for _, reading := range sr.Readings {
			reading.Time = convertTime(reading.Time, UOT_NS, stream_uot)
		}
		ret[i] = sr
	}
	return ret, nil
}
