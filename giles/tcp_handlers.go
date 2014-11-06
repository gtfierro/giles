package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func unescape(s string) string {
	return strings.Replace(s, "%3D", "=", -1)
}

/**
 * Handles POSTing of new data
 * The handleJSON method parses the message received from the sMAP drivers
 * and delivers them as an array. Because metadata is delivered as k/v pairs
 * representing a tree, we have a pre-loop that stores the metadata values at
 * the higher levels of the tree. Then, when we loop through the data to add it
 * to the leaves of the tree (the actual timeseries), we query the prefixes
 * of the timeseries path to get all the 'trickle down' metadata from the higher
 * parts of the metadata tree. That logic takes place in store.SavePathMetadata and
 * store.SaveMetadata
**/
func AddReadingHandler(rw http.ResponseWriter, req *http.Request) {
	//TODO: add transaction coalescing
	defer req.Body.Close()
	vars := mux.Vars(req)
	apikey := unescape(vars["key"])
	messages, err := handleJSON(req.Body)
	if err != nil {
		log.Error("Error handling JSON: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	ok, err := store.CheckKey(apikey, messages)
	if err != nil {
		log.Info("Error checking API key %v: %v", apikey, err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	if !ok {
		rw.WriteHeader(400)
		rw.Write([]byte("Unauthorized api key " + apikey))
		return
	}
	store.SavePathMetadata(&messages)
	for _, msg := range messages {
		go store.SaveMetadata(msg)
		go republisher.Republish(msg)
		tsdb.Add(msg.Readings)
		incomingcounter.Mark()
	}
	rw.WriteHeader(200)
}

/**
 * Receives POST request which contains metadata query. Subscribes the
 * requester to readings from streams which match that metadata query
**/
func RepublishHandler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error handling republish: %v", err)
	}
	republisher.HandleSubscriber(rw, string(stringquery))
}

/**
 * Resolves sMAP queries and returns results
**/
func QueryHandler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	vars := mux.Vars(req)
	key := unescape(vars["key"])
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error reading query: %v", err)
	}
	res, err := store.Query(stringquery, key)
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
func TagsHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	rw.Header().Set("Content-Type", "application/json")
	res, err := store.TagsUUID(uuid)
	if err != nil {
		log.Error("Error evaluating tags: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
	rw.Write(res)
}

//TODO: limit should not be unsigned
func DataHandler(rw http.ResponseWriter, req *http.Request) {
	var starttime, endtime uint64
	var limit int64
	var startstr, endstr, timeunitstr, limitstr []string
	var querytimeunit string
	var response []SmapResponse
	var err error
	var found bool

	unitmultiplier := map[string]uint64{"ns": 1000000000, "us": 1000000, "ms": 1000, "s": 1}

	// extract URL query parameters into the req.Form map
	req.ParseForm()
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	method := vars["method"]

	streamtimeunit := store.GetUnitofTime(uuid)
	// get the unit of time for the query
	if timeunitstr, found = req.Form["unit"]; !found {
		querytimeunit = "ms"
	} else {
		querytimeunit = timeunitstr[0]
	}

	// get the limit on the time series
	if limitstr, found = req.Form["limit"]; !found {
		limit = -1
	} else {
		limit, _ = strconv.ParseInt(limitstr[0], 10, 32)
	}

	// parse out start and end times, or default to
	if startstr, found = req.Form["starttime"]; found {
		starttime, _ = strconv.ParseUint(startstr[0], 10, 64)
		starttime /= unitmultiplier[querytimeunit]
	} else {
		starttime = uint64(time.Now().Unix()) - 3600*24
	}
	starttime *= unitmultiplier[streamtimeunit]

	if endstr, found = req.Form["endtime"]; found {
		endtime, _ = strconv.ParseUint(endstr[0], 10, 64)
		endtime /= unitmultiplier[querytimeunit]
	} else {
		endtime = uint64(time.Now().Unix())
	}
	endtime *= unitmultiplier[streamtimeunit]

	rw.Header().Set("Content-Type", "application/json")
	log.Debug("method: %v, limit %v, start: %v, end: %v", method, limit, starttime, endtime)
	switch method {
	case "data":
		response, err = tsdb.GetData([]string{uuid}, starttime, endtime)
	case "prev":
		response, err = tsdb.Prev([]string{uuid}, starttime, uint32(limit))
	case "next":
		response, err = tsdb.Next([]string{uuid}, starttime, uint32(limit))
	}
	if err != nil {
		log.Error("Error fetching data: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	log.Debug("response %v", response)
	res, err := json.Marshal(response)
	if err != nil {
		log.Error("Error fetching data: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
	rw.Write(res)
}
