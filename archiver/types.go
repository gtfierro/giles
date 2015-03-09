package archiver

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
)

// an internal type we can use instead of bson.M
type Dict map[string]interface{}

// Struct representing data readings to and from sMAP
type SmapReading struct {
	// Readings will be interpreted as a list of [uint64, float64] = [time, value]
	Readings [][]interface{}
	// Unique identifier for this stream
	UUID string `json:"uuid"`
}

// This is the general-purpose struct for all incoming sMAP messages. This struct
// is designed to match the format of sMAP JSON, as that is the primary data format.
type SmapMessage struct {
	// Readings for this message
	Readings [][]interface{}
	// If this struct corresponds to a sMAP collection,
	// then Contents contains a list of paths contained within
	// this collection
	Contents []string `json:",omitempty"`
	// Map of the metadata
	Metadata bson.M `json:",omitempty"`
	// Map containing the actuator reference
	Actuator bson.M `json:",omitempty"`
	// Map of the properties
	Properties bson.M `json:",omitempty"`
	// Unique identifier for this stream. Should be empty for Collections
	UUID string
	// Path of this stream (thus far)
	Path string
}

// Convenience method to turn a sMAP message into
// marshaled JSON
func (sm *SmapMessage) ToJson() []byte {
	b, err := json.Marshal(sm)
	if err != nil {
		log.Error("Error marshalling to JSON", err)
		return []byte{}
	}
	return b
}

// unit of time indicators
type UnitOfTime uint

const (
	// nanoseconds 1000000000
	UOT_NS = iota
	// microseconds 1000000
	UOT_US
	// milliseconds 1000
	UOT_MS
	// seconds 1
	UOT_S
)
