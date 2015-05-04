package archiver

import (
	"sync/atomic"
)

// This is a helper type for basic counting stats. Calling counter.Mark()
// will atomically add 1 to the internal count. Calling counter.Reset() will
// return the current count and return the count to 0. Counter.Last contains
// the value returned by the last Reset().
type counter struct {
	Count uint64
	Last  uint64
}

func newCounter() *counter {
	return &counter{Count: 0, Last: 0}
}

func (c *counter) Mark() {
	atomic.AddUint64(&c.Count, 1)
}

func (c *counter) Reset() uint64 {
	uint64 returncount = atomic.LoadUint64(&c.Count)
	atomic.StoreUint64(&c.Count, 0)
	c.Last = returncount
	return returncount
}

/**
 * Prints status of the archiver:
 ** number of connected clients
 ** size of UUID cache
 ** connection status to database
 ** connection status to Mongo
 ** amount of incoming traffic since last call
 ** amount of api requests since last call
**/
func (a *Archiver) status() {
	log.Info("Repub clients:%d--Recv Adds:%d--Pend Write:%d--Live Conn:%d",
		len(a.republisher.clients),
		a.incomingcounter.Reset(),
		a.pendingwritescounter.Reset(),
		a.tsdb.LiveConnections())
}

type gilesStats struct {
	NumRepubClients  int    `json:"num_repub_clients"`
	IncomingMessages uint64 `json:"incoming_counter"`
	PendingWrites    uint64 `json:"pending_writes"`
	TSDBConnections  int    `json:"tsdb_connections"`
}

func (a *Archiver) Stats() *gilesStats {
	return &gilesStats{
		NumRepubClients:  len(a.republisher.clients),
		IncomingMessages: a.incomingcounter.Last,
		PendingWrites:    a.pendingwritescounter.Last,
		TSDBConnections:  a.tsdb.LiveConnections(),
	}
}
