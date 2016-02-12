package archiver

import (
	"fmt"
	UUID "github.com/pborman/uuid"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"testing"
)

var configfile = "archiver_test.cfg"
var config = LoadConfig(configfile)
var a = NewArchiver(config)

var SmapMsgPool1Path = &sync.Pool{
	New: func() interface{} {
		return map[string]*SmapMessage{
			"/sensor1": &SmapMessage{
				Readings: [][]interface{}{
					[]interface{}{uint64(100), float64(1)},
				},
				Metadata: bson.M{},
			},
		}
	},
}

var SmapMsgPool10Path = &sync.Pool{
	New: func() interface{} {
		ret := make(map[string]*SmapMessage, 10)
		for i := 0; i < 10; i++ {
			ret[fmt.Sprintf("/sensor%s", i)] = &SmapMessage{
				UUID: UUID.New(),
				Readings: [][]interface{}{
					[]interface{}{uint64(100), float64(1)},
				},
				Metadata: bson.M{},
			}
		}
		return ret
	},
}

var SmapMsgPool10PathMetadata = &sync.Pool{
	New: func() interface{} {
		ret := make(map[string]*SmapMessage, 10)
		for i := 0; i < 10; i++ {
			ret[fmt.Sprintf("/sensor%s", i)] = &SmapMessage{
				UUID: UUID.New(),
				Readings: [][]interface{}{
					[]interface{}{uint64(100), float64(1)},
				},
				Metadata: bson.M{"key1": "val1", "key2": "val2"},
			}
		}
		return ret
	},
}

func BenchmarkArchiverAddData1(b *testing.B) {
	uuid := UUID.New()
	a.manager.DeleteKeyByName("name", "email")
	api, err := a.manager.NewKey("name", "email", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < b.N; i++ {
		msgs := SmapMsgPool1Path.Get()
		msgs.(map[string]*SmapMessage)["/sensor1"].UUID = uuid
		err = a.AddData(msgs.(map[string]*SmapMessage), api)
		if err != nil {
			return
		}
		SmapMsgPool1Path.Put(msgs)
	}
}

func BenchmarkArchiverAddData10(b *testing.B) {
	a.manager.DeleteKeyByName("name", "email")
	api, err := a.manager.NewKey("name", "email", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < b.N; i++ {
		msgs := SmapMsgPool10Path.Get()
		err = a.AddData(msgs.(map[string]*SmapMessage), api)
		if err != nil {
			return
		}
		SmapMsgPool10Path.Put(msgs)
	}
}

func BenchmarkArchiverAddData1WithMetadata(b *testing.B) {
	uuid := UUID.New()
	a.manager.DeleteKeyByName("name", "email")
	api, err := a.manager.NewKey("name", "email", true)
	metadata := bson.M{"key1": "val1", "key2": "val2"}
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < b.N; i++ {
		msgs := SmapMsgPool1Path.Get()
		msgs.(map[string]*SmapMessage)["/sensor1"].UUID = uuid
		msgs.(map[string]*SmapMessage)["/sensor1"].Metadata = metadata
		err = a.AddData(msgs.(map[string]*SmapMessage), api)
		if err != nil {
			return
		}
		SmapMsgPool1Path.Put(msgs)
	}
}

func BenchmarkArchiverAddData10WithMetadata(b *testing.B) {
	a.manager.DeleteKeyByName("name", "email")
	api, err := a.manager.NewKey("name", "email", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < b.N; i++ {
		msgs := SmapMsgPool10PathMetadata.Get()
		err = a.AddData(msgs.(map[string]*SmapMessage), api)
		if err != nil {
			return
		}
		SmapMsgPool10PathMetadata.Put(msgs)
	}
}
