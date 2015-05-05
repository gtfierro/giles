package archiver

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	COALESCE_TIMEOUT = 1000  // milliseconds
	COALESCE_MAX     = 16384 // num readings
)

type StreamMap map[string](*StreamBuf)
type StreamLockMap map[string](*sync.Mutex)

type StreamBuf struct {
	incoming   chan *SmapMessage
	uuid       string
	unitOfTime UnitOfTime
	readings   [][]interface{}
	txc        *TransactionCoalescer
	closed     atomic.Value
	timeout    <-chan time.Time
	abort      chan bool
	num        int64
	idx        int
	sync.Mutex
}

func NewStreamBuf(uuid string, uot UnitOfTime, readings [][]interface{}, txc *TransactionCoalescer) *StreamBuf {
	sb := &StreamBuf{uuid: uuid, unitOfTime: uot,
		incoming: make(chan *SmapMessage, COALESCE_MAX),
		num:      0,
		idx:      0,
		txc:      txc,
		readings: readings,
		abort:    make(chan bool, 1),
		timeout:  time.After(time.Duration(COALESCE_TIMEOUT) * time.Millisecond)}
	sb.closed.Store(false)
	go sb.watch()
	return sb
}

func (sb *StreamBuf) watch() {
	select {
	case <-sb.timeout:
		sb.commit()
	case <-sb.abort:
	}
}

func (sb *StreamBuf) isClosed() bool {
	return sb.closed.Load().(bool)
}

// Returns true if successfully added SmapMessage to the buffer,
// and false if the buffer is already closed
func (sb *StreamBuf) add(sm *SmapMessage) bool {
	// if no longer accepting readings, return false
	if sb.isClosed() {
		return false
	}

	sb.Lock()
	// if we are short some readings, append the space to the end
	if diff := (len(sm.Readings) + sb.idx) - COALESCE_MAX; diff > 0 {
		log.Debug("extending by %v", diff)
		sb.readings = append(sb.readings, make([][]interface{}, diff)...)
	}
	// copy over all the readings
	for idx, rdg := range sm.Readings {
		sb.readings[sb.idx+idx] = rdg
	}
	// advance our pointer
	sb.idx += len(sm.Readings)

	if sb.idx >= COALESCE_MAX {
		sb.abort <- true // cancels the timeout
		sb.Unlock()
		sb.commit()
		return true
	}
	sb.Unlock()
	return true
}

func (sb *StreamBuf) commit() {
	// close from further readings
	sb.closed.Store(true)
	// dispatch the commit
	sb.txc.Commit(sb)
}

type TransactionCoalescer struct {
	tsdb        *TSDB
	store       *MetadataStore
	streams     atomic.Value
	streamLocks atomic.Value
	bufpool     sync.Pool
	sync.Mutex
}

func NewTransactionCoalescer(tsdb *TSDB, store *MetadataStore) *TransactionCoalescer {
	txc := &TransactionCoalescer{tsdb: tsdb, store: store}
	txc.streams.Store(make(StreamMap))
	txc.streamLocks.Store(make(StreamLockMap))
	txc.bufpool = sync.Pool{
		New: func() interface{} {
			return make([][]interface{}, COALESCE_MAX)
		},
	}
	return txc
}

// Called to add an incoming SmapMessage to the underlying timeseries database. A SmapMessage contains
// an array of Readings and the UUID for the stream the readings belong to. The Readings must be added to
// a StreamBuffer for coalescing. This StreamBuffer is either a) pre-existing and still open, b) pre-existing and committing or
// c) not existing. In the
func (txc *TransactionCoalescer) AddSmapMessage(sm *SmapMessage) {
	var sb *StreamBuf

	// if we find the stream buffer and it is still accepting data, we write to that
	// stream and then return
	streams := txc.streams.Load().(StreamMap)
	if sb, found := streams[sm.UUID]; found && sb != nil {
		if sb.add(sm) {
			return
		}
	}

	txc.Lock()
	streams = txc.streams.Load().(StreamMap)
	// check again
	if sb, found := streams[sm.UUID]; found && sb != nil {
		if sb.add(sm) {
			txc.Unlock()
			return
		}
	}
	uot := (*txc.store).GetUnitOfTime(sm.UUID)
	sb = NewStreamBuf(sm.UUID, uot, txc.bufpool.Get().([][]interface{}), txc)
	newStreams := make(StreamMap, len(streams)+1)
	for k, v := range streams {
		newStreams[k] = v
	}
	newStreams[sm.UUID] = sb
	txc.streams.Store(newStreams)
	txc.Unlock()
	txc.AddSmapMessage(sm)
}

func (txc *TransactionCoalescer) Commit(sb *StreamBuf) {
	streams := txc.streams.Load().(StreamMap)
	if streams[sb.uuid] == sb {
		txc.Lock()
		newStreams := make(StreamMap, len(streams))
		for k, v := range streams {
			newStreams[k] = v
		}
		delete(newStreams, sb.uuid)
		txc.streams.Store(streams)
		txc.Unlock()
	}
	sb.Lock()
	(*txc.tsdb).Add(sb)
	txc.bufpool.Put(sb.readings)
	sb.Unlock()
}
