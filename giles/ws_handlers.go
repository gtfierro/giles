package giles

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }}

func WsAddReadingHandler(a *Archiver, rw http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error establishing websocket: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	vars := mux.Vars(req)
	apikey := unescape(vars["key"])
	messages, err := handleJSON(req.Body)
	if err != nil {
		log.Error("Error handling JSON: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	log.Debug("msgtype: %v, msg: %v, err: %v", msgtype, msg, err)
	err = a.AddData(messages, apikey)
	if err != nil {
		ws.WriteJSON(map[string]string{"error": err.Error()})
		return
	}
	ws.WriteJSON(map[string]string{"status": "ok"})
}

func WsTagsHandler(a *Archiver, rw http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error establishing websocket: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	log.Debug("msgtype: %v, msg: %v, err: %v", msgtype, msg, err)
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	res, err := a.store.TagsUUID(uuid)
	ws.WriteJSON(res)
	log.Debug("got uuid %v", uuid, ws)
}

func WsQueryHandler(a *Archiver, rw http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	log.Debug("msgtype: %v, msg: %v, err: %v", msgtype, msg, err)
}
