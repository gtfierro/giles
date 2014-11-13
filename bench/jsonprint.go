package main

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

func AddReadingHandler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(200)
	log.Println(string(data))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/add", AddReadingHandler).Methods("POST")
	r.HandleFunc("/add/{key}", AddReadingHandler).Methods("POST")
	http.Handle("/", r)

	srv := &http.Server{
		Addr: "0.0.0.0:8079",
	}
	log.Panic(srv.ListenAndServe())
}
