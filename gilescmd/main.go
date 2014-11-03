package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
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

// make streams private by default
var public = flag.Bool("public", false, "Streams with this API are public?")

/* command flags */
var newapikey = flag.Bool("newkey", false, "Generate a new API key")

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
	return &Mongo{session: session, db: db, apikeys: apikeys}
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

func main() {
	flag.Parse()
	mongo = NewMongo(*mongoip, *mongoport)
	if *newapikey {
		fmt.Println(mongo.NewAPIKey(*name, *email, *public))
	}
}
