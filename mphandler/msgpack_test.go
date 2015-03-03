package mphandler

import (
	UUID "code.google.com/p/go-uuid/uuid"
	"github.com/gtfierro/giles/archiver"
	"github.com/gtfierro/msgpack"
	"github.com/ugorji/go/codec"
	"os"
	"sync"
	"testing"
)

/** Archiver setup **/
var configfile = "mphandler_test.cfg"
var config = archiver.LoadConfig(configfile)
var a = archiver.NewArchiver(config)

var mh codec.MsgpackHandle

type MsgPackSmap struct {
	Path       string
	UUID       string `codec:"uuid"`
	Key        string `codec:"key"`
	Properties map[string]interface{}
	Metadata   map[string]interface{}
	Readings   [][2]interface{}
}

func TestMain(m *testing.M) {
	go HandleUDP(a, *config.MsgPack.UdpPort)
	os.Exit(m.Run())
}

var MsgPackPool1 = &sync.Pool{
	New: func() interface{} {
		msg := MsgPackSmap{
			Path:     "/sensor1",
			UUID:     UUID.New(),
			Readings: make([][2]interface{}, 1),
		}
		msg.Readings[0] = [2]interface{}{uint64(100), float64(1)}
		bytes := make([]byte, 400)
		msgpack.Encode(msg, &bytes)
		return bytes
	},
}

func BenchmarkAddReading1(b *testing.B) {
	msg := MsgPackSmap{
		Path:     "/sensor1",
		UUID:     UUID.New(),
		Readings: make([][2]interface{}, 1),
	}
	msg.Readings[0] = [2]interface{}{uint64(100), int64(1)}
	bytes := []byte{}
	encoder := codec.NewEncoderBytes(&bytes, &mh)
	encoder.Encode(msg)
	for i := 0; i < b.N; i++ {
		handleUDPConn(a, bytes)
	}
}

func BenchmarkAddReading1Metadata(b *testing.B) {
	msg := MsgPackSmap{
		Path:     "/sensor1",
		UUID:     UUID.New(),
		Metadata: map[string]interface{}{"Site": "Test Site", "Nested": map[string]interface{}{"key": "value", "other": "value"}},
		Readings: make([][2]interface{}, 1),
	}
	msg.Readings[0] = [2]interface{}{uint64(100), int64(1)}
	bytes := []byte{}
	encoder := codec.NewEncoderBytes(&bytes, &mh)
	encoder.Encode(msg)
	for i := 0; i < b.N; i++ {
		handleUDPConn(a, bytes)
	}
}
