package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

type Counter struct {
	Count uint64
}

func NewCounter() *Counter {
	return &Counter{Count: 0}
}

func (c *Counter) Mark() {
	atomic.AddUint64(&c.Count, 1)
}

func (c *Counter) Reset() uint64 {
	var returncount = c.Count
	atomic.StoreUint64(&c.Count, 0)
	return returncount
}

var incomingcounter = NewCounter()

func HandlePost(rw http.ResponseWriter, req *http.Request) {
	incomingcounter.Mark()
	rw.WriteHeader(200)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	r := mux.NewRouter()
	r.HandleFunc("/add", HandlePost).Methods("POST")
	r.HandleFunc("/add/{key}", HandlePost).Methods("POST")
	http.Handle("/", r)

	srv := &http.Server{
		Addr: "0.0.0.0:8079",
	}

	go func() {
		for {
			log.Printf("Recv Adds:%d", incomingcounter.Reset())
			time.Sleep(1 * time.Second)
		}
	}()

	log.Println("hey")
	log.Panic(srv.ListenAndServe())
}
