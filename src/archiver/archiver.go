package main

import (
    "log"
    _ "fmt"
    "io/ioutil"
    _ "net/http/pprof"

    "runtime"
    "encoding/json"
    "net/http"

    //"code.google.com/p/goprotobuf/proto"
)

type SmapReading struct {
    Readings [][]int64
    UUID string
}

func processSmapReading(jdata *[]byte) {
    var reading map[string]*json.RawMessage
    err := json.Unmarshal(*jdata, &reading) 
    //println(jdata)
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
      //println(sr.Readings)
      //println(sr.UUID)

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
    //reading := &Reading {
    //  Timestamp: proto.Uint32(1),
    //  Value: proto.Float64(2),
    //}

    //data, err := proto.Marshal(reading)
    //if err != nil {
    //    log.Fatal("marshaling error: ", err)
    //}

    //newTest := &Reading{}
    //err = proto.Unmarshal(data, newTest)
    //if err != nil {
    //    log.Fatal("unmarshaling error: ", err)
    //}
    //// Now test and newTest contain the same data.
    //if reading.GetValue() != newTest.GetValue() {
    //    log.Fatalf("data mismatch %q != %q", reading.GetValue(), newTest.GetValue())
    //}

    runtime.GOMAXPROCS(runtime.NumCPU())
    log.Println("Utilizing",runtime.NumCPU(),"CPUs")

    http.HandleFunc("/add", sMAPAddHandler)

    log.Println("Starting HTTP Server on port 8079...")
    log.Panic(http.ListenAndServe(":8079", nil))
}
