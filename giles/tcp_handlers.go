package main

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

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
	//TODO: check we have permission to write
	//vars := mux.Vars(req)
	//apikey := vars["key"]
	messages, err := handleJSON(req.Body)
	if err != nil {
		log.Error("Error handling JSON", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	incomingcounter.Mark()
	//ok, err := store.CheckKey(apikey, messages)
	//if err != nil {
	//	log.Println(err)
	//	rw.WriteHeader(500)
	//	rw.Write([]byte(err.Error()))
	//	return
	//}
	//if !ok {
	//	rw.WriteHeader(400)
	//	rw.Write([]byte("Unauthorized api key " + apikey))
	//	return
	//}
	store.SavePathMetadata(&messages)
	for _, msg := range messages {
		go store.SaveMetadata(msg)
		go republisher.Republish(msg)
		tsdb.Add(msg.Readings)
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
		log.Error("Error handling republish", err)
	}
	republisher.HandleSubscriber(rw, string(stringquery))
}

/**
 * Resolves sMAP queries and returns results
**/
func QueryHandler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	vars := mux.Vars(req)
	key := vars["key"]
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error reading query", err)
	}
	res, err := store.Query(stringquery, key)
	if err != nil {
		log.Error("Error evaluating query", err)
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
		log.Error("Error evaluating tags", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
	rw.Write(res)
}
