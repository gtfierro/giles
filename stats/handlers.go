// Package stats serves statistics about the system and Giles.
package stats

import (
	"fmt"
	"github.com/gtfierro/giles/archiver"
	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"html/template"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
)

var log = logging.MustGetLogger("stats")

func Handle(a *archiver.Archiver, port int) {
	go collectSystemStats()

	r := httprouter.New()
	r.GET("/", serveTemplate)
	r.GET("/api/systemstats", systemStatsHandler)
	r.GET("/api/gilesstats", curryhandler(a, gilesStatsHandler))
	r.ServeFiles("/static/*filepath", http.Dir("stats/static"))

	address, err :=
		net.ResolveTCPAddr("tcp4", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("Error resolving address %v: %v",
			"0.0.0.0:"+strconv.Itoa(port),
			err)
	}

	log.Notice("Starting Stats on %v", address.String())
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), r))
}

func serveTemplate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := path.Join("stats/templates", "layout.html")
	fp := path.Join("stats/templates", r.URL.Path)

	fmt.Printf("%v, %v\n", lp, fp)

	// Return a 404 if template doesn't exist
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Does not exist.\n")
			http.NotFound(w, r)
			return
		}
	}

	if info.IsDir() {
		fp += "/index.html"
	}

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		log.Info(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
		log.Info(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func curryhandler(a *archiver.Archiver, f func(*archiver.Archiver, http.ResponseWriter, *http.Request, httprouter.Params)) func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	return func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		f(a, rw, req, ps)
	}
}
