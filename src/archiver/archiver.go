package main

import (
	_ "fmt"
	"io/ioutil"
	"log"
	_ "net/http/pprof"

	"encoding/json"
	"net/http"
	"runtime"
)

var rdb *RDB

type SmapReading struct {
	Readings [][]uint64
	UUID     string
}

func processSmapReading(jdata *[]byte) {
	var reading map[string]*json.RawMessage
	err := json.Unmarshal(*jdata, &reading)
	if err != nil {
		log.Panic(err)
		return
	}

	for _, v := range reading {
		var sr SmapReading
		err = json.Unmarshal(*v, &sr)
		if err != nil {
			log.Panic(err)
			return
		}
		//if len(sr.Readings) > 0 {
		//  println(sr.Readings[0][0], sr.Readings[0][1])
		//}
		rdb.Add(&sr)
	}
}

func sMAPAddHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	defer req.Body.Close()
	jdata, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Panic(err)
		rw.WriteHeader(500)
		return
	}
	processSmapReading(&jdata)
	rw.WriteHeader(200)

}

func main() {

	rdb = NewReadingDB("localhost:4242")
	rdb.Connect()
	go rdb.DoWrites()

	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Println("Utilizing", runtime.NumCPU(), "CPUs")

	http.HandleFunc("/add", sMAPAddHandler)

	log.Println("Starting HTTP Server on port 8079...")
	log.Panic(http.ListenAndServe("0.0.0.0:8079", nil))
}
