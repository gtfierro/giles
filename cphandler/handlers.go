package cphandler

import (
	"bytes"
	capn "github.com/glycerine/go-capnproto"
	"github.com/gtfierro/giles/archiver"
	"github.com/op/go-logging"
	"net"
	"os"
	"strconv"
)

var log = logging.MustGetLogger("cphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

func Handle(a *archiver.Archiver, port int) {

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		log.Error("Error resolving UDP address for capn proto: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Error("Error on listening: %v", err)
	}
	log.Notice("Starting CapnProto on %v", addr.String())
	defer conn.Close()
	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		buffer := bytes.NewBuffer(buf[:n])
		segment, err := capn.ReadFromStream(buffer, nil)
		if err != nil {
			log.Debug("Error recv: %v", err)
		}
		req := ReadRootRequest(segment)
		switch req.Which() {

		case REQUEST_WRITEDATA:
			AddReadings(a, req)

		case REQUEST_QUERY:
			DoQuery(a, req)

		case REQUEST_VOID:
			log.Debug("got a void")
		}

	}
}

func AddReadings(a *archiver.Archiver, req Request) {
	smapmsgs := CapnpToStruct(req.WriteData().Messages().ToArray())
	a.AddData(smapmsgs, req.Apikey())
}

func DoQuery(a *archiver.Archiver, req Request) {
	//res, err := a.HandleQuery(req.Query().Query(), req.Apikey())
}
