package wshandler

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gtfierro/giles/giles"
	"net/http"
	"strings"
)

// Creates routes for WebSocket endpoints. These are the same as the normal HTTP/TCP endpoints, but are
// preceeded with '/ws/`. Not served until Archiver.Serve() is called.
func Handle(a *giles.Archiver) {
	log.Notice("Handling WebSockets")
	a.R.HandleFunc("/ws/api/query", curryhandler(a, WsQueryHandler)).Methods("POST")
	a.R.HandleFunc("/ws/tags/uuid", curryhandler(a, WsTagsHandler)).Methods("GET")
	a.R.HandleFunc("/ws/tags/uuid/{uuid}", curryhandler(a, WsTagsHandler)).Methods("GET")
}


var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }}

func WsAddReadingHandler(a *giles.Archiver, rw http.ResponseWriter, req *http.Request) {
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

func WsTagsHandler(a *giles.Archiver, rw http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("Error establishing websocket: %v", err)
		return
	}
	msgtype, msg, err := ws.ReadMessage()
	log.Debug("msgtype: %v, msg: %v, err: %v", msgtype, msg, err)
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	res, err := a.TagsUUID(uuid)
	ws.WriteJSON(res)
	log.Debug("got uuid %v", uuid, ws)
}

func WsQueryHandler(a *giles.Archiver, rw http.ResponseWriter, req *http.Request) {
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

func curryhandler(a *giles.Archiver, f func(*giles.Archiver, http.ResponseWriter, *http.Request)) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		f(a, rw, req)
	}
}
