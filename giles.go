package main

import (
	"flag"
	"github.com/gtfierro/giles/giles"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

// config flags
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")
var archiverport = flag.Int("port", 8079, "archiver service port")
var readingdbip = flag.String("rdbip", "localhost", "ReadingDB IP address")
var readingdbport = flag.Int("rdbport", 4242, "ReadingDB Port")
var mongoip = flag.String("mongoip", "localhost", "MongoDB IP address")
var mongoport = flag.Int("mongoport", 27017, "MongoDB Port")
var tsdbstring = flag.String("tsdb", "readingdb", "Type of timeseries database to use: 'readingdb' or 'quasar'")
var tsdbkeepalive = flag.Int("keepalive", 30, "Number of seconds to keep TSDB connection alive per stream for reads")
var benchmarktimer = flag.Int("benchmark", 60, "Number of seconds to benchmark before quitting and writing profiles")

func main() {
	flag.Parse()
	log.Println("Serving on port %v", *archiverport)
	log.Println("ReadingDB server %v", *readingdbip)
	log.Println("ReadingDB port %v", *readingdbport)
	log.Println("Mongo server %v", *mongoip)
	log.Println("Mongo port %v", *mongoport)
	log.Println("Using TSDB %v", *tsdbstring)
	log.Println("TSDB Keepalive %v", *tsdbkeepalive)

	/** Configure CPU profiling */
	if *cpuprofile != "" {
		log.Println("Benchmarking for %v seconds", *benchmarktimer)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		f2, err := os.Create("blockprofile.db")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		runtime.SetBlockProfileRate(1)
		defer runtime.SetBlockProfileRate(0)
		defer pprof.Lookup("block").WriteTo(f2, 1)
		defer pprof.StopCPUProfile()
	}

	/** connect to Metadata store*/
	store := giles.NewStore(*mongoip, *mongoport)
	if store == nil {
		log.Fatal("Error connection to MongoDB instance")
	}

	var tsdb giles.TSDB
	switch *tsdbstring {
	case "readingdb":
		/** connect to ReadingDB */
		tsdb = giles.NewReadingDB(*readingdbip, *readingdbport, *tsdbkeepalive)
		if tsdb == nil {
			log.Fatal("Error connecting to ReadingDB instance")
		}
	case "quasar":
		log.Fatal("quasar")
	default:
		log.Fatal(*tsdbstring, " is not a valid timeseries database")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	a := giles.NewArchiver(tsdb, store, "0.0.0.0:"+strconv.Itoa(*archiverport))
	go a.ServeHTTP()

	//go periodicCall(1*time.Second, status) // status from stats.go
	log.Println("...connected!")
	idx := 0
	for {
		time.Sleep(5 * time.Second)
		idx += 5
		if idx == *benchmarktimer {
			if *memprofile != "" {
				f, err := os.Create(*memprofile)
				if err != nil {
					log.Panic(err)
				}
				pprof.WriteHeapProfile(f)
				f.Close()
				return
			}
			if *cpuprofile != "" {
				return
			}
		}
	}
	//log.Panic(srv.ListenAndServe())

}
