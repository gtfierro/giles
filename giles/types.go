package giles

import (
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
)

type SmapReading struct {
	Readings [][]interface{}
	UUID     string `json:"uuid"`
}

type SmapMessage struct {
	Readings   *SmapReading
	Contents   []string
	Metadata   bson.M
	Actuator   bson.M
	Properties bson.M
	UUID       string
	Path       string
}

func (sm *SmapMessage) ToJson() []byte {
	towrite := map[string]*SmapReading{}
	towrite[sm.Path] = sm.Readings
	b, err := json.Marshal(towrite)
	if err != nil {
		log.Error("Error marshalling to JSON", err)
		return []byte{}
	}
	return b
}

