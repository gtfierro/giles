package main

import (
	uuid "code.google.com/p/go-uuid/uuid"
	"github.com/ugorji/go/codec"
	"log"
	"net"
	"sync"
)

const (
	NUM_STREAMS  = 1
	NUM_READINGS = 1
)

var mh codec.MsgpackHandle

var wg sync.WaitGroup

type MsgPackSmap struct {
	Path       string
	UUID       string `codec:"uuid"`
	Key        string `codec:"key"`
	Properties map[string]interface{}
	Metadata   map[string]interface{}
	Readings   MsgPackReadings
}

type MsgPackReadings struct {
	Readings [][2]interface{}
	UUID     string `codec:"uuid"`
}

var writepool = sync.Pool{
	New: func() interface{} {
		return MsgPackSmap{
			Path:       "/sensor/0",
			Key:        "pihlHaUYQGcgOleO-l5-fg6-WxyPJw76s4orcrpA0JC_v8r1wxZiWu1ODhklLwcs9BAXs6B0Soaggd3mFcJYVw==",
			Properties: map[string]interface{}{},
			Metadata:   map[string]interface{}{},
			Readings: MsgPackReadings{
				Readings: make([][2]interface{}, 1, 1),
			},
		}
	},
}

func writeMsgPack(conn *net.Conn, uuid string, time, reading uint64, buf []byte) {
	mps := writepool.Get().(MsgPackSmap)
	mps.UUID = uuid
	mps.Readings.UUID = uuid
	mps.Readings.Readings[0][0] = time
	mps.Readings.Readings[0][1] = reading

	encoder := codec.NewEncoderBytes(&buf, &mh)
	encoder.Encode(mps)
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
			for x := 0; x < NUM_READINGS; x++ {
				writeMsgPack(&conn, uuid, 90000000+uint64(x), uint64(x), buf)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
