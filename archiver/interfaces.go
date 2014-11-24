// Defines the common interfaces that components of the archiver must implement
// in order to be compatible with Giles.

package archiver

import (
	"net"
)

// TSDB (or TimeSeries DataBase) is a subset of functionality expected by Giles
// for (timestamp, value) oriented database. The relevant read/write types are
// SmapReading and SmapResponse and can be found in json.go and readingdb.go
// respectively (although their locations are likely to change)

type TSDB interface {
	Add(*SmapReading) bool
	Prev([]string, uint64, int32) ([]SmapResponse, error)
	Next([]string, uint64, int32) ([]SmapResponse, error)
	GetData([]string, uint64, uint64) ([]SmapResponse, error)
	GetConnection() (net.Conn, error)
	LiveConnections() int
	// Adds a pointer to metadata store for streamid/uuid conversion and the like
	AddStore(*Store)
}
