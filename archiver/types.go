package archiver

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

// Struct representing data readings to and from sMAP
type SmapReading struct {
	// Readings will be interpreted as a list of [uint64, float64] = [time, value]
	// OR as a lsit of [uint64, []byte] = [time, value]
	Readings [][]interface{}
	// Unique identifier for this stream
	UUID string `json:"uuid"`
}

type SmapNumberReading struct {
	// uint64 timestamp
	Time uint64
	// value associated with this timestamp
	Value float64
}

func (s *SmapNumberReading) MarshalJSON() ([]byte, error) {
	floatString := strconv.FormatFloat(s.Value, 'f', -1, 64)
	timeString := strconv.FormatUint(s.Time, 10)
	return json.Marshal([]json.Number{json.Number(timeString), json.Number(floatString)})
}

type SmapObjectReading struct {
	// uint64 timestamp
	Time uint64
	// value associated with this timestamp
	Value interface{}
}

func (s *SmapObjectReading) MarshalJSON() ([]byte, error) {
	timeString := strconv.FormatUint(s.Time, 10)
	return json.Marshal([]interface{}{json.Number(timeString), s.Value})
}

type SmapNumbersResponse struct {
	Readings []*SmapNumberReading
	UUID     string `json:"uuid"`
}

type SmapObjectResponse struct {
	Readings []*SmapObjectReading
	UUID     string `json:"uuid"`
}

type SmapItem struct {
	Data interface{}
	UUID string `json:"uuid"`
}

// This is the general-purpose struct for all INCOMING sMAP messages. This struct
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
	UOT_NS UnitOfTime = 1
	// microseconds 1000000
	UOT_US UnitOfTime = 2
	// milliseconds 1000
	UOT_MS UnitOfTime = 3
	// seconds 1
	UOT_S UnitOfTime = 4
)

type StreamType uint

const (
	OBJECT_STREAM StreamType = iota
	NUMERIC_STREAM
)

// Handy function to transform a []SmapReading into something msgpack friendly
func transformSR(srs []SmapReading) []map[string]interface{} {
	result := make([]map[string]interface{}, len(srs))
	for idx, sr := range srs {
		m := make(map[string]interface{})
		m["uuid"] = sr.UUID
		m["Readings"] = sr.Readings
		result[idx] = m
	}
	return result
}

// Handy function to transform a []SmapReading into something msgpack friendly
func transformSmapNumResp(srs []SmapNumbersResponse) []map[string]interface{} {
	result := make([]map[string]interface{}, len(srs))
	for idx, sr := range srs {
		m := make(map[string]interface{})
		m["uuid"] = sr.UUID
		m["Readings"] = make([][]interface{}, len(sr.Readings))
		for idx, rdg := range sr.Readings {
			m["Readings"].([][]interface{})[idx] = []interface{}{rdg.Time, rdg.Value}
		}
		result[idx] = m
	}
	return result
}

func transformSmapItem(srs []*SmapItem) []map[string]interface{} {
	result := make([]map[string]interface{}, len(srs))
	for idx, sr := range srs {
		m := make(map[string]interface{})
		m["uuid"] = sr.UUID
		m["Data"] = sr.Data
		result[idx] = m
	}
	return result
}
