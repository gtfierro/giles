package mphandler

import (
	"github.com/gtfierro/giles/archiver"
	"github.com/op/go-logging"
	"github.com/ugorji/go/codec"
	"net"
	"os"
	"reflect"
)

var log = logging.MustGetLogger("mphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)
var mh codec.MsgpackHandle

func Handle(a *archiver.Archiver) {
	log.Notice("Handling MsgPack")
}

func ServeTCP(a *archiver.Archiver, tcpaddr *net.TCPAddr) {
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	listener, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		log.Error("Error on listening: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Error("Error accepting connection: %v", err)
			}
			go handleConn(a, conn)
		}
	}()
}

func handleConn(a *archiver.Archiver, conn net.Conn) {
	var v interface{} // value to decode/encode into
	buf := make([]byte, 1024)
	for {
		n, _ := conn.Read(buf)
		if n == 0 {
			continue
		}
		log.Debug("in: %v", buf[:n])
		dec := codec.NewDecoderBytes(buf[:n], &mh)
		dec.Decode(&v)
		AddReadings(a, v.(map[string]interface{}))
	}
}

func AddReadings(a *archiver.Archiver, input map[string]interface{}) {
	apikey := string(input["key"].([]uint8))
	ret := map[string]*archiver.SmapMessage{}
	for path, md := range input {
		m, ok := md.(map[string]interface{})
		if !ok {
			continue
		}
		if readings, found := m["Readings"]; found {
			uuid := string(m["uuid"].([]uint8))
			sm := &archiver.SmapMessage{Path: path,
				UUID: uuid,
			}
			if metadata, found := m["Metadata"]; found {
				sm.Metadata = metadata.(map[string]interface{})
			}
			if properties, found := m["Properties"]; found {
				sm.Properties = properties.(map[string]interface{})
			}
			sr := &archiver.SmapReading{UUID: uuid}
			srs := make([][]interface{}, len(readings.([]interface{})))
			for idx, smr := range readings.([]interface{}) {
				time, ok := smr.([]interface{})[0].(uint64)
				if !ok { // is int64
					time = uint64(smr.([]interface{})[0].(int64))
				}
				if value, ok := smr.([]interface{})[1].(float64); !ok {
					srs[idx] = []interface{}{time, float64(smr.([]interface{})[1].(int64))}
				} else {
					srs[idx] = []interface{}{time, value}
				}
			}
			sr.Readings = srs
			sm.Readings = sr
			ret[path] = sm
		}
	}
	a.AddData(ret, apikey)
}
