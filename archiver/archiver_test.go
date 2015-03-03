package archiver

import (
	UUID "code.google.com/p/go-uuid/uuid"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"testing"
)

var configfile = "../giles.cfg"
var config = LoadConfig(configfile)
var a = NewArchiver(config)

var SmapMsgPool = &sync.Pool{
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

func BenchmarkArchiverAddData(b *testing.B) {
	uuid := UUID.New()
	a.store.delkey_byname("name", "email")
	api, err := a.store.newkey("name", "email", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < b.N; i++ {
		msgs := SmapMsgPool.Get()
		msgs.(map[string]*SmapMessage)["/sensor1"].UUID = uuid
		err = a.AddData(msgs.(map[string]*SmapMessage), api)
		if err != nil {
			return
		}
		SmapMsgPool.Put(msgs)
	}
}
