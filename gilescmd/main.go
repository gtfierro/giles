package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"strconv"
)

/* config flags */
var mongoip = flag.String("mongoip", "localhost", "MongoDB IP address")
var mongoport = flag.Int("mongoport", 27017, "MongoDB Port")

/* input flags */
var name = flag.String("name", "", "Name for API key")
var email = flag.String("email", "", "Email of user for API key")

var uuid = flag.String("uuid", "", "UUID to lookup")
var apikey = flag.String("apikey", "", "API key owner of UUID stream")

// make streams private by default
var public = flag.Bool("public", false, "Streams with this API are public?")

/* command flags */
var newapikey = flag.Bool("newkey", false, "Generate a new API key")
var streamidlookup = flag.Bool("streamid", false, "Lookup streamid for UUID")

var mongo *Mongo

type Mongo struct {
	session      *mgo.Session
	db           *mgo.Database
	streams      *mgo.Collection
	metadata     *mgo.Collection
	pathmetadata *mgo.Collection
	apikeys      *mgo.Collection
}

type APIKeyRecord struct {
	Key    string
	Name   string
	Email  string
	Public bool
	UUIDS  map[string]struct{}
}

func NewMongo(ip string, port int) *Mongo {
	address := ip + ":" + strconv.Itoa(port)
	session, err := mgo.Dial(address)
	if err != nil {
		log.Panic(err)
		return nil
	}
	//session.SetMode(mgo.Eventual, true)
	db := session.DB("archiver")
	apikeys := db.C("apikeys")
	metadata := db.C("metadata")
	streams := db.C("streams")
	return &Mongo{session: session, db: db, apikeys: apikeys, metadata: metadata, streams: streams}
}

func (m *Mongo) NewAPIKey(name, email string, public bool) string {
	if name == "" || email == "" {
		fmt.Println("-newkey requires that -name and -email cannot be blank")
		flag.Usage()
		return ""
	}
	f, err := os.Open("/dev/urandom")
	if err != nil {
		fmt.Print("whoops no /dev/urandom")
		return ""
	}
	buf := make([]byte, 64)
	_, err = f.Read(buf)
	key := base64.URLEncoding.EncodeToString(buf)
	record := APIKeyRecord{Key: key, Name: name, Email: email, Public: public, UUIDS: map[string]struct{}{}}
	err = m.apikeys.Insert(record)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return key
}

func (m *Mongo) LookupStreamid(apikey, uuid string) uint32 {
	var q *mgo.Query
	var answer bson.M
	var streamid = uint32(0)
	if apikey == "" || uuid == "" {
		fmt.Println("-streamid requires both -apikey and -uuid")
		flag.Usage()
		return streamid
	}
	q = m.metadata.Find(bson.M{"uuid": uuid, "_api": apikey})
	count, err := q.Count()
	if err != nil {
		fmt.Println("Error talking to mongo", err)
		return streamid
	}
	if count == 0 {
		fmt.Println("No UUID found for that API key")
		return streamid
	}
	// now we know that the uuid is owned by this apikey, so we can safely return the answer
	q = m.streams.Find(bson.M{"uuid": uuid}).Select(bson.M{"streamid": 1})
	err = q.One(&answer)
	if err != nil {
		fmt.Println("Error extracting streamid", err)
		return streamid
	}
	return uint32(answer["streamid"].(int))
}

func main() {
	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}
	mongo = NewMongo(*mongoip, *mongoport)
	if *newapikey {
		fmt.Println(mongo.NewAPIKey(*name, *email, *public))
	} else if *streamidlookup {
		fmt.Println(mongo.LookupStreamid(*apikey, *uuid))
	}
}
