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
}

func NewStreamBuf(uuid string, uot UnitOfTime) *StreamBuf {
	sb := &StreamBuf{uuid: uuid, unitOfTime: uot,
		incoming: make(chan *SmapMessage),
		readings: make([][]interface{}, 0, 100)}
	go sb.listen()
	return sb
}

func (sb *StreamBuf) listen() {
	timeout := time.After(time.Duration(COALESCE_TIMEOUT) * time.Millisecond)
	abort := make(chan bool, 1)
	for {
		select {
		case sm := <-sb.incoming:
			sb.readings = append(sb.readings, sm.Readings...)
			if len(sb.readings) >= COALESCE_MAX {
				abort <- true
				sb.txc.Commit(sb)
				break
			}
		case <-timeout:
			sb.readings = make([][]interface{}, 0, 100)
			sb.txc.Commit(sb)
			break
		case <-abort:
			break
		}
	}
}

type TransactionCoalescer struct {
	tsdb    *TSDB
	store   *MetadataStore
	streams atomic.Value
	sync.Mutex
}

func NewTransactionCoalescer(tsdb *TSDB, store *MetadataStore) *TransactionCoalescer {
	txc := &TransactionCoalescer{tsdb: tsdb, store: store}
	txc.streams.Store(make(StreamMap))
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
	sb = NewStreamBuf(sm.UUID, uot)
	sb.txc = txc
	txc.Lock()
	oldStreams := txc.streams.Load().(StreamMap)
	newStreams := make(StreamMap)
	for k, v := range oldStreams {
		newStreams[k] = v
	}
	newStreams[sm.UUID] = sb
	txc.streams.Store(newStreams)
	txc.Unlock()
	txc.AddSmapMessage(sm)
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
	delete(newStreams, sb.uuid)
	txc.streams.Store(newStreams)
}
