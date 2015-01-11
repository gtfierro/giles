package main

import (
	uuid "code.google.com/p/go-uuid/uuid"
	"github.com/gtfierro/giles/mphandler"
	"github.com/ugorji/go/codec"
	"log"
	"net"
	"sync"
)

const (
	NUM_STREAMS  = 10
	NUM_READINGS = 200
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
	payload_length := uint32(len(*msg) + 3) // 1 byte for length-length, 4 for payload_length, 1 for msg type
	log.Println("payload length is ", payload_length)
	header := make([]byte, 3, 3)
	header[0] = byte(payload_length & 0xff)
	header[1] = byte(payload_length << 8)
	header[2] = mphandler.DATA_WRITE
	log.Println("header", header)
	(*msg) = append(header, *msg...)
	log.Println(mphandler.ParseHeader(msg, 0))
}

func writeMsgPack(conn *net.Conn, uuid string, time, reading uint64, buf []byte) {
	mps := writepool.Get().(MsgPackSmap)
	mps.UUID = uuid
	mps.Readings[0][0] = time
	mps.Readings[0][1] = reading

	encoder := codec.NewEncoderBytes(&buf, &mh)
	encoder.Encode(mps)
	addMessage(&buf)
	log.Println(mps.Readings)
	_, err := (*conn).Write(buf)
	if err != nil {
		log.Println("error writing", err)
	}
	writepool.Put(mps)
}

func main() {
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
				writeMsgPack(&conn, uuid, 1351043722500+uint64(x), uint64(x), buf)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
