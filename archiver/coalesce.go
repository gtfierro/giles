package archiver

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	COALESCE_TIMEOUT = 500  // milliseconds
	COALESCE_MAX     = 4000 // num readings
)

type StreamMap map[string](*StreamBuf)

type StreamBuf struct {
	incoming   chan *SmapMessage
	readings   [][]interface{}
	uuid       string
	unitOfTime UnitOfTime
	txc        *TransactionCoalescer
	idx        int
}

func NewStreamBuf(uuid string, uot UnitOfTime, readings [][]interface{}) *StreamBuf {
	sb := &StreamBuf{uuid: uuid, unitOfTime: uot,
		incoming: make(chan *SmapMessage, COALESCE_MAX),
		readings: readings,
		idx:      0}
	go sb.listen()
	return sb
}

func (sb *StreamBuf) listen() {
	timeout := time.After(time.Duration(COALESCE_TIMEOUT) * time.Millisecond)
	abort := make(chan bool, 1)
	for {
		select {
		case sm := <-sb.incoming:
			if diff := (len(sm.Readings) + sb.idx) - COALESCE_MAX; diff > 0 {
				sb.readings = append(sb.readings, make([][]interface{}, diff)...)
			}
			for idx, rdg := range sm.Readings {
				sb.readings[sb.idx+idx] = rdg
			}
			sb.idx += len(sm.Readings)
			if sb.idx >= COALESCE_MAX {
				abort <- true
				sb.txc.Commit(sb)
				return
			}
		case <-timeout:
			sb.txc.Commit(sb)
			return
		case <-abort:
			return
		}
	}
}

type TransactionCoalescer struct {
	tsdb    *TSDB
	store   *MetadataStore
	streams atomic.Value
	bufpool sync.Pool
	sync.Mutex
}

func NewTransactionCoalescer(tsdb *TSDB, store *MetadataStore) *TransactionCoalescer {
	txc := &TransactionCoalescer{tsdb: tsdb, store: store}
	txc.streams.Store(make(StreamMap))
	txc.bufpool = sync.Pool{
		New: func() interface{} {
			return make([][]interface{}, COALESCE_MAX)
		},
	}
	return txc
}

func (txc *TransactionCoalescer) AddSmapMessage(sm *SmapMessage) {
	var sb *StreamBuf
	streams := txc.streams.Load().(StreamMap)
	if sb, found := streams[sm.UUID]; found {
		sb.incoming <- sm
		return
	}
	uot := (*txc.store).GetUnitOfTime(sm.UUID)
	sb = NewStreamBuf(sm.UUID, uot, txc.bufpool.Get().([][]interface{}))
	sb.txc = txc
	txc.Lock()
	oldStreams := txc.streams.Load().(StreamMap)
	// check again
	if sb, found := oldStreams[sm.UUID]; found {
		sb.incoming <- sm
		txc.Unlock()
		return
	}
	newStreams := make(StreamMap)
	for k, v := range oldStreams {
		newStreams[k] = v
	}
	newStreams[sm.UUID] = sb
	txc.streams.Store(newStreams)
	txc.Unlock()
	go txc.AddSmapMessage(sm)
}

func (txc *TransactionCoalescer) Commit(sb *StreamBuf) {
	(*txc.tsdb).Add(sb)
	txc.Lock()
	defer txc.Unlock()
	oldStreams := txc.streams.Load().(StreamMap)
	newStreams := make(StreamMap)
	for k, v := range oldStreams {
		newStreams[k] = v
	}
	txc.bufpool.Put(sb.readings)
	delete(newStreams, sb.uuid)
	txc.streams.Store(newStreams)
}
