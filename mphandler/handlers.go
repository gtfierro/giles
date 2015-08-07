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
	"github.com/gtfierro/msgpack"
	"github.com/op/go-logging"
	"net"
	"os"
	"strconv"
)

const (
	BUFFER_SIZE = 65536
)

var log = logging.MustGetLogger("mphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

func HandleTCP(a *archiver.Archiver, port int) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		log.Error("Error resolving TCP address for msgpack %v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		log.Error("Error on listening: %v", err)
	}

	log.Notice("Starting MsgPack on TCP %v", tcpaddr.String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting connection: %v", err)
		}
		go handleConn(a, conn)
	}
}

func HandleUDP(a *archiver.Archiver, port int) {
	udpaddr, err := net.ResolveUDPAddr("udp6", "[::]:"+strconv.Itoa(port))
	if err != nil {
		log.Error("Error resolving UDP address for msgpack %v", err)
	}

	conn, err := net.ListenUDP("udp6", udpaddr)
	if err != nil {
		log.Error("Error on listening (%v)", err)
	}

	log.Notice("Starting MsgPack on UDP %v", udpaddr.String())
	for {
		buf := make([]byte, 1024)
		n, fromaddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Error("Problem reading connection %v (%v)", fromaddr, err)
		}
		if n > 0 {
			go handleUDPConn(a, buf[:n])
		}
	}
}

func handleUDPConn(a *archiver.Archiver, buf []byte) {
	_, decoded := msgpack.Decode(&buf, 0)
	AddReadings(a, decoded.(map[string]interface{}))
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
		n, err := conn.Read(buf)
		if n == 0 { // didn't read anything
			continue
		}
		if err != nil {
			log.Error("Reading socket error: %v", err)
			break
		}
		offset := 0
		if leftover > 0 { // have a partial packet we need to finish reading
			if needheader { // need to read header first
				dec = append(old[:readalready], buf[:leftover]...)
				_, packetlength := ParseHeader(&dec, 0)
				dec = buf[leftover : packetlength+leftover]
				offset = leftover + packetlength - 3
			} else {
				dec = append(old[:readalready], buf[:leftover]...)
				offset = leftover
			}
			_, decoded := msgpack.Decode(&dec, 0)
			AddReadings(a, decoded.(map[string]interface{}))
			old = old[:cap(old)]
			readalready = 0
			leftover = 0
			needheader = false
			if offset == n {
				continue
			}
		}

		for { // read/decode packets until no room
			if offset == n {
				break
			}
			if BUFFER_SIZE-offset < 3 {
				copy(old, buf[offset:])
				readalready = len(buf[offset:])
				leftover = 3 - readalready
				needheader = true
				break
			} else {
				_, packetlength := ParseHeader(&buf, offset)
				offset += 3
				packetlength -= 3
				if offset+packetlength <= BUFFER_SIZE { // still have room
					newoffset, decoded := msgpack.Decode(&buf, offset)
					AddReadings(a, decoded.(map[string]interface{}))
					offset = newoffset
				} else { // not enough!
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
		Readings: make([]archiver.Reading, 0, len(md["Readings"].([]interface{}))),
	}
	for _, rdg := range md["Readings"].([]interface{}) {
		var timestamp uint64
		if timestamp_uint64, ok := rdg.([]interface{})[0].(uint64); ok {
			timestamp = timestamp_uint64
		} else {
			timestamp = uint64(rdg.([]interface{})[0].(int64))
		}
		if reading, ok := rdg.([]interface{})[1].(int64); ok {
			sm.Readings = append(sm.Readings, &archiver.SmapNumberReading{timestamp, float64(reading)})
		} else if freading, ok := rdg.([]interface{})[1].(float64); ok {
			sm.Readings = append(sm.Readings, &archiver.SmapNumberReading{timestamp, freading})
		}
	}
	ret[sm.Path] = sm
	a.AddData(ret, md["key"].(string))
}
