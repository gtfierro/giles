package httphandler

import (
	"encoding/json"
	"github.com/gtfierro/giles/archiver"
	"net/http"
	"sync"
)

// Implements the archiver.Subscriber interface for the Republish mechanism
type HTTPSubscriber struct {
	rw      http.ResponseWriter
	_notify <-chan bool
	notify  chan bool
	closed  bool
	sync.RWMutex
}

func NewHTTPSubscriber(rw http.ResponseWriter) *HTTPSubscriber {
	rw.Header().Set("Content-Type", "application/json")
	_notify := rw.(http.CloseNotifier).CloseNotify()
	notify := make(chan bool)
	hs := &HTTPSubscriber{rw: rw, notify: notify, _notify: _notify, closed: false}
	go hs.watchForClose()
	return hs
}

// called when we receive a new message
func (hs *HTTPSubscriber) Send(msg *archiver.SmapMessage) {
	towrite := make(map[string]interface{})
	towrite[msg.Path] = archiver.SmapReading{Readings: msg.Readings, UUID: msg.UUID}
	bytes, err := json.Marshal(towrite)
	if hs.closed {
		return
	}
	if err != nil {
		log.Error("HTTP Subscribe Error %v", err)
		hs.rw.WriteHeader(500)
		hs.rw.Write([]byte(err.Error()))
	} else {
		hs.rw.Write(bytes)
		hs.rw.Write([]byte{'\n', '\n'})
	}

	hs.Lock()
	if flusher, ok := hs.rw.(http.Flusher); ok && !hs.closed {
		flusher.Flush()
	}
	hs.Unlock()
}
func (hs *HTTPSubscriber) SendError(e error) {
	log.Error("HTTP Subscribe Error %v", e)
	hs.rw.WriteHeader(500)
	hs.rw.Write([]byte(e.Error()))
}

func (hs *HTTPSubscriber) GetNotify() <-chan bool {
	return hs.notify
}

func (hs *HTTPSubscriber) watchForClose() {
	<-hs._notify
	hs.Lock()
	hs.closed = true
	hs.notify <- true
	hs.Unlock()
}
