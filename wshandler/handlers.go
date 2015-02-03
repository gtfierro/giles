// License stuff

// Package wshandler implements a WebSockets interface to the Archiver API at
// http://godoc.org/github.com/gtfierro/giles/archiver
//
// Overview
//
// The WebSockets interface is designed as a less-hacky version of the HTTP republish
// mechanism, offering slightly augmented semantics for the connection. This interface
// is only for streaming data described by queries such as "select data before now where
// UUID = 123" or "select distinct Metadata/HVACZone". In the first case, we have a data
// query, so every time a new point is published to a stream matching the WHERE clause,
// a JSON message is sent over the WebSocket to the concerned client. In the second case,
// we have a metadata query, so every time the results of that query change, a JSON
// message is sent. These JSON messages have the same schema as the usual sMAP messages.
//
// To initialize a connection, the client uses its usual WebSocket setup/upgrade to
// the desired URL (e.g. '/ws/api/query'), and then sends the query along the WebSocket
// to the server. Whenever the server receives a message from a client, that message
// will be evaluated as a query and will change the nature of the republish subscription.
// If the query is invalid, the server will send back an error message and maintain the
// current subscription.
package wshandler

import (
	"github.com/gorilla/websocket"
	"github.com/gtfierro/giles/archiver"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

// Creates routes for WebSocket endpoints. Not served until Archiver.Serve() is called.
func Handle(a *archiver.Archiver) {
	log.Notice("Handling WebSockets")
	a.R.GET("/ws/republish", curryhandler(a, RepublishHandler))
	//a.R.POST("/ws/api/query", curryhandler(a, WsQueryHandler))
	//a.R.GET("/ws/tags/uuid", curryhandler(a, WsTagsHandler))
	go m.start()
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }}

func RepublishHandler(a *archiver.Archiver, rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error establishing websocket: %v", err)
		return
	}
	//TODO: check message type
	msgtype, msg, err := ws.ReadMessage()
	apikey := unescape(ps.ByName("key"))
	s := NewWSSubscriber(ws, rw)
	a.HandleSubscriber(s, string(msg), apikey)
	log.Debug("msgtype: %v, msg: %v, err: %v, apikey: %v", msgtype, msg, err, apikey)
}

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
