package main

import (
	"net/http"
)

type RepublishClient struct {
	uuids  []string
	in     chan []byte
	writer http.ResponseWriter
}

/**
 * The republisher has a couple functions:
 * When a new WHERE clause is received, it creates a new RepublishClient
 * and adds the pointer to the correct maps of the desired UUIDs.
**/
type Republisher struct {
	Clients     [](*RepublishClient)
	Subscribers map[string][](*RepublishClient)
}

func NewRepublisher() *Republisher {
	return &Republisher{[](*RepublishClient){}, make(map[string][](*RepublishClient))}
}

func (r *Republisher) HandleSubscriber(rw http.ResponseWriter, query string) {
	tokens := tokenize(query)
	where := parseWhere(&tokens)
	uuids := store.GetUUIDs(where.ToBson())
	client := &RepublishClient{uuids: uuids, in: make(chan []byte), writer: rw}
	r.Clients = append(r.Clients, client)
	for _, uuid := range uuids {
		r.Subscribers[uuid] = append(r.Subscribers[uuid], client)
	}
	log.Info("New subscriber for query", query)
	log.Info("Clients:", len(r.Clients))

	rw.Header().Set("Content-Type", "application/json")
	notify := rw.(http.CloseNotifier).CloseNotify()

	// wait for client to close connection, then tear down client
	<-notify
	for i, pubclient := range r.Clients {
		if pubclient == client {
			r.Clients = append(r.Clients[:i], r.Clients[i+1:]...)
			break
		}
	}
	for uuid, clientlist := range r.Subscribers {
		for i, pubclient := range clientlist {
			if pubclient == client {
				clientlist = append(clientlist[:i], clientlist[i+1:]...)
			}
		}
		r.Subscribers[uuid] = clientlist
	}
}

func (r *Republisher) Republish(msg *SmapMessage) {
	for _, client := range r.Subscribers[msg.UUID] {
		client.writer.Write(msg.ToJson())
		client.writer.Write([]byte{'\n', '\n'})
		if flusher, ok := client.writer.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}
