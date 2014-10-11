package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

var rdb *RDB
var tsdb TSDB
var store *Store
var UUIDCache = make(map[string]uint32)
var republisher *Republisher
var cm *ConnectionMap
var incomingcounter = NewCounter()
var pendingwritescounter = NewCounter()

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
		log.Println(err)
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
		go tsdb.Add(msg.Readings)
		go store.SaveMetadata(msg)
		go republisher.Republish(msg)
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
		fmt.Println(err)
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
		fmt.Println(err)
	}
	res, err := store.Query(stringquery, key)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
	rw.Write(res)
}

// config flags
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")
var archiverport = flag.Int("port", 8079, "archiver service port")
var readingdbip = flag.String("rdbip", "localhost", "ReadingDB IP address")
var readingdbport = flag.Int("rdbport", 4242, "ReadingDB Port")
var mongoip = flag.String("mongoip", "localhost", "MongoDB IP address")
var mongoport = flag.Int("mongoport", 27017, "MongoDB Port")
var tsdbstring = flag.String("tsdb", "readingdb", "Type of timeseries database to use: 'readingdb' or 'quasar'")
var tsdbkeepalive = flag.Int("keepalive", 30, "Number of seconds to keep TSDB connection alive per stream for reads")

func main() {
	flag.Parse()
	log.Println("Serving on port", *archiverport)
	log.Println("ReadingDB server", *readingdbip)
	log.Println("ReadingDB port", *readingdbport)
	log.Println("Mongo server", *mongoip)
	log.Println("Mongo port", *mongoport)
	log.Println("Using TSDB", *tsdbstring)
	log.Println("TSDB Keepalive", *tsdbkeepalive)

	/** Configure CPU profiling */
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		f2, err := os.Create("blockprofile.db")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		runtime.SetBlockProfileRate(1)
		defer runtime.SetBlockProfileRate(0)
		defer pprof.Lookup("block").WriteTo(f2, 1)
		defer pprof.StopCPUProfile()
	}
	republisher = NewRepublisher()

	/** connect to Metadata store*/
	store = NewStore(*mongoip, *mongoport)
	if store == nil {
		log.Fatal("Error connection to MongoDB instance")
	}

	cm = &ConnectionMap{streams: map[string]*Connection{}, keepalive: *tsdbkeepalive}

	switch *tsdbstring {
	case "readingdb":
		/** connect to ReadingDB */
		tsdb = NewReadingDB(*readingdbip, *readingdbport, cm)
		if tsdb == nil {
			log.Fatal("Error connecting to ReadingDB instance")
		}
	case "quasar":
		log.Fatal("quasar")
	default:
		log.Fatal(*tsdbstring, " is not a valid timeseries database")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	r := mux.NewRouter()
	r.HandleFunc("/add", AddReadingHandler).Methods("POST")
	r.HandleFunc("/add/{key}", AddReadingHandler).Methods("POST")
	r.HandleFunc("/republish", RepublishHandler).Methods("POST")
	r.HandleFunc("/api/query", QueryHandler).Queries("key", "{key:[A-Za-z0-9]+}").Methods("POST")
	r.HandleFunc("/api/query", QueryHandler).Methods("POST")
	r.HandleFunc("/api/tags/uuid/{uuid}", TagsHandler).Methods("GET")

	http.Handle("/", r)

	srv := &http.Server{
		Addr: "0.0.0.0:" + strconv.Itoa(*archiverport),
	}

	log.Println("Starting HTTP Server on port " + strconv.Itoa(*archiverport) + "...")
	go srv.ListenAndServe()
	go periodicCall(1*time.Second, status) // status from stats.go
	idx := 0
	for {
		time.Sleep(5 * time.Second)
		idx += 5
		if idx == 60 {
			if *memprofile != "" {
				f, err := os.Create(*memprofile)
				if err != nil {
					log.Panic(err)
				}
				pprof.WriteHeapProfile(f)
				f.Close()
				return
			}
			if *cpuprofile != "" {
				return
			}
		}
	}
	//log.Panic(srv.ListenAndServe())

}
