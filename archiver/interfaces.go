// Defines the common interfaces that components of the archiver must implement
// in order to be compatible with Giles.

package archiver

import (
	"gopkg.in/mgo.v2/bson"
	"net"
)

// TSDB (or TimeSeries DataBase) is a subset of functionality expected by Giles
// for (timestamp, value) oriented database. The relevant read/write types are
// SmapReading and can be found in json.go and readingdb.go
// respectively (although their locations are likely to change).
// The UnitOfTime parameters indicate how to interpret the timesteps that are
// given as parameters
type TSDB interface {
	// add the following SmapReading to the timeseries database
	Add(*StreamBuf) bool
	// uuids, reference time, limit, unit of time
	// retrieve data before reference time
	Prev([]string, uint64, int32, UnitOfTime) ([]SmapNumbersResponse, error)
	// retrieve data after reference time
	Next([]string, uint64, int32, UnitOfTime) ([]SmapNumbersResponse, error)
	// uuids, start time, end time, unit of time
	GetData([]string, uint64, uint64, UnitOfTime) ([]SmapNumbersResponse, error)
	// get a new connection to the timeseries database
	GetConnection() (net.Conn, error)
	// return the number of live connections
	LiveConnections() int
	// Adds a pointer to metadata store for streamid/uuid conversion and the like
	AddStore(MetadataStore)
}

// The object store is a database for binary blobs rather than explicit
// timeseries data.
type ObjectStore interface {
	// archive the given SmapMessage that contains non-numerical Readings
	AddObject(*SmapMessage) (bool, error)
	// retrieve blob closest before the reference time for the given UUID
	PrevObject(string, uint64, UnitOfTime) (SmapObjectResponse, error)
	// retrieve blob closest after the reference time for the given UUIDs
	NextObject(string, uint64, UnitOfTime) (SmapObjectResponse, error)
	// retrieves all blobs between the start/end times for the given UUIDs
	GetObjects(string, uint64, uint64, UnitOfTime) (SmapObjectResponse, error)
	// Adds a pointer to metadata store for streamid/uuid conversion and the like
	AddStore(MetadataStore)
}

// The metadata store should support the following operations
type MetadataStore interface {
	// if called with True (this is default), checks all API keys. For testing or
	// "sandbox" deployments, it can be helpful to call this with False, which will
	// allow ALL operations on ANY streams.
	EnforceKeys(enforce bool)

	// Returns true if the key @apikey is allowed to write to each of the
	// streams listed in @messages. Should check each SmapMessage.UUID value.
	CheckKey(apikey string, messages map[string]*SmapMessage) (bool, error)

	// Associates metadata k/v pairs with non-terminal (non-timeseries) Paths
	SavePathMetadata(messages map[string]*SmapMessage) error

	// Associates metadata k/v pairs with timeserise paths. Inherits
	// from PathMetadata before applying timeseries-specific tags
	SaveTimeseriesMetadata(messages map[string]*SmapMessage) error

	// Retrieves the tags indicated by @target for documents that match the
	// @where clause. If @is_distinct is true, then it will return a list of
	// distinct values for the tag @distinct_key
	GetTags(target bson.M, is_distinct bool, distinct_key string, where bson.M) ([]interface{}, error)

	// Normal metadata save method
	SaveTags(messages map[string]*SmapMessage) error

	// For all documents that match the where clause @where, apply the updates
	// contained in @updates, provided that the key @apikey is valid for all of
	// them
	UpdateTags(updates bson.M, apikey string, where bson.M) (bson.M, error)

	// Removes all documents that match the where clause @where, provided that the
	// key @apikey is valid for them
	RemoveDocs(apikey string, where bson.M) (bson.M, error)

	// Unapplies all tags in the list @target for all documents that match the
	// where clause @where, after checking the API key
	RemoveTags(target bson.M, apikey string, where bson.M) (bson.M, error)

	// Returns all metadata for a given UUID
	UUIDTags(uuid string) (bson.M, error)

	// Resolve a where clause to a slice of UUIDs
	GetUUIDs(where bson.M) ([]string, error)

	// Returns the unit of time for the stream identified by the given UUID.
	GetUnitOfTime(uuid string) UnitOfTime

	// Returns the stream type for the stream identified by the given UUID
	GetStreamType(uuid string) StreamType
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

type Operator interface {
	Run(input interface{}) (interface{}, error)
}

type Reading interface {
	GetTime() uint64
	GetValue() interface{}
	IsObject() bool
}
