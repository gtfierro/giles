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

/**
 * Handles POSTing of new data
**/
func AddReadingHandler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	messages, err := handleJSON(req.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
	for _, msg := range messages {
		go tsdb.Add(msg.Readings)
		go store.SaveMetadata(msg)
		go republisher.Republish(msg)
	}
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
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
	}
	res, err := store.Query(stringquery)
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

/**
 * Prints status of the archiver:
 ** number of connected clients
 ** size of UUID cache
 ** connection status to database
 ** connection status to Mongo
 ** amount of incoming traffic since last call
 ** amount of api requests since last call
**/
func status() {
	log.Print("Still alive at: ", time.Now())
	log.Print("UUID Cache size: ", len(UUIDCache))
	log.Print("Connected republish clients: ", len(republisher.Clients))
	cm.Stats()
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
		log.Println("quasar")
	default:
		log.Fatal(*tsdbstring, " is not a valid timeseries database")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	r := mux.NewRouter()
	r.HandleFunc("/add", AddReadingHandler).Methods("POST")
	r.HandleFunc("/add/{key}", AddReadingHandler).Methods("POST")
	r.HandleFunc("/republish", RepublishHandler).Methods("POST")
	r.HandleFunc("/api/query", QueryHandler).Methods("POST")
	r.HandleFunc("/api/tags/uuid/{uuid}", TagsHandler).Methods("GET")

	http.Handle("/", r)

	srv := &http.Server{
		Addr: "0.0.0.0:" + strconv.Itoa(*archiverport),
	}

	log.Println("Starting HTTP Server on port " + strconv.Itoa(*archiverport) + "...")
	go srv.ListenAndServe()
	go periodicCall(5*time.Second, status)
	idx := 0
	for {
		log.Println("still alive", idx)
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
