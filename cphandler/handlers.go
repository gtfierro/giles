package cphandler

import (
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

func ServeUDP(udpaddr *net.UDPAddr) {
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		log.Error("Error on listening: %v", err)
	}
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Debug("Error recv: %v", err)
		}
		log.Debug("receiv bytes %v", buf[:n])
	}
}
