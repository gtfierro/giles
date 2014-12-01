package archiver

import (
	"net"
)

// Configuration for the archiver
type Config struct {
	// port on which to serve the archiver API
	Port int
	// which timeseries database is being used: "readingdb" or "quasar"
	TSDB string
	// IP:Port of the timeseries database
	TSDBAddress *net.TCPAddr
	// IP:Port of the MongoDB instance
	MongoAddress *net.TCPAddr
	// How long each connection to the timeseries database is kept open
	Keepalive int
}
