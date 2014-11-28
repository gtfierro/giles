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
	// add the following SmapReading to the timeseries database
	Add(*SmapReading) bool
	// uuids, reference time, limit
	// retrieve data before reference time
	Prev([]string, uint64, int32) ([]SmapResponse, error)
	// retrieve data after reference time
	Next([]string, uint64, int32) ([]SmapResponse, error)
	// uuids, start time, end time
	GetData([]string, uint64, uint64) ([]SmapResponse, error)
	// get a new connection to the timeseries database
	GetConnection() (net.Conn, error)
	// return the number of live connections
	LiveConnections() int
	// Adds a pointer to metadata store for streamid/uuid conversion and the like
	AddStore(*Store)
}
