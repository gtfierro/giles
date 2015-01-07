package mphandler

import (
	"github.com/gtfierro/giles/archiver"
	"github.com/op/go-logging"
	"net"
	"os"
)

var log = logging.MustGetLogger("mphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

func Handle(a *archiver.Archiver) {
	log.Notice("Handling MsgPack")
}

func ServeTCP(a *archiver.Archiver, tcpaddr *net.TCPAddr) {
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
	buf := make([]byte, 1024)
	for {
		n, _ := conn.Read(buf)
		if n == 0 {
			continue
		}
		leftover, decoded := decode(buf[:n])
		log.Debug("leftover length: %v", len(leftover))
		AddReadings(a, decoded)
	}
}

func AddReadings(a *archiver.Archiver, md map[string]interface{}) {
	log.Debug("in: %v", md)
	apikey := md["key"].(string)
	ret := map[string]*archiver.SmapMessage{}
	log.Debug("md:%v", md)
	sm := &archiver.SmapMessage{Path: md["Path"].(string),
		UUID:     md["uuid"].(string),
		Readings: make([][]interface{}, 0, len(md["Readings"].([]interface{}))),
	}
	log.Debug("len of readings %v", len(md["Readings"].([]interface{})))
	for _, rdg := range md["Readings"].([]interface{}) {
		sm.Readings = append(sm.Readings, []interface{}{rdg.([]interface{})[0].(uint64), rdg.([]interface{})[1].(uint64)})
	}
	a.AddData(ret, apikey)
}
