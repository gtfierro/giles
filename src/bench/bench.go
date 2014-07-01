package main

import (
  "fmt"
  "net/http"
  "runtime"
  "log"

  "sync"
  "io"
  "sync/atomic"

  "bytes"
)

var data = []byte(`{"/sensor0" : {"Metadata" : {"SourceName" : "Test Source","Location" : { "City" : "Berkeley" }},"Properties": {"Timezone": "America/Los_Angeles","UnitofMeasure": "Watt","ReadingType": "double"},"Readings" : [[1351043674000, 0], [1351043675000, 1]],"uuid" : "d24325e6-1d7d-11e2-ad69-a7c2fa8dba61"}}`)
var wg sync.WaitGroup
var bad uint64
var gud uint64

func makePost(s *sync.WaitGroup) {
  defer s.Done()
  var postdata io.Reader
  postdata = bytes.NewBuffer(data)

  _, err := http.Post("http://localhost:8079/add", "application/json", postdata)
  if err != nil {
    atomic.AddUint64(&bad, 1)
    log.Println(err)
  } else {
    atomic.AddUint64(&gud, 1)
  }
  //resp.Body.Close()
}

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())

  num := 100000
  wg.Add(num)
  for x:=0; x<num; x+=1 {
    makePost(&wg)
  }
  wg.Wait()
  fmt.Println("bad", bad)
  fmt.Println("gud", gud)

}
