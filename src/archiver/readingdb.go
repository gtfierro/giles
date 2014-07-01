package main


import (
    _ "code.google.com/p/goprotobuf/proto"
    "code.google.com/p/go-uuid/uuid"
)

func rdb_add(sr *SmapReading) {
    var seqno uint32 = 0
    var timestamp uint32
    var value float64
    streamid, _ := uuid.Parse(sr.UUID).Id()
    readingset := &ReadingSet{Streamid: &streamid, Data: make([](*Reading), len(sr.Readings))}
    for i, reading := range sr.Readings {
      timestamp = uint32(reading[0])
      value = float64(reading[1])
      (*readingset).Data[i] = &Reading{Timestamp: &timestamp, Seqno: &seqno, Value: &value}
    }
    //TODO: put this reading somewhere...
    //println(readingset.Streamid)
    //println((*readingset.Data[0].Timestamp))
    //println((*readingset.Data[0].Seqno))
    //println((*readingset.Data[0].Value))
}
