package main

import (
	"flag"
	"github.com/gtfierro/giles/archiver"
	"github.com/gtfierro/giles/cphandler"
	"github.com/gtfierro/giles/httphandler"
	"github.com/gtfierro/giles/mphandler"
	"github.com/gtfierro/giles/wshandler"
	"log"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// config flags
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")
var archiverport = flag.Int("port", 8079, "archiver service port")
var readingdbip = flag.String("rdbip", "localhost", "ReadingDB IP address")
var readingdbport = flag.String("rdbport", "4242", "ReadingDB Port")
var mongoip = flag.String("mongoip", "localhost", "MongoDB IP address")
var mongoport = flag.String("mongoport", "27017", "MongoDB Port")
var tsdbstring = flag.String("tsdb", "readingdb", "Type of timeseries database to use: 'readingdb' or 'quasar'")
var tsdbkeepalive = flag.Int("keepalive", 30, "Number of seconds to keep TSDB connection alive per stream for reads")
var benchmarktimer = flag.Int("benchmark", 60, "Number of seconds to benchmark before quitting and writing profiles")
var configfile = flag.String("c", "giles.conf", "Path to Giles configuration file")

func main() {
	flag.Parse()
	config := archiver.LoadConfig(*configfile)
	archiver.PrintConfig(config)

	log.Println("Serving on port %v", *archiverport)
	log.Println("ReadingDB server %v", *readingdbip)
	log.Println("ReadingDB port %v", *readingdbport)
	log.Println("Mongo server %v", *mongoip)
	log.Println("Mongo port %v", *mongoport)
	log.Println("Using TSDB %v", *tsdbstring)
	log.Println("TSDB Keepalive %v", *tsdbkeepalive)

	/** Configure CPU profiling */
	if config.Profile.Enabled {
		log.Println("Benchmarking for %v seconds", *config.Profile.BenchmarkTimer)
		f, err := os.Create(*config.Profile.CpuProfile)
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

	runtime.GOMAXPROCS(runtime.NumCPU())
	a := archiver.NewArchiver(config)
	go a.PrintStatus()
	httphandler.Handle(a)
	wshandler.Handle(a)
	cphandler.Handle(a)
	mphandler.Handle(a)
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8002")
	if err != nil {
		log.Println("Error resolving UDP address for capn proto: %v", err)
	}
	go cphandler.ServeUDP(a, addr)
	tcpaddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:8003")
	if err != nil {
		log.Println("Error resolving TCP address for msgpack %v", err)
	}
	go mphandler.ServeTCP(a, tcpaddr)
	go a.Serve()
	idx := 0
	for {
		time.Sleep(5 * time.Second)
		idx += 5
		if idx == *config.Profile.BenchmarkTimer {
			if *config.Profile.MemProfile != "" {
				f, err := os.Create(*config.Profile.MemProfile)
				if err != nil {
					log.Panic(err)
				}
				pprof.WriteHeapProfile(f)
				f.Close()
				return
			}
			if *config.Profile.CpuProfile != "" {
				return
			}
		}
	}
	//log.Panic(srv.ListenAndServe())

}
