package cphandler

import (
	//capn "github.com/glycerine/go-capnproto"
	"github.com/gtfierro/giles/archiver"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// Takes in an array of CapnProto SmapMessage and converts each into the
// native SmapMessage type as defined by the archiver, and then returns
// a constructed map of (path, sMAP object)
func CapnpToStruct(messages []SmapMessage) map[string]*archiver.SmapMessage {
	ret := map[string]*archiver.SmapMessage{}
	for _, msg := range messages {
		sm := &archiver.SmapMessage{Path: msg.Path(),
			UUID:       string(msg.Uuid()),
			Metadata:   bson.M{},
			Properties: bson.M{}}
		// Contents
		sm.Contents = msg.Contents().ToArray()

		// Metadata
		for _, pair := range msg.Metadata().ToArray() {
			key := strings.Replace(pair.Key(), "/", ".", -1)
			sm.Metadata[key] = pair.Value()
		}

		// Properties
		for _, pair := range msg.Properties().ToArray() {
			key := strings.Replace(pair.Key(), "/", ".", -1)
			sm.Properties[key] = pair.Value()
		}

		// Readings
		sr := &archiver.SmapReading{}
		sr.UUID = string(msg.Uuid())
		srs := make([][]interface{}, msg.Readings().Len())
		for idx, smr := range msg.Readings().ToArray() {
			srs[idx] = []interface{}{smr.Time(), smr.Data()}
		}
		sr.Readings = srs
		ret[msg.Path()] = sm
	}
	return ret
}
