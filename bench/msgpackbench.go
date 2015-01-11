package main

import (
	uuid "code.google.com/p/go-uuid/uuid"
	"github.com/gtfierro/giles/mphandler"
	"github.com/ugorji/go/codec"
	"log"
	"net"
	"runtime"
	"sync"
)

const (
	NUM_STREAMS  = 100
	NUM_READINGS = 200000
)

var mh codec.MsgpackHandle

var wg sync.WaitGroup

type MsgPackSmap struct {
	Path       string
	UUID       string `codec:"uuid"`
	Key        string `codec:"key"`
	Properties map[string]interface{}
	Metadata   map[string]interface{}
	Readings   [][2]interface{}
}

var writepool = sync.Pool{
	New: func() interface{} {
		return MsgPackSmap{
			Path:       "/path/0",
			Key:        "jgkiXElqZwAIItiOruwjv87EjDbKpng2OocC1TjVbo4jeZ61QBqvE5eHQ5AvsSsNO-v9DunHlhjwJWd9npo_RA==",
			Properties: map[string]interface{}{},
			Metadata:   map[string]interface{}{},
			Readings:   make([][2]interface{}, 1, 1),
		}
	},
}

func addMessage(msg *[]byte) {
	payload_length := uint32(len(*msg) + 3) // 3 bytes for header
	header := make([]byte, 3, 3)
	header[0] = byte(payload_length & 0xff)
	header[1] = byte(payload_length << 8)
	header[2] = mphandler.DATA_WRITE
	(*msg) = append(header, *msg...)
}

func writeMsgPack(conn *net.Conn, uuid string, time, reading uint64, buf []byte) {
	mps := writepool.Get().(MsgPackSmap)
	mps.UUID = uuid
	mps.Readings[0][0] = time
	mps.Readings[0][1] = reading

	encoder := codec.NewEncoderBytes(&buf, &mh)
	encoder.Encode(mps)
	addMessage(&buf)
	_, err := (*conn).Write(buf)
	if err != nil {
		log.Fatal("error writing", err)
	}
	writepool.Put(mps)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Println("NUM streams:", NUM_STREAMS)
	wg.Add(NUM_STREAMS)
	for i := 0; i < NUM_STREAMS; i++ {
		go func() {
			uuid := uuid.NewUUID().String()
			log.Println("UUID:", uuid)
			conn, err := net.Dial("tcp", "0.0.0.0:8003")
			if err != nil {
				log.Println("ERROR:", err)
			}
			buf := []byte{}
			for x := 1; x < NUM_READINGS+1; x++ {
				writeMsgPack(&conn, uuid, 1351043722500+uint64(x), 1, buf)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
