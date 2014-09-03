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
var store *Store
var UUIDCache = make(map[string]uint32)
var Clients [](*RepublishClient)
var Subscribers = make(map[string][](*RepublishClient))

type RepublishClient struct {
	uuids  []string
	in     chan []byte
	writer http.ResponseWriter
}

func AddReadingHandler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	jdata, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(500)
		return
	}
	messages, err := handleJSON(&jdata)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(500)
		return
	}
	rw.WriteHeader(200)
	for _, msg := range messages {
		go rdb.Add(msg.Readings)
		go store.SaveMetadata(msg)
	}
}

func RepublishHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	rw.Header().Set("Content-Type", "application/json")
	notify := rw.(http.CloseNotifier).CloseNotify()

	client := &RepublishClient{[]string{uuid}, make(chan []byte), rw}
	Clients = append(Clients, client)
	Subscribers[uuid] = append(Subscribers[uuid], client)

	for {
		select {
		case <-notify:
			// remove client from Clients
			for i, pubclient := range Clients {
				if pubclient == client {
					Clients = append(Clients[:i], Clients[i+1:]...)
				}
			}
			// remove client from Subscribers
			for uuid, clientlist := range Subscribers {
				for i, pubclient := range clientlist {
					if pubclient == client {
						clientlist = append(clientlist[:i], clientlist[i+1:]...)
					}
				}
				Subscribers[uuid] = clientlist
			}
			return
		case datum := <-client.in:
			rw.Write(datum)
			rw.Write([]byte{'\n', '\n'})
			if flusher, ok := rw.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}

}

func QueryHandler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	stringquery, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
	}
	res, err := store.Query(stringquery)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
	}
	rw.WriteHeader(200)
	rw.Write(res)
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")
var archiverport = flag.Int("port", 8079, "archiver service port")
var readingdbip = flag.String("rdbip", "localhost", "ReadingDB IP address")
var readingdbport = flag.Int("rdbport", 4242, "ReadingDB Port")
var mongoip = flag.String("mongoip", "localhost", "MongoDB IP address")
var mongoport = flag.Int("mongoport", 27017, "MongoDB Port")

func main() {
	flag.Parse()
	log.Println("Serving on port", *archiverport)
	log.Println("ReadingDB server", *readingdbip)
	log.Println("ReadingDB port", *readingdbport)
	log.Println("Mongo server", *mongoip)
	log.Println("Mongo port", *mongoport)

	/** Configure CPU profiling */
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	/** connect to Metadata store*/
	store = NewStore(*mongoip, *mongoport)
	if store == nil {
		log.Fatal("Error connection to MongoDB instance")
	}

	/** connect to ReadingDB */
	rdb = NewReadingDB(*readingdbip, *readingdbport)
	if rdb == nil {
		log.Fatal("Error connecting to ReadingDB instance")
	}
	rdb.Connect()
	go rdb.DoWrites()

	runtime.GOMAXPROCS(runtime.NumCPU())

	r := mux.NewRouter()
	r.HandleFunc("/add", AddReadingHandler).Methods("POST")
	r.HandleFunc("/add/{key}", AddReadingHandler).Methods("POST")
	r.HandleFunc("/republish/{uuid}", RepublishHandler).Methods("POST")
	r.HandleFunc("/api/query", QueryHandler).Methods("POST")

	http.Handle("/", r)

	srv := &http.Server{
		Addr: "0.0.0.0:" + strconv.Itoa(*archiverport),
	}

	log.Println("Starting HTTP Server on port " + strconv.Itoa(*archiverport) + "...")
	go srv.ListenAndServe()
	idx := 0
	for {
		log.Println("still alive", idx, idx*5/10)
		time.Sleep(5 * time.Second)
		idx++
		if idx*5/10 == 10 {
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
