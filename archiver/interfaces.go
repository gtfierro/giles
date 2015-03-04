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

// The metadata store should support the following operations
type MetadataStore interface {
	// Returns true if the key @apikey is allowed to write to each of the
	// streams listed in @messages. Should check each SmapMessage.UUID value.
	CheckKey(apikey string, messages map[string]*SmapMessage) (bool, error)
	// Commits the metadata contained in each SmapMessage to the metadata
	// store. Should consult the following properties of SmapMessage:
	// Properties, Metadata, Actuator
	SaveMetadata(messages map[string]*SmapMessage)
	// Retrieves the tags indicated by @target for documents that match the
	// @where clause. If @is_distinct is true, then it will return a list of
	// distinct values for the tag @distinct_key
	GetTags(target Dict, is_distinct bool, distinct_key string, where Dict) ([]interface{}, error)
	// For all documents that match the where clause @where, apply the updates
	// contained in @updates, provided that the key @apikey is valid for all of
	// them
	SetTags(updates Dict, apikey string, where Dict) (Dict, error)
	// Returns all metadata for a given UUID
	TagsUUID(uuid string) (Dict, error)
	// Resolve a where clause to a slice of UUIDs
	GetUUIDs(where Dict) ([]string, error)
	// Returns the unit of time for the stream identified by the given UUID.
	GetUnitOfTime(uuid string) UnitOfTime
}
