package archiver

import (
	"sync/atomic"
)

// This is a helper type for basic counting stats. Calling counter.Mark()
// will atomically add 1 to the internal count. Calling counter.Reset() will
// return the current count and return the count to 0
type counter struct {
	Count uint64
}

func newCounter() *counter {
	return &counter{Count: 0}
}

func (c *counter) Mark() {
	atomic.AddUint64(&c.Count, 1)
}

func (c *counter) Reset() uint64 {
	var returncount = c.Count
	atomic.StoreUint64(&c.Count, 0)
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
