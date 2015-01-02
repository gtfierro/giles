// License stuff

// Package wshandler implements a WebSockets interface to the Archiver API at
// http://godoc.org/github.com/gtfierro/giles/archiver
//
// Overview
//
// The WebSockets interface expects data to be in the same JSON format as the HTTP
// interface. The routes are the same too, but are prefixed with '/ws', so '/api/query'
// becomes '/ws/api/query'.
//
// A DDP interface is also planned for the Giles Archiver
package wshandler

import (
	"github.com/gorilla/websocket"
	"github.com/gtfierro/giles/archiver"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

// Creates routes for WebSocket endpoints. These are the same as the normal HTTP/TCP endpoints, but are
// preceeded with '/ws/`. Not served until Archiver.Serve() is called.
func Handle(a *archiver.Archiver) {
	log.Notice("Handling WebSockets")
	a.R.POST("/ws/api/query", curryhandler(a, WsQueryHandler))
	a.R.GET("/ws/tags/uuid", curryhandler(a, WsTagsHandler))
	a.R.GET("/ws/tags/uuid/{uuid}", curryhandler(a, WsTagsHandler))
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }}

func WsAddReadingHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error establishing websocket: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	apikey := unescape(ps.ByName("key"))
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

func WsTagsHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error establishing websocket: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	log.Debug("msgtype: %v, msg: %v, err: %v", msgtype, msg, err)
	uuid := ps.ByName("uuid")
	res, err := a.TagsUUID(uuid)
	ws.WriteJSON(res)
	log.Debug("got uuid %v", uuid, ws)
}

func WsQueryHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	log.Debug("msgtype: %v, msg: %v, err: %v", msgtype, msg, err)
}

func unescape(s string) string {
	return strings.Replace(s, "%3D", "=", -1)
}

func curryhandler(a *archiver.Archiver, f func(*archiver.Archiver, http.ResponseWriter, *http.Request, httprouter.Params)) func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	return func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		f(a, rw, req, ps)
	}
}
