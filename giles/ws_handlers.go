package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }}

func WsTagsHandler(rw http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error establishing websocket: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	log.Debug("msgtype: %v, msg: %v, err: %v", msgtype, msg, err)
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	res, err := store.TagsUUID(uuid)
	ws.WriteJSON(res)
	log.Debug("got uuid %v", uuid, ws)
}

func WsQueryHandler(rw http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	log.Debug("msgtype: %v, msg: %v, err: %v", msgtype, msg, err)
}
