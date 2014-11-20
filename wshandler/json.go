package wshandler

import (
	"encoding/json"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/gtfierro/giles/giles"
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"strconv"
)

var log = logging.MustGetLogger("httphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

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
func handleJSON(r io.Reader) (map[string]*giles.SmapMessage, error) {
	/*
	 * we receive a bunch of top-level keys that we don't know, so we unmarshal them into a
	 * map, and then parse each of the internal objects individually
	 */

	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	var e error
	var rawmessage map[string]*json.RawMessage
	var decodedjson = map[string]*giles.SmapMessage{}
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

		message := &giles.SmapMessage{Path: path, UUID: uuid, Contents: []string{}}

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
		sr := &giles.SmapReading{UUID: uuid}
		srs := make([][]interface{}, len(readingarray))
		for idx, readings := range readingarray {
			reading := readings.([]interface{})
			ts, e := strconv.ParseUint(string(reading[0].(json.Number)), 10, 64)
			if e != nil {
				return decodedjson, e
			}
			val, e := strconv.ParseFloat(string(reading[1].(json.Number)), 64)
			if e != nil {
				return decodedjson, e
			}
			srs[idx] = []interface{}{ts, val}
		}
		sr.Readings = srs
		message.Readings = sr

		// get actuator
		actuator := js.Get("Actuator").MustMap()
		if actuator != nil {
			message.Actuator = bson.M(actuator)
		}
		decodedjson[path] = message

	}
	return decodedjson, e
}
