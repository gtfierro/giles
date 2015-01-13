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
var configfile = flag.String("c", "giles.cfg", "Path to Giles configuration file")

func main() {
	flag.Parse()
	config := archiver.LoadConfig(*configfile)
	archiver.PrintConfig(config)

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
		if config.Profile.Enabled && idx == *config.Profile.BenchmarkTimer {
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
