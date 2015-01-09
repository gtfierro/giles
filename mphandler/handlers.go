// MsgPack Handler Overview
//
// The MsgPack format for sMAP is designed to look very similar to the JSON
// format, while also making it possible to handle different commands (e.g. not
// just reads) as well as permissions including an API key.
//
//      type MsgPackSmap struct {
//      	Path       string
//      	UUID       string `codec:"uuid"`
//      	Key        string `codec:"key"`
//      	Properties map[string]interface{}
//      	Metadata   map[string]interface{}
//      	Readings   [][2]interface{}
//      }
//
// We need to augment this struct with some information in a simple packet
// header that gives us the ability to describe packet length and packet
// command.
//
// Header:
//      +---------------------+----------------------+----------------------+----
//      | len prefix (n bits) | packet len (n bytes) | packet type (1 byte) | packet contents...
//      +---------------------+----------------------+----------------------+----
//
// The length prefix is a huffman coding where the length in bits tells us how
// many bytes come next. Those bytes contain the exact length of the packet (in bytes).
// Afterwards comes a single byte that contains the packet type (this will be a value
// from a predetermined Enum that will be described below. Following this header comes
// the actual packet contents

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
	buf := make([]byte, 4096)
	for {
		n, _ := conn.Read(buf)
		if n == 0 {
			continue
		}
		log.Debug("read %v", n)
		offset := 0
		for {
			newoff, decoded := decode(&buf, offset)
			if md, ok := decoded.(map[string]interface{}); ok {
				AddReadings(a, md)
			} else {
				log.Debug("bad in data: %v", decoded)
			}
			if n == newoff { // finished buffer
				break
			} else { // still stuff in buffer
				offset = newoff
			}
		}
	}
}

//TODO: check for malformed
func AddReadings(a *archiver.Archiver, md map[string]interface{}) {
	ret := map[string]*archiver.SmapMessage{}
	sm := &archiver.SmapMessage{Path: md["Path"].(string),
		UUID:     md["uuid"].(string),
		Readings: make([][]interface{}, 0, len(md["Readings"].([]interface{}))),
	}
	for _, rdg := range md["Readings"].([]interface{}) {
		if reading, ok := rdg.([]interface{})[1].(int64); ok {
			sm.Readings = append(sm.Readings, []interface{}{rdg.([]interface{})[0].(uint64), float64(reading)})
		} else if freading, ok := rdg.([]interface{})[1].(float64); ok {
			sm.Readings = append(sm.Readings, []interface{}{rdg.([]interface{})[0].(uint64), freading})
		}
	}
	ret[sm.Path] = sm
	a.AddData(ret, md["key"].(string))
}
