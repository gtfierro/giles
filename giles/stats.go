package giles

import (
	"sync/atomic"
)

type Counter struct {
	Count uint64
}

func NewCounter() *Counter {
	return &Counter{Count: 0}
}

func (c *Counter) Mark() {
	atomic.AddUint64(&c.Count, 1)
}

func (c *Counter) Reset() uint64 {
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
func status() {
	log.Info("Repub clients:%d--Recv Adds:%d--Pend Write:%d--Live Conn:%d",
		len(republisher.Clients),
		incomingcounter.Reset(),
		pendingwritescounter.Reset(),
		tsdb.LiveConnections())
}
