package main

import (
	"encoding/json"
	simplejson "github.com/bitly/go-simplejson"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
	"strings"
)

type SmapReading struct {
	Readings [][]uint64
	Metadata interface{}
	UUID     string
}

type SmapMessage struct {
	Readings   *SmapReading
	Metadata   bson.M
	Properties bson.M
	UUID       string
	path       string
}

func processJSON(bytes *[]byte) ([](*SmapReading), error) {
	var reading map[string]*json.RawMessage
	var dest [](*SmapReading)
	err := json.Unmarshal(*bytes, &reading)
	if err != nil {
		return dest, err
	}

	for _, v := range reading {
		if v == nil {
			continue
		}
		var sr SmapReading
		err = json.Unmarshal(*v, &sr)
		if err != nil {
			return nil, err
		}
		dest = append(dest, &sr)
	}
	return dest, nil
}

/*
  We receive the following keys:
  - Metadata: send directly to mongo, if we can
  - Actuator: send directly to mongo, if we can
  - uuid: parse this out for adding to timeseries
  - Readings: parse these out for adding to timeseries
  - Properties: send to mongo, but need to parse out ReadingType to help with parsing Readings
*/
func handleJSON(bytes *[]byte) ([](*SmapMessage), error) {
	/*
	 * we receive a bunch of top-level keys that we don't know, so we unmarshal them into a
	 * map, and then parse each of the internal objects individually
	 */

	var ret [](*SmapMessage)
	var e error
	var rawmessage map[string]*json.RawMessage
	err := json.Unmarshal(*bytes, &rawmessage)
	if err != nil {
		return ret, err
	}

	/*
	   global metadata
	   We populate this for every non-endpoint path
	   we come across
	*/
	pathmetadata := make(map[string]interface{})
	isendpoint := true

	for path, reading := range rawmessage {
		isendpoint = true

		js, err := simplejson.NewJson([]byte(*reading))
		if err != nil {
			e = err
		}

		// get uuid
		uuid := js.Get("uuid").MustString("")
		if uuid == "" { // no UUID means no endpoint.
			isendpoint = false
		}

		//get metadata
		localmetadata := js.Get("Metadata").MustMap()

		// if not endpoint, set metadata for this path and then exit
		if !isendpoint {
			pathmetadata[path] = localmetadata
			continue
		}

		message := &SmapMessage{path: path, UUID: uuid}

		if localmetadata != nil {
			message.Metadata = bson.M(localmetadata)
		}

		// get properties
		properties := js.Get("Properties").MustMap()
		if properties != nil {
			message.Properties = bson.M(properties)
		}

		readingarray := js.Get("Readings").MustArray()
		sr := &SmapReading{UUID: uuid}
		srs := make([][]uint64, len(readingarray))
		for idx, readings := range readingarray {
			reading := readings.([]interface{})
			ts, _ := strconv.ParseUint(string(reading[0].(json.Number)), 10, 64)
			val, _ := strconv.ParseUint(string(reading[1].(json.Number)), 10, 64)
			srs[idx] = []uint64{ts, val}
		}
		sr.Readings = srs
		message.Readings = sr

		//TODO get actuator

		ret = append(ret, message)

	}
	//loop through all path metadata and apply to messages
	for prefix, md := range pathmetadata {
		for idx, msg := range ret {
			if (*msg).Metadata == nil {
				(*msg).Metadata = bson.M(md.(map[string]interface{}))
				ret[idx] = msg
				break
			}
			if strings.HasPrefix((*msg).path, prefix) {
				for k, v := range md.(map[string]interface{}) {
					if (*msg).Metadata[k] == nil {
						(*msg).Metadata[k] = v
					}
				}
				ret[idx] = msg
			}
		}
	}
	return ret, e
}
