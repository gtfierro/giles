// License?

// Package mphandler implements a MsgPack/TCP interface to the Archiver API
// at http://godoc.org/github.com/gtfierro/giles/archiver
//
// Overview
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
//      +----------------------+----------------------+----
//      | packet len (2 bytes) | packet type (1 byte) | packet contents...
//      +----------------------+----------------------+----
//
// Packet length is 2 bytes. Afterwards comes a single byte that contains the
// packet type (this will be a value from a predetermined Enum that will be
// described below. Following this header comes the actual packet contents

package mphandler

import (
	"github.com/gtfierro/giles/archiver"
	"github.com/op/go-logging"
	"net"
	"os"
)

const (
	BUFFER_SIZE = 16384
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

// How do we efficiently handle lots of packets on a single connection?
// 0. Initialize offset to 0
// 1. Read bytes into the buffer (right now 4k, but may increase?). Possibility of
//    partial packet at end. If we have bytes in our holding buffer, read the number
//    of leftover bytes into a decoding buffer and combine it with the old buffer, decode and increase
//    the offset
// (if enough space for header)
// 2. Read header of packet and retrieve packetlength. Increase offset by header size (3 bytes)
// 3. If enough space left in buffer for packet, read whole packet, decode, and increase the
//    offset by the packetsize.
// 4. If NOT enough space left in buffer, read from the offset til the end of the buffer and place
//    it into a holding buffer and keep track of how many bytes left to read. Go back to step 1

func handleConn(a *archiver.Archiver, conn net.Conn) {
	var dec []byte
	leftover := 0
	readalready := 0
	needheader := false
	old := make([]byte, BUFFER_SIZE)
	buf := make([]byte, BUFFER_SIZE)
	for {
		if len(old) <= 3 {
			log.Debug("TOP LOOP old is now %v", old)
		} else {
			log.Debug("TOP LOOP old is now %v", old[:3])
		}
		n, err := conn.Read(buf)
		if n == 0 { // didn't read anything
			log.Debug("NO READ")
			continue
		}
		if err != nil {
			log.Error("Reading socket error: %v", err)
			break
		}
		log.Debug("read %v", n)
		log.Debug("old is now now %v", old)
		offset := 0
		if leftover > 0 { // have a partial packet we need to finish reading
			if needheader { // need to read header first
				log.Debug("handling rest of header %v", old)
				log.Debug("readalready: %v, leftover: %v", readalready, leftover)
				dec = append(old[:readalready], buf[:leftover]...)
				log.Debug("header : %v, old %v", dec[:3], old[:readalready])
				_, packetlength := ParseHeader(&dec, 0)
				log.Debug("packet len: %v", packetlength)
				dec = buf[leftover : packetlength+leftover]
				log.Debug("first byte of new packet %v", buf, leftover)
				offset = leftover + packetlength
			} else {
				//log.Debug("readalready: %v, leftover: %v", readalready, leftover)
				dec = append(old[:readalready], buf[:leftover]...)
				offset = leftover
			}
			_, decoded := decode(&dec, 0)
			AddReadings(a, decoded.(map[string]interface{}))
			old = old[:cap(old)]
			readalready = 0
			leftover = 0
			needheader = false
			if offset == n {
				continue
			}
		} else {
			if len(old) <= 3 {
				log.Debug("ELSE old is now %v", old)
			} else {
				log.Debug("ELSE old is now %v", old[:3])
			}
		}

		for { // read/decode packets until no room
			if offset == n {
				break
			}
			if BUFFER_SIZE-offset < 3 {
				log.Debug("header overlap %v", buf[offset:])
				old = buf[offset:]
				readalready = len(buf[offset:])
				leftover = 3 - readalready
				needheader = true
				log.Debug("old is now %v", old)
				break
			} else {
				_, packetlength := ParseHeader(&buf, offset)
				offset += 3
				packetlength -= 3
				if offset+packetlength <= BUFFER_SIZE { // still have room
					newoffset, decoded := decode(&buf, offset)
					AddReadings(a, decoded.(map[string]interface{}))
					offset = newoffset
				} else { // not enough!
					if needheader {
						log.Debug("old write")
					}
					copy(old, buf[offset:])
					leftover = packetlength - (BUFFER_SIZE - offset)
					readalready = len(buf[offset:])
					break
				}
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
