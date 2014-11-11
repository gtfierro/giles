package giles

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"net/http"
	"os"
)

var log = logging.MustGetLogger("archiver")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

//TODO: probably name this one 'archiver' and rename 'archiver.go' to 'giles.go'

type Archiver struct {
	address              string
	tsdb                 TSDB
	store                *Store
	republisher          *Republisher
	incomingcounter      *Counter
	pendingwritescounter *Counter
}

// Creates new archiver
func NewArchiver(tsdb TSDB, store *Store, address string) *Archiver {
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	return &Archiver{tsdb: tsdb,
		store:                store,
		republisher:          NewRepublisher(),
		incomingcounter:      NewCounter(),
		pendingwritescounter: NewCounter()}
}

// Serves HTTP endpoints
func (a *Archiver) ServeHTTP() {
	r := mux.NewRouter()
	r.HandleFunc("/add/{key}", curryhandler(a, AddReadingHandler)).Methods("POST")
	r.HandleFunc("/republish", RepublishHandler).Methods("POST")
	r.HandleFunc("/api/query", QueryHandler).Queries("key", "{key:[A-Za-z0-9-_=%]+}").Methods("POST")
	r.HandleFunc("/api/query", QueryHandler).Methods("POST")
	r.HandleFunc("/api/tags/uuid/{uuid}", TagsHandler).Methods("GET")
	r.HandleFunc("/api/{method}/uuid/{uuid}", DataHandler).Methods("GET")

	//r.HandleFunc("/ws/api/query", WsQueryHandler).Methods("POST")
	//r.HandleFunc("/ws/tags/uuid", WsTagsHandler).Methods("GET")
	//r.HandleFunc("/ws/tags/uuid/{uuid}", WsTagsHandler).Methods("GET")
	http.Handle("/", r)

	srv := &http.Server{
		Addr: a.address,
	}
	srv.ListenAndServe()
}

// Takes a map of string/SmapMessage (path, sMAP JSON object) and commits them to
// the underlying databases. First, checks that write permission is granted with the accompanied
// apikey (generated with the gilescmd CLI tool), then saves the metadata, pushes the readings
// out to any concerned republish clients, and commits the reading to the timeseries database.
// Returns an error, which is nil if all went well
func (a *Archiver) AddData(readings map[string]*SmapMessage, apikey string) error {
	ok, err := a.store.CheckKey(apikey, readings)
	if err != nil {
		log.Info("Error checking API key %v: %v", apikey, err)
		return err
	}
	if !ok {
		return errors.New("Unauthorized api key " + apikey)
	}
	a.store.SavePathMetadata(&readings)
	for _, msg := range readings {
		go a.store.SaveMetadata(msg)
		go a.republisher.Republish(msg)
		a.tsdb.Add(msg.Readings)
		a.incomingcounter.Mark()
	}
	return nil
}

func (a *Archiver) GetData(streamids []string, start, end uint64) ([]SmapResponse, error) {
	return a.tsdb.GetData(streamids, start, end)
}

func (a *Archiver) PrevData(streamids []string, start uint64, limit int32) ([]SmapResponse, error) {
	return a.tsdb.Prev(streamids, start, limit)
}

func (a *Archiver) NextData(streamids []string, start uint64, limit int32) ([]SmapResponse, error) {
	return a.tsdb.Next(streamids, start, limit)
}

func (a *Archiver) GetTags(select_tags, where_tags map[string]interface{}) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (a *Archiver) GetUUIDs(where_tags map[string]interface{}) ([]string, error) {
	return []string{}, nil
}

func (a *Archiver) SetTags(update_tags, where_tags map[string]interface{}) (int, error) {
	return 0, nil
}
