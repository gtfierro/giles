package httphandler

import (
	"encoding/json"
	"github.com/gtfierro/giles/archiver"
	"net/http"
)

// Implements the archiver.Subscriber interface for the Republish mechanism
type HTTPSubscriber struct {
	rw      http.ResponseWriter
	_notify <-chan bool
	notify  chan bool
	closed  bool
}

func NewHTTPSubscriber(rw http.ResponseWriter) *HTTPSubscriber {
	rw.Header().Set("Content-Type", "application/json")
	_notify := rw.(http.CloseNotifier).CloseNotify()
	notify := make(chan bool)
	return &HTTPSubscriber{rw: rw, notify: notify, _notify: _notify, closed: false}
}

// called when we receive a new message
func (hs HTTPSubscriber) Send(msg *archiver.SmapMessage) {
	towrite := make(map[string]interface{})
	towrite[msg.Path] = archiver.SmapReading{Readings: msg.Readings, UUID: msg.UUID}
	bytes, err := json.Marshal(towrite)
	if err != nil {
		hs.rw.WriteHeader(500)
	} else {
		hs.rw.Write(bytes)
		hs.rw.Write([]byte{'\n', '\n'})
	}
	go func() {
		if flusher, ok := hs.rw.(http.Flusher); ok && !hs.closed {
			flusher.Flush()
		}
	}()
}
func (hs HTTPSubscriber) SendError(e error) {
	hs.rw.WriteHeader(500)
	hs.rw.Write([]byte(e.Error()))
}

func (hs HTTPSubscriber) GetNotify() <-chan bool {
	return hs.notify
}

func (hs HTTPSubscriber) watchForClose() {
	<-hs._notify
	hs.closed = true
	hs.notify <- true
}
