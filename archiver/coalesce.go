package archiver

import (
	"sync"
	"time"
)

/**
How is transaction coalescing going to work?
PARAMS:
	coalesce timeout: 500 ms?
	early timeout: 100 messages?

We receive a *SmapMessage. Check rdb.sendbufs to see if we have a record for the UUID
IF WE DON'T: create a new slice [](*SmapMessage) for that UUID in rdb.sendbufs and create
			 a new mutex. Start a goroutine that starts a timer.After with the coalescing
			 timeout. If it is hit, then we COMMIT the slice of messages we have
IF WE DO: add the msg onto the found slice. If the slice now contains more than the limit,
		  then we COMMIT early

COMMIT: create a new RDB message (use a variant of NewMessage that takes in a slice) and send
it off. Might use sync.Pool for this later, but not sure. Erase the slice when that's done

**/

const (
	COALESCE_TIMEOUT = 500
	COALESCE_MAX     = 4000
)

type StreamBuf struct {
	sync.Mutex
	readings [][]interface{}
	abort    chan bool
	uuid     string
}

type Coalescer struct {
	tsdb *TSDB
	sync.Mutex
	streams map[string]*StreamBuf
}

func NewCoalescer(tsdb *TSDB) *Coalescer {
	return &Coalescer{tsdb: tsdb, streams: make(map[string]*StreamBuf, 100)}
}

func (c *Coalescer) GetStreamBuf(uuid string) *StreamBuf {
	c.Lock()
	defer c.Unlock()
	if sm, found := c.streams[uuid]; found {
		return sm
	}
	sm := &StreamBuf{uuid: uuid, readings: make([][]interface{}, 0, 100)}
	c.streams[uuid] = sm
	return sm
}

func (c *Coalescer) Add(sm *SmapMessage) {
	if sm.Readings == nil || len(sm.Readings) == 0 {
		return
	} // return early

	if len(sm.UUID) == 0 {
		log.Error("Reading has no UUID!")
		return
	}

	sb := c.GetStreamBuf(sm.UUID)

	sb.Lock()
	sb.readings = append(sb.readings, sm.Readings...)
	if len(sb.readings) == len(sm.Readings) { // empty! start afresh
		sb.abort = make(chan bool, 1)
		go func(abort chan bool, uuid string) {
			timeout := time.After(time.Duration(COALESCE_TIMEOUT) * time.Millisecond)
			select {
			case <-timeout:
				sb.Lock()
				c.commit(uuid)
				sb.Unlock()
				break
			case <-abort:
				break
			}
		}(sb.abort, sm.UUID)
	}

	// here we know we have a streambuf to use
	if len(sb.readings) >= COALESCE_MAX {
		sb.abort <- true // abort the timer
		c.commit(sm.UUID)
	}
	sb.Unlock()

}

func (c *Coalescer) commit(uuid string) {
	sb := c.GetStreamBuf(uuid)
	c.Lock()
	(*c.tsdb).Add(sb)
	delete(c.streams, uuid)
	c.Unlock()
}
