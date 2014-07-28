package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)

var rdb *RDB
var Clients [](*RepublishClient)
var Subscribers = make(map[string][](*RepublishClient))

type RepublishClient struct {
	uuids  []string
	in     chan []byte
	writer http.ResponseWriter
}

type SmapReading struct {
	Readings [][]uint64
	UUID     string
}

func processJSON(bytes *[]byte) [](*SmapReading) {
	var reading map[string]*json.RawMessage
	var dest [](*SmapReading)
	err := json.Unmarshal(*bytes, &reading)
	if err != nil {
		log.Panic(err)
		return nil
	}

	for _, v := range reading {
		if v == nil {
			continue
		}
		var sr SmapReading
		err = json.Unmarshal(*v, &sr)
		if err != nil {
			log.Panic(err)
			return nil
		}
		dest = append(dest, &sr)
	}
	return dest
}

func AddReadingHandler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	jdata, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Panic(err)
		rw.WriteHeader(500)
		return
	}
	readings := processJSON(&jdata)
	rw.WriteHeader(200)

	// rdb.Add(smap reading)
	for _, reading := range readings {
		rdb.Add(reading)
		go func(reading *SmapReading) {
			for _, client := range Subscribers[reading.UUID] {
				go func(client *RepublishClient) {
					bytes, err := json.Marshal(reading)
					if err != nil {
						log.Panic(err)
					}
					client.in <- bytes
				}(client)
			}
		}(reading)
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

func main() {

	rdb = NewReadingDB("localhost:4242")
	rdb.Connect()
	go rdb.DoWrites()

	runtime.GOMAXPROCS(runtime.NumCPU())

	r := mux.NewRouter()
	r.HandleFunc("/add", AddReadingHandler).Methods("POST")
	r.HandleFunc("/add/{key}", AddReadingHandler).Methods("POST")
	r.HandleFunc("/republish/{uuid}", RepublishHandler).Methods("POST")

	http.Handle("/", r)

	srv := &http.Server{
		Addr: "0.0.0.0:8079",
	}

	log.Println("Starting HTTP Server on port 8079...")
	log.Panic(srv.ListenAndServe())

}
