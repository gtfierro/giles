package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	_ "gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"strconv"
)

/* config flags */
var mongoip = flag.String("mongoip", "localhost", "MongoDB IP address")
var mongoport = flag.Int("mongoport", 27017, "MongoDB Port")

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

func (m *Mongo) NewAPIKey() string {
	f, err := os.Open("/dev/random")
	if err != nil {
		log.Fatal("whoops no random")
		return ""
	}
	buf := make([]byte, 128)
	_, err = f.Read(buf)
	key := base64.URLEncoding.EncodeToString(buf)
	return key
}

func main() {
	flag.Parse()
	mongo = NewMongo(*mongoip, *mongoport)
	if *newapikey {
		fmt.Println(mongo.NewAPIKey())
	}
}
