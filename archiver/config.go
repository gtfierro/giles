package archiver

import (
	"code.google.com/p/gcfg"
	"fmt"
)

type Config struct {
	Archiver struct {
		TSDB        *string
		Metadata    *string
		Keepalive   *int
		EnforceKeys bool
		LogLevel    *string
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

	Venkman struct {
		Port    *string
		Address *string
	}

	HTTP struct {
		Enabled bool
		Port    *int
	}
	WebSockets struct {
		Enabled bool
		Port    *int
	}
	CapnProto struct {
		Enabled bool
		Port    *int
	}
	MsgPack struct {
		TcpEnabled bool
		TcpPort    *int
		UdpEnabled bool
		UdpPort    *int
	}

	SSH struct {
		Enabled            bool
		Port               *string
		PrivateKey         *string
		AuthorizedKeysFile *string
		User               *string
		Pass               *string
		PasswordEnabled    bool
		KeyAuthEnabled     bool
	}

	Profile struct {
		CpuProfile     *string
		MemProfile     *string
		BenchmarkTimer *int
		Enabled        bool
	}
}

func LoadConfig(filename string) *Config {
	var configuration Config
	err := gcfg.ReadFileInto(&configuration, filename)
	if err != nil {
		log.Error("No configuration file found at %v, so checking current directory for giles.cfg (%v)", filename, err)
	} else {
		return &configuration
	}
	err = gcfg.ReadFileInto(&configuration, "./giles.cfg")
	if err != nil {
		log.Fatal("Could not find configuration files ./giles.cfg. Try retreiving a sample from github.com/gtfierro/giles")
	} else {
		return &configuration
	}
	return &configuration
}

func PrintConfig(c *Config) {
	fmt.Println("Giles Configuration")
	fmt.Println("Connecting to Mongo at", *c.Mongo.Address, ":", *c.Mongo.Port)
	fmt.Println("Using Timeseries DB", *c.Archiver.TSDB)
	switch *c.Archiver.TSDB {
	case "readingdb":
		fmt.Println("	at address", *c.ReadingDB.Address, ":", *c.ReadingDB.Port)
	case "quasar":
		fmt.Println("	at address", *c.Quasar.Address, ":", *c.Quasar.Port)
	}
	fmt.Println("	with keepalive", *c.Archiver.Keepalive)

	if c.Profile.Enabled {
		fmt.Println("Profiling enabled for", *c.Profile.BenchmarkTimer, "seconds!")
		fmt.Println("CPU:", *c.Profile.CpuProfile)
		fmt.Println("Mem:", *c.Profile.MemProfile)
	} else {
		fmt.Println("Profiling disabled")
	}
}
