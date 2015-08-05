package httphandler

import (
	"encoding/json"
	"errors"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/gtfierro/giles/archiver"
	"gopkg.in/mgo.v2/bson"
	"io"
	"strconv"
)

/*
  We receive the following keys:
  - Metadata: send directly to mongo, if we can
  - Actuator: send directly to mongo, if we can
  - uuid: parse this out for adding to timeseries
  - Readings: parse these out for adding to timeseries
  - Contents: list of resources underneat this path
  - Properties: send to mongo, but need to parse out ReadingType to help with parsing Readings

This should not do any fancy sMAP-related work; that's a job for the store. Here we just return the
object-versions of all the data.
*/
func HandleJSON(r io.Reader) (map[string]*archiver.SmapMessage, error) {
	/*
	 * we receive a bunch of top-level keys that we don't know, so we unmarshal them into a
	 * map, and then parse each of the internal objects individually
	 */
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	var e error
	var rawmessage map[string]*json.RawMessage
	var decodedjson = map[string]*archiver.SmapMessage{}
	err := decoder.Decode(&rawmessage)
	if err != nil {
		return decodedjson, err
	}

	for path, reading := range rawmessage {

		js, err := simplejson.NewJson([]byte(*reading))
		if err != nil {
			e = err
		}

		// get uuid
		uuid := js.Get("uuid").MustString("")

		message := &archiver.SmapMessage{Path: path, UUID: uuid, Contents: []string{}}

		// get metadata
		localmetadata := js.Get("Metadata").MustMap()
		if localmetadata != nil {
			message.Metadata = bson.M(localmetadata)
		}

		// get contents
		contents := js.Get("Contents").MustArray()
		if len(contents) > 0 {
			for _, arg := range contents {
				message.Contents = append(message.Contents, arg.(string))
			}
		}

		// get properties
		properties := js.Get("Properties").MustMap()
		if properties != nil {
			message.Properties = bson.M(properties)
		}

		// get readings
		readingarray := js.Get("Readings").MustArray()
		srs := make([][]interface{}, len(readingarray))
		for idx, readings := range readingarray {
			reading := readings.([]interface{})
			ts_num, ok := reading[0].(json.Number)
			if !ok {
				return decodedjson, errors.New("Timestamp is not a number")
			}
			ts, e := strconv.ParseUint(string(ts_num), 10, 64)
			if e != nil {
				return decodedjson, e
			}
			// if reading[1] is a number, parse it, else keep it as its default type
			if val_num, ok := reading[1].(json.Number); ok {
				val, e := strconv.ParseFloat(string(val_num), 64)
				if e != nil {
					return decodedjson, e
				}
				srs[idx] = []interface{}{ts, val}
			} else {
				srs[idx] = []interface{}{ts, reading[1]}
			}
		}
		message.Readings = srs

		// get actuator
		actuator := js.Get("Actuator").MustMap()
		if actuator != nil {
			message.Actuator = bson.M(actuator)
		}
		decodedjson[path] = message

	}
	return decodedjson, e
}
