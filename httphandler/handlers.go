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
	"encoding/json"
	"github.com/gtfierro/giles/archiver"
	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"os"
	_ "strconv"
	"strings"
	_ "time"
)

var log = logging.MustGetLogger("httphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

func Handle(a *archiver.Archiver) {
	log.Notice("Handling HTTP/TCP")
	a.R.POST("/add/:key", curryhandler(a, AddReadingHandler))
	a.R.POST("/republish", curryhandler(a, RepublishHandler))
	//a.R.POST("/api/query?:key", curryhandler(a, QueryHandler))
	a.R.POST("/api/query", curryhandler(a, QueryHandler))
	a.R.GET("/api/tags/uuid/:uuid", curryhandler(a, TagsHandler))
	//a.R.HandleFunc("/api/{method}/uuid/{uuid}", curryhandler(a, DataHandler)).Methods("GET")
}

func curryhandler(a *archiver.Archiver, f func(*archiver.Archiver, http.ResponseWriter, *http.Request, httprouter.Params)) func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	return func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		f(a, rw, req, ps)
	}
}

// Handles POSTing of new data
// The handleJSON method parses the message received from the sMAP drivers
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
	//TODO: add transaction coalescing
	defer req.Body.Close()
	apikey := ps.ByName("key")
	messages, err := handleJSON(req.Body)
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
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error handling republish: %v", err, stringquery)
	}
	s := NewHTTPSubscriber(rw)
	a.HandleSubscriber(s, string(stringquery))
}

// Resolves sMAP queries and returns results
func QueryHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
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
	rw.WriteHeader(200)
	rw.Write(res)
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

//func DataHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request) {
//	var starttime, endtime uint64
//	var limit int64
//	var startstr, endstr, timeunitstr, limitstr []string
//	var querytimeunit string
//	var response []SmapResponse
//	var err error
//	var found bool
//
//	unitmultiplier := map[string]uint64{"ns": 1000000000, "us": 1000000, "ms": 1000, "s": 1}
//
//	// extract URL query parameters into the req.Form map
//	req.ParseForm()
//	vars := mux.Vars(req)
//	uuid := vars["uuid"]
//	method := vars["method"]
//
//	streamtimeunit := a.store.GetUnitofTime(uuid)
//	// get the unit of time for the query
//	if timeunitstr, found = req.Form["unit"]; !found {
//		querytimeunit = "ms"
//	} else {
//		querytimeunit = timeunitstr[0]
//	}
//
//	// get the limit on the time series
//	if limitstr, found = req.Form["limit"]; !found {
//		limit = -1
//	} else {
//		limit, _ = strconv.ParseInt(limitstr[0], 10, 32)
//	}
//
//	// parse out start and end times, or default to
//	if startstr, found = req.Form["starttime"]; found {
//		starttime, _ = strconv.ParseUint(startstr[0], 10, 64)
//		starttime /= unitmultiplier[querytimeunit]
//	} else {
//		starttime = uint64(time.Now().Unix()) - 3600*24
//	}
//	starttime *= unitmultiplier[streamtimeunit]
//
//	if endstr, found = req.Form["endtime"]; found {
//		endtime, _ = strconv.ParseUint(endstr[0], 10, 64)
//		endtime /= unitmultiplier[querytimeunit]
//	} else {
//		endtime = uint64(time.Now().Unix())
//	}
//	endtime *= unitmultiplier[streamtimeunit]
//
//	rw.Header().Set("Content-Type", "application/json")
//	log.Debug("method: %v, limit %v, start: %v, end: %v", method, limit, starttime, endtime)
//	switch method {
//	case "data":
//		response, err = a.GetData([]string{uuid}, starttime, endtime)
//	case "prev":
//		response, err = a.PrevData([]string{uuid}, starttime, int32(limit))
//	case "next":
//		response, err = a.NextData([]string{uuid}, starttime, int32(limit))
//	}
//	if err != nil {
//		log.Error("Error fetching data: %v", err)
//		rw.WriteHeader(500)
//		rw.Write([]byte(err.Error()))
//		return
//	}
//	log.Debug("response %v", response)
//	res, err := json.Marshal(response)
//	if err != nil {
//		log.Error("Error fetching data: %v", err)
//		rw.WriteHeader(500)
//		rw.Write([]byte(err.Error()))
//		return
//	}
//	rw.WriteHeader(200)
//	rw.Write(res)
//}

func unescape(s string) string {
	return strings.Replace(s, "%3D", "=", -1)
}
