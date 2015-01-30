package wshandler

import (
	"github.com/gorilla/websocket"
	"github.com/gtfierro/giles/archiver"
	"net/http"
)

type WSSubscriber struct {
	ws     *websocket.Conn
	rw     http.ResponseWriter
	notify <-chan bool
}

func NewWSSubscriber(ws *websocket.Conn, rw http.ResponseWriter) *WSSubscriber {
	notify := rw.(http.CloseNotifier).CloseNotify()
	return &WSSubscriber{ws: ws, rw: rw, notify: notify}
}

func (wss WSSubscriber) Send(msg *archiver.SmapMessage) {
	wss.ws.WriteJSON(msg)
}

func (wss WSSubscriber) SendError(e error) {
	wss.ws.WriteMessage(websocket.TextMessage, []byte(e.Error()))
}

func (wss WSSubscriber) GetNotify() *<-chan bool {
	return &wss.notify
}
