package cphandler

import (
	"bytes"
	capn "github.com/glycerine/go-capnproto"
	"github.com/gtfierro/giles/archiver"
	"github.com/op/go-logging"
	"net"
	"os"
)

var log = logging.MustGetLogger("cphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

func Handle(a *archiver.Archiver) {
	log.Notice("Handling Capn Proto")
}

func ServeUDP(a *archiver.Archiver, udpaddr *net.UDPAddr) {
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		log.Error("Error on listening: %v", err)
	}
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
			log.Debug("got a writedata")
			AddReadings(a, req)

		case REQUEST_VOID:
			log.Debug("got a void")
		}

	}
}

func AddReadings(a *archiver.Archiver, req Request) {
	smapmsgs := CapnpToStruct(req.WriteData().Messages().ToArray())
	a.AddData(smapmsgs, req.Apikey())
}
