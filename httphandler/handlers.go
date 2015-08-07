// License stuff

// Package httphandler implements an HTTP interface to the Archiver API at
// http://godoc.org/github.com/gtfierro/giles/archiver
//
// Overview
//
// This HTTP interface to the sMAP archiver presents most of the expected
// interface as described in the Archiver documentation at
// http://pythonhosted.org/Smap/en/2.0/archiver.html. With the exception of
// DELETE action and the /api/query interface, this interface matches what is
// expected.
//
// For sample usage, look at https://github.com/gtfierro/giles/blob/master/giles/giles.go
package httphandler

import (
	"bytes"
	"encoding/json"
	"github.com/gtfierro/giles/archiver"
	"github.com/gtfierro/msgpack"
	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var log = logging.MustGetLogger("httphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

func Handle(a *archiver.Archiver, port int) {
	r := httprouter.New()
	r.POST("/add/:key", curryhandler(a, AddReadingHandler))
	r.POST("/api/query", curryhandler(a, QueryHandler))
	r.POST("/api/test", curryhandler(a, Query2Handler))
	r.GET("/api/tags/uuid/:uuid", curryhandler(a, TagsHandler))

	r.POST("/api/streamingquery", curryhandler(a, StreamingQueryHandler))

	r.POST("/republish2", curryhandler(a, RepublishHandler2))
	r.POST("/republish", curryhandler(a, RepublishHandler))
	r.POST("/republish/data", curryhandler(a, RepublishHandler))
	r.POST("/republish/uuids", curryhandler(a, UUIDRepublishHandler))
	r.POST("/republish/query", curryhandler(a, QueryRepublishHandler))

	address, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("Error resolving address %v: %v", "0.0.0.0:"+strconv.Itoa(port), err)
	}
	http.Handle("/", r)
	log.Notice("Starting HTTP on %v", address.String())

	srv := &http.Server{
		Addr: address.String(),
	}
	srv.ListenAndServe()
}

func curryhandler(a *archiver.Archiver, f func(*archiver.Archiver, http.ResponseWriter, *http.Request, httprouter.Params)) func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	return func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		f(a, rw, req, ps)
	}
}

// Handles POSTing of new data
// The HandleJSON method parses the message received from the sMAP drivers
// and delivers them as an array. Because metadata is delivered as k/v pairs
// representing a tree, we have a pre-loop that stores the metadata values at
// the higher levels of the tree. Then, when we loop through the data to add it
// to the leaves of the tree (the actual timeseries), we query the prefixes
// of the timeseries path to get all the 'trickle down' metadata from the higher
// parts of the metadata tree. That logic takes place in store.SavePathMetadata and
// store.SaveMetadata
//
// An example of a valid sMAP object is
//    {
//      "/sensor0" : {
//        "Metadata" : {
//          "SourceName" : "Test Source",
//            "Location" : { "City" : "Berkeley" }
//        },
//          "Properties": {
//            "Timezone": "America/Los_Angeles",
//            "UnitofMeasure": "Watt",
//            "ReadingType": "double"
//          },
//          "Readings" : [[1351043674000, 0], [1351043675000, 1]],
//          "uuid" : "d24325e6-1d7d-11e2-ad69-a7c2fa8dba61"
//      }
//    }
func AddReadingHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	apikey := ps.ByName("key")
	messages, err := HandleJSON(req.Body)
	if err != nil {
		log.Error("Error handling JSON: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	err = a.AddData(messages, apikey)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
}

// Receives POST request which contains metadata query. Subscribes the
// requester to readings from streams which match that metadata query
func RepublishHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	apikey := unescape(ps.ByName("key"))
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error handling republish: %v", err, stringquery)
	}
	s := NewHTTPSubscriber(rw)
	a.HandleSubscriber(s, string(stringquery), apikey)
}

// Receives POST request which contains metadata query. Subscribes the
// requester to readings from streams which match that metadata query
func RepublishHandler2(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	apikey := unescape(ps.ByName("key"))
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error handling republish: %v", err, stringquery)
	}
	s := NewHTTPSubscriber(rw)
	a.HandleSubscriber2(s, string(stringquery), apikey)
}

func MetadataRepublishHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	apikey := unescape(ps.ByName("key"))
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error handling republish: %v", err, stringquery)
	}
	s := NewHTTPSubscriber(rw)
	a.HandleMetadataSubscriber(s, string(stringquery), apikey)
}

func UUIDRepublishHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	var uuids []string
	defer req.Body.Close()
	decoder := json.NewDecoder(req.Body)
	decodeErr := decoder.Decode(&uuids)
	if decodeErr != nil {
		log.Error("Error decoding list of UUIDs (%v)", decodeErr)
		return
	}
	apikey := unescape(ps.ByName("key"))
	s := NewHTTPSubscriber(rw)
	a.HandleUUIDSubscriber(s, uuids, apikey)
}

// Receives POST request which contains metadata query. Subscribes the
// requester to readings from streams which match that metadata query
func QueryRepublishHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	apikey := unescape(ps.ByName("key"))
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error handling republish: %v", err, stringquery)
	}
	s := NewHTTPSubscriber(rw)
	a.HandleQuerySubscriber(s, string(stringquery), apikey)
}

// Resolves sMAP queries and returns results
func QueryHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	writer := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	key := unescape(ps.ByName("key"))
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error reading query: %v", err)
	}
	res, err := a.HandleQuery(string(stringquery), key)
	if err != nil {
		log.Error("Error evaluating query: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	switch res.(type) {
	case []archiver.SmapNumbersResponse:
		err = writer.Encode(res.([]archiver.SmapNumbersResponse))
	default:
		err = writer.Encode(res)
	}
	if err != nil {
		log.Error("Error evaluating query: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
}

// Resolves sMAP queries and returns results
func Query2Handler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	key := unescape(ps.ByName("key"))
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error reading query: %v", err)
	}
	var b bytes.Buffer
	err = a.Query2(string(stringquery), key, &b)
	if err != nil {
		log.Error("Error evaluating query: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	encodedbytes := b.Bytes()
	_, decoded := msgpack.Decode(&encodedbytes, 0)
	res, err := json.Marshal(decoded)
	if err != nil {
		log.Error("Error converting to json: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
	rw.Write(res)
}

func StreamingQueryHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	key := unescape(ps.ByName("key"))
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error reading query: %v", err)
	}
	s := NewHTTPSubscriber(rw)
	err = a.StreamingQuery(string(stringquery), key, s)
	if err != nil {
		log.Error("Error evaluating query: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
}

/**
 * Returns metadata for a uuid. A limited GET alternative to the POST query handler
**/
func TagsHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	uuid := ps.ByName("uuid")
	rw.Header().Set("Content-Type", "application/json")
	jsonres, err := a.TagsUUID(uuid)
	if err != nil {
		log.Error("Error evaluating tags: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	res, err := json.Marshal(jsonres)
	if err != nil {
		log.Error("Error converting to json: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
	rw.Write(res)
}

func unescape(s string) string {
	return strings.Replace(s, "%3D", "=", -1)
}
