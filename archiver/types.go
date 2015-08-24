package archiver

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strconv"
	"strings"
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

func (s *SmapNumberReading) GetTime() uint64 {
	return s.Time
}

func (s *SmapNumberReading) IsObject() bool {
	return false
}

func (s *SmapNumberReading) GetValue() interface{} {
	return s.Value
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

func (s *SmapObjectReading) GetTime() uint64 {
	return s.Time
}

func (s *SmapObjectReading) IsObject() bool {
	return true
}

func (s *SmapObjectReading) GetValue() interface{} {
	return s.Value
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
	Readings []Reading
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
	UUID string `json:"uuid"`
	// Path of this stream (thus far)
	Path string
}

func (sm *SmapMessage) UnmarshalJSON(b []byte) (err error) {
	var (
		incoming  = new(IncomingSmapMessage)
		time      uint64
		value_num float64
		value_obj interface{}
	)

	// unmarshal to an intermediary struct that matches the format
	// of the incoming messages
	err = json.Unmarshal(b, incoming)
	if err != nil {
		return
	}

	// copy the values over that we don't need to translate
	sm.UUID = incoming.UUID
	sm.Path = incoming.Path
	sm.Metadata = flatten(incoming.Metadata)
	sm.Properties = flatten(incoming.Properties)
	sm.Actuator = flatten(incoming.Actuator)
	sm.Contents = incoming.Contents

	// convert the readings depending if they are numeric or object
	sm.Readings = make([]Reading, len(incoming.Readings))
	for idx, reading := range incoming.Readings {
		// time should be a uint64 no matter what
		err = json.Unmarshal(reading[0], &time)
		if err != nil {
			return
		}

		// check if we have a numerical value
		err = json.Unmarshal(reading[1], &value_num)
		if err != nil {
			// if we don't, then we treat as an object reading
			err = json.Unmarshal(reading[1], &value_obj)
			sm.Readings[idx] = &SmapObjectReading{time, value_obj}
		} else {
			sm.Readings[idx] = &SmapNumberReading{time, value_num}
		}
	}
	return
}

func (sm *SmapMessage) ToSmapReading() *SmapReading {
	rdg := new(SmapReading)
	rdg.UUID = sm.UUID
	rdg.Readings = make([][]interface{}, len(sm.Readings))
	for idx, r := range sm.Readings {
		rdg.Readings[idx] = []interface{}{r.GetTime(), r.GetValue()}
	}
	return rdg
}

func (sm *SmapMessage) IsTimeseries() bool {
	return sm.UUID != ""
}

func (sm *SmapMessage) HasMetadata() bool {
	return sm.Metadata != nil || sm.Properties != nil || sm.Actuator != nil
}

// Key names like uuid, Metadata.Key, Properties.Key, etc. Fetches the corresponding
// value from the
func (sm *SmapMessage) GetKey(key string) interface{} {
	// try a quick match on top-level keys
	switch key {
	case "uuid":
		return sm.UUID
	case "Path":
		return sm.Path
	case "Metadata":
		return sm.Metadata
	case "Properties":
		return sm.Properties
	case "Actuator":
		return sm.Actuator
	}

	// handle nested keys
	var keyPos = strings.Index(key, ".")
	if keyPos == -1 {
		return nil
	}
	var nestedKey = key[keyPos+1:]
	var section = key[:keyPos]
	fmt.Println("nested key", nestedKey, "sectioN", section)
	switch section {
	case "Metadata":
		return sm.Metadata[nestedKey]
	case "Properties":
		return sm.Properties[nestedKey]
	case "Actuator":
		return sm.Actuator[nestedKey]
	}
	return nil
}

func (sm *SmapMessage) GetValuesFor(q *Query) {
	for _, key := range q.target {
		sm.GetKey(key)
	}
}

// Returns true if the current message contains keys mentioned in the provided list
func (sm *SmapMessage) HasKeysFrom(keys []string) bool {
	var (
		keyPos    int
		nestedKey string
		section   string
		found     bool = false
	)
	if len(keys) == 0 { // the select * case
		return true
	}
	for _, key := range keys {
		// try quick match on top-level keys
		switch key {
		case "uuid":
			return len(sm.UUID) > 0
		case "Path":
			return len(sm.Path) > 0
		case "Metadata":
			return sm.Metadata != nil && len(sm.Metadata) > 0
		case "Properties":
			return sm.Properties != nil && len(sm.Properties) > 0
		case "Actuator":
			return sm.Actuator != nil && len(sm.Actuator) > 0
		}

		keyPos = strings.Index(key, ".")
		if keyPos == -1 { // not a nested key, so false
			return false
		}
		nestedKey = key[keyPos+1:]
		section = key[:keyPos]
		switch section {
		case "Metadata":
			_, found = sm.Metadata[nestedKey]
		case "Properties":
			_, found = sm.Properties[nestedKey]
		case "Actuator":
			_, found = sm.Actuator[nestedKey]
		}
		if found {
			return found
		}
	}
	return false
}

type IncomingSmapMessage struct {
	// Readings for this message
	Readings [][]json.RawMessage
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
	UUID string `json:"uuid"`
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

type TieredSmapMessage map[string]*SmapMessage

// This performs the metadata inheritance for the paths and messages inside
// this collection of SmapMessages. Inheritance starts from the root path "/"
// can progresses towards the leaves.
// First, get a list of all of the potential timeseries (any path that contains a UUID)
// Then, for each of the prefixes for the path of that timeserie (util.getPrefixes), grab
// the paths from the TieredSmapMessage that match the prefixes. Sort these in "decreasing" order
// and apply to the metadata.
// Finally, delete all non-timeseries paths
func (tsm *TieredSmapMessage) CollapseToTimeseries() {
	var (
		prefixMsg *SmapMessage
		found     bool
	)
	for path, msg := range *tsm {
		if !msg.IsTimeseries() {
			continue
		}
		prefixes := getPrefixes(path)
		sort.Sort(sort.Reverse(sort.StringSlice(prefixes)))
		for _, prefix := range prefixes {
			// if we don't find the prefix OR it exists but doesn't have metadata, we skip
			prefixMsg, found = (*tsm)[prefix]
			if !found || prefixMsg == nil || (prefixMsg != nil && !prefixMsg.HasMetadata()) {
				continue
			}
			// otherwise, we apply keys from paths higher up if our timeseries doesn't already have the key
			// (this is reverse inheritance)
			if prefixMsg.Metadata != nil && len(prefixMsg.Metadata) > 0 {
				for k, v := range prefixMsg.Metadata {
					if _, hasKey := msg.Metadata[k]; !hasKey {
						if msg.Metadata == nil {
							msg.Metadata = make(bson.M)
						}
						msg.Metadata[k] = v
					}
				}
			}
			if prefixMsg.Properties != nil && len(prefixMsg.Properties) > 0 {
				for k, v := range prefixMsg.Properties {
					if _, hasKey := msg.Properties[k]; !hasKey {
						if msg.Properties == nil {
							msg.Properties = make(bson.M)
						}
						msg.Properties[k] = v
					}
				}
			}
			if prefixMsg.Actuator != nil && len(prefixMsg.Actuator) > 0 {
				for k, v := range prefixMsg.Actuator {
					if _, hasKey := msg.Actuator[k]; !hasKey {
						if msg.Actuator == nil {
							msg.Actuator = make(bson.M)
						}
						msg.Actuator[k] = v
					}
				}
			}
			(*tsm)[path] = msg
		}
	}
	// when done, delete all non timeseries paths
	for path, msg := range *tsm {
		if !msg.IsTimeseries() {
			delete(*tsm, path)
		}
	}
}

func (tsm *TieredSmapMessage) ToBson() []bson.M {
	var (
		ret = make([]bson.M, len(*tsm))
		idx = 0
	)
	for _, msg := range *tsm {
		msgBson := bson.M{
			"uuid": msg.UUID,
			"Path": msg.Path,
		}
		if msg.Metadata != nil && len(msg.Metadata) > 0 {
			for k, v := range msg.Metadata {
				msgBson["Metadata."+k] = v
			}
		}
		if msg.Properties != nil && len(msg.Properties) > 0 {
			for k, v := range msg.Properties {
				msgBson["Properties."+k] = v
			}
		}
		if msg.Actuator != nil && len(msg.Actuator) > 0 {
			for k, v := range msg.Actuator {
				msgBson["Actuator."+k] = v
			}
		}
		ret[idx] = msgBson
		idx += 1
	}
	return ret
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
