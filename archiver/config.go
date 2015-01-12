package archiver

import (
	"code.google.com/p/gcfg"
	"fmt"
)

type Config struct {
	Archiver struct {
		HttpPort  *string
		TSDB      *string
		Keepalive *int
	}

	ReadingDB struct {
		Port    *string
		Address *string
	}

	Quasar struct {
		Port    *string
		Address *string
	}

	Mongo struct {
		Port    *string
		Address *string
	}

	Profile struct {
		CpuProfile     *string
		MemProfile     *string
		BenchmarkTimer *int
		Enabled        bool
	}
}

func LoadConfig(filename string) *Config {
	var configuration *Config
	err := gcfg.ReadFileInto(&configuration, filename)
	if err != nil {
		log.Error("No configuration file found at %v, so checking current directory for giles.conf", filename)
	} else {
		return configuration
	}
	err = gcfg.ReadFileInto(&configuration, "./giles.conf")
	if err != nil {
		log.Fatal("Could not find configuration files ./giles.conf. Try retreiving a sample from github.com/gtfierro/giles")
	} else {
		return configuration
	}
	return configuration
}

func PrintConfig(c *Config) {
	fmt.Println("config")
}
