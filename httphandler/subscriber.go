package httphandler

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Implements the archiver.Subscriber interface for the Republish mechanism
type HTTPSubscriber struct {
	rw      http.ResponseWriter
	_notify <-chan bool
	notify  chan bool
	send    chan interface{}
	closed  bool
	sync.RWMutex
}

func NewHTTPSubscriber(rw http.ResponseWriter) *HTTPSubscriber {
	rw.Header().Set("Content-Type", "application/json")
	_notify := rw.(http.CloseNotifier).CloseNotify()
	notify := make(chan bool)
	hs := &HTTPSubscriber{rw: rw,
		notify:  notify,
		_notify: _notify,
		send:    make(chan interface{}, 1000),
		closed:  false}
	go hs.watchForClose()
	go hs.flushSend()
	return hs
}

// called when we receive a new message
func (hs *HTTPSubscriber) Send(msg interface{}) {
	hs.send <- msg
}

func (hs *HTTPSubscriber) flushSend() {
	for msg := range hs.send {
		bytes, err := json.Marshal(msg)
		hs.writeAndFlush(bytes, err)
	}
}

func (hs *HTTPSubscriber) writeAndFlush(data []byte, err error) {
	hs.Lock()
	if hs.closed {
		return
	}
	if err != nil {
		log.Error("HTTP Subscribe Error %v", err)
		hs.rw.WriteHeader(500)
		hs.rw.Write([]byte(err.Error()))
	} else {
		hs.rw.Write(data)
		hs.rw.Write([]byte{'\n', '\n'})
	}

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
