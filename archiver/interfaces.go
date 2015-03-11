// Defines the common interfaces that components of the archiver must implement
// in order to be compatible with Giles.

package archiver

import (
	"gopkg.in/mgo.v2/bson"
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
	AddStore(MetadataStore)
}

// The metadata store should support the following operations
type MetadataStore interface {
	// Returns true if the key @apikey is allowed to write to each of the
	// streams listed in @messages. Should check each SmapMessage.UUID value.
	CheckKey(apikey string, messages map[string]*SmapMessage) (bool, error)
	// Commits the metadata contained in each SmapMessage to the metadata
	// store. Should consult the following properties of SmapMessage:
	// Properties, Metadata, Actuator
	SaveTags(messages map[string]*SmapMessage)
	// Retrieves the tags indicated by @target for documents that match the
	// @where clause. If @is_distinct is true, then it will return a list of
	// distinct values for the tag @distinct_key
	GetTags(target bson.M, is_distinct bool, distinct_key string, where bson.M) ([]interface{}, error)
	// For all documents that match the where clause @where, apply the updates
	// contained in @updates, provided that the key @apikey is valid for all of
	// them
	UpdateTags(updates bson.M, apikey string, where bson.M) (bson.M, error)
	// Returns all metadata for a given UUID
	UUIDTags(uuid string) (bson.M, error)
	// Resolve a where clause to a slice of UUIDs
	GetUUIDs(where bson.M) ([]string, error)
	// Returns the unit of time for the stream identified by the given UUID.
	GetUnitOfTime(uuid string) UnitOfTime
	// For the given UUID, returns the uint32 stream id.
	//TODO: this is only needed for ReadingDB. Figure out a better way of using this method
	GetStreamId(uuid string) uint32
}

type APIKeyManager interface {

	// Returns True if the given api key exists
	ApiKeyExists(apikey string) (bool, error)

	// Creates a new key with the given name registered to the given email. The public argument
	// maps to the public attribute of the key. Returns the key
	NewKey(name, email string, public bool) (string, error)

	// Retrieves the key with the given name registered to the given email
	GetKey(name, email string) (string, error)

	// Lists all keys registered to the given email. Returns a list of k/v pairs for each
	// found key, giving us name, public, etc
	ListKeys(email string) ([]map[string]interface{}, error)

	// Deletes the key with the given name registered to the given email. Returns the key as well
	DeleteKeyByName(name, email string) (string, error)

	// Deletes the key with the given value.
	DeleteKeyByValue(key string) (string, error)

	// Retrieves the owner information for the given key
	Owner(key string) (map[string]interface{}, error)
}
