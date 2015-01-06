// Defines the common interfaces that components of the archiver must implement
// in order to be compatible with Giles.

package archiver

import (
	"net"
)

// TSDB (or TimeSeries DataBase) is a subset of functionality expected by Giles
// for (timestamp, value) oriented database. The relevant read/write types are
// SmapReading and SmapResponse and can be found in json.go and readingdb.go
// respectively (although their locations are likely to change).
// The UnitOfTime parameters indicate how to interpret the timesteps that are
// given as parameters
type TSDB interface {
	// add the following SmapReading to the timeseries database
	Add(*StreamBuf) bool
	// uuids, reference time, limit, unit of time
	// retrieve data before reference time
	Prev([]string, uint64, int32, UnitOfTime) ([]SmapResponse, error)
	// retrieve data after reference time
	Next([]string, uint64, int32, UnitOfTime) ([]SmapResponse, error)
	// uuids, start time, end time, unit of time
	GetData([]string, uint64, uint64, UnitOfTime) ([]SmapResponse, error)
	// get a new connection to the timeseries database
	GetConnection() (net.Conn, error)
	// return the number of live connections
	LiveConnections() int
	// Adds a pointer to metadata store for streamid/uuid conversion and the like
	AddStore(*Store)
}
