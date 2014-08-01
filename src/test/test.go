package main

import (
	proto "code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
)

type Header struct {
	Type   MessageType
	Length uint32
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

func main() {
	// connect to readingdb
	address := "192.168.59.103:4242"
	//address := "localhost:4242"
	tcpaddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Panic("Error resolving TCP address", address, err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpaddr)

	var timestamp uint64 = 1
	var value float64 = 1
	var streamid uint32 = 1
	var seqno uint64 = 1
	// FOR SOME REASON, THIS MUST BE SUBSTREAM-1??
	var substream uint32 = 0

	reading := &Reading{Timestamp: &timestamp, Seqno: &seqno, Value: &value}
	readingset := &ReadingSet{Streamid: &streamid,
		Substream: &substream,
		Data:      make([](*Reading), 1)}
	(*readingset).Data[0] = reading

	data, err := proto.Marshal(readingset)
	if err != nil {
		log.Panic("Error marshaling ReadingSet", err)
	}

	test := &ReadingSet{}
	err = proto.Unmarshal(data, test)
	println("streamid, substream", test.GetStreamid(), test.GetSubstream())
	for _, d := range test.GetData() {
		println("timestamp, value, seqno", d.GetTimestamp(), d.GetValue(), d.GetSeqno())
	}

	h := &Header{Type: MessageType_READINGSET, Length: uint32(len(data))}

	onthewire := make([]byte, 8)
	binary.BigEndian.PutUint32(onthewire, uint32(h.Type))
	binary.BigEndian.PutUint32(onthewire[4:8], h.Length)
	println(onthewire)

	onthewire = append(onthewire, data...)

	n, err := conn.Write(onthewire)
	if err != nil {
		log.Println("Error writing data to ReadingDB", err, len(data), n)
	}
	println("written", n)

}
