package wshandler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	pongPeriod     = 60 * time.Second
	pingPeriod     = 10 * time.Second
	writeWait      = 10 * time.Second
	maxMessageSize = 512
)

type WSSubscriber struct {
	ws       *websocket.Conn
	rw       http.ResponseWriter
	closed   bool
	outbound chan []byte
	notify   chan bool
}

func NewWSSubscriber(ws *websocket.Conn, rw http.ResponseWriter) *WSSubscriber {
	wss := &WSSubscriber{ws: ws, rw: rw, notify: make(chan bool, 1), outbound: make(chan []byte, 512), closed: false}
	m.initialize <- wss
	go wss.dowrites()
	return wss
}

func (wss *WSSubscriber) Send(msg interface{}) {
	if msg != nil && !wss.closed {
		b, _ := json.Marshal(msg)
		wss.outbound <- b
	}
}

func (wss *WSSubscriber) SendError(e error) {
	log.Error("WS error", e.Error())
	//wss.ws.WriteMessage(websocket.TextMessage, []byte(e.Error()))
}

func (wss *WSSubscriber) GetNotify() <-chan bool {
	return wss.notify
}

func (wss *WSSubscriber) write(msgtype int, payload []byte) error {
	wss.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return wss.ws.WriteMessage(msgtype, payload)
}

func (wss *WSSubscriber) dowrites() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		m.remove <- wss
		wss.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-wss.outbound:
			if !ok {
				wss.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := wss.write(websocket.TextMessage, msg); err != nil && !wss.closed {
				return
			}
		case <-ticker.C:
			if err := wss.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (wss *WSSubscriber) doreads() {
	defer func() {
		m.remove <- wss
		wss.ws.Close()
	}()
	wss.ws.SetReadLimit(maxMessageSize)
	wss.ws.SetReadDeadline(time.Now().Add(pongPeriod))
	wss.ws.SetPongHandler(func(string) error { wss.ws.SetReadDeadline(time.Now().Add(pongPeriod)); return nil })
	//for {
	//	_, message, err := wss.ws.ReadMessage()
	//	if err != nil {
	//		break
	//	}
	//	log.Debug("message: %v", message)
	//}
}
