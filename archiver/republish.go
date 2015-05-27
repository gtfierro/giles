package archiver

import (
	"gopkg.in/mgo.v2/bson"
	"strings"
	"sync"
)

type QueryHash string

// hashable (Type A in a map) version of a query
type Query struct {
	// list of keys associated with this query
	keys []string
	// parsed where clause
	where bson.M
	// used to compare two different query objects
	hash QueryHash
	// UUIDs that match this query
	m_uuids map[string]UUIDSTATE
}

type UUIDSTATE uint

const (
	OLD UUIDSTATE = iota
	NEW
	SAME
	DEL
)

// Given a query string, tokenize and parse the query, also keeping
// track of what keys are mentioned in the query.
func (r *Republisher) HandleQuery(querystring string) *Query {
	q := &Query{}
	lex := r.a.qp.Parse("select * where " + querystring)
	q.where = lex.query.WhereBson()
	q.keys = lex.keys
	q.hash = QueryHash(strings.Join(lex.tokens, ""))
	q.m_uuids = make(map[string]UUIDSTATE)
	return q
}

// Subscriber is an interface that should be implemented by each protocol
// adapter that wants to support sMAP republish pub-sub.
type Subscriber interface {
	// Called by the Republisher when there is a new message to send to the
	// client. Send should transform the message to the appropriate format
	// before forwarding to the actual client.
	Send(*SmapMessage)

	// Called by Republisher when there is an error with the subscription
	SendError(error)

	// GetNotify is called by the Republisher to get a pointer to a "notify"
	// channel. When the client is closed and no longer wants to subscribe,
	// a value should be sent on the returned channel to signal to the Republisher
	// to unsubscribe the client. The client can of course disconnect on its own
	// without notifying the Republisher, but this means we cannot protect against
	// memory leaks resulting from infinitely adding new clients
	GetNotify() <-chan bool
}

// This is the type used within the Republisher to track the subscribers
type RepublishClient struct {
	// query made by this client
	query string

	// a bool is sent on this channel when the client wants to be closed
	notify <-chan bool

	// this is how we handle writes back to the client
	subscriber Subscriber
}

// This is a more thought-out version of the republisher that was first
// included in Giles.  The focus of this version of the republisher is SPEED:
// efficient discovery of who to deliver a new message to, and efficient
// reevaluation of queries in the face of new commands + data
type Republisher struct {
	sync.RWMutex

	// Pointer to archiver
	a *Archiver

	// list of all republish clients (unique)
	clients [](*RepublishClient)

	// stores hash -> query object
	queries map[QueryHash]*Query

	// query -> list of clients
	queryConcern map[QueryHash][](*RepublishClient)

	// key -> list of queries
	keyConcern map[string][]QueryHash

	// uuid -> queries concerning uuid
	uuidConcern map[string][]QueryHash
}

func NewRepublisher(a *Archiver) *Republisher {
	return &Republisher{
		a:            a,
		clients:      [](*RepublishClient){},
		queries:      make(map[QueryHash]*Query),
		queryConcern: make(map[QueryHash][](*RepublishClient)),
		keyConcern:   make(map[string][]QueryHash),
		uuidConcern:  make(map[string][]QueryHash)}
}

// This is the Archiver API call. @s is a Subscriber, an interface that allows
// us to agnostically send "published" messages to the subscribed clients.
// Subscriber is wrapped in a RepublishClient internally so that the
// Republisher can keep track of necessary state. @query is the query string
// that describes what the client is subscribing to. This query should be a
// valid sMAP query
func (r *Republisher) HandleSubscriber(s Subscriber, query, apikey string) {
	q := r.HandleQuery(query)
	if prev_q, found := r.queries[q.hash]; found {
		// this query has already been done
		q = prev_q
	} else {
		// add it to the cache of queries
		uuids, err := r.a.store.GetUUIDs(q.where)
		//q.uuids, err = r.a.store.GetUUIDs(q.where)
		if err != nil {
			s.SendError(err)
			return
		}

		for _, uuid := range uuids {
			q.m_uuids[uuid] = OLD
		}

		r.queries[q.hash] = q

		// for each matched UUID, store the query that matched it
		for uuid, _ := range q.m_uuids {
			var list []QueryHash
			var found bool
			if list, found = r.uuidConcern[uuid]; found {
				list = append(list, q.hash)
			} else {
				list = []QueryHash{q.hash}
			}
			r.uuidConcern[uuid] = list
		}

		// for each key in the query, store the query that mentions it
		for _, key := range q.keys {
			if queries, found := r.keyConcern[key]; found {
				queries = append(queries, q.hash)
				r.keyConcern[key] = queries
			} else {
				r.keyConcern[key] = []QueryHash{q.hash}
			}
		}
	}

	// create new instance of a client
	client := &RepublishClient{notify: s.GetNotify(), subscriber: s, query: query}

	r.Lock()
	{ // begin lock
		// add the client to the relevant lists
		if clients, found := r.queryConcern[q.hash]; found {
			clients = append(clients, client)
			r.queryConcern[q.hash] = clients
		} else {
			r.queryConcern[q.hash] = [](*RepublishClient){client}
		}

		r.clients = append(r.clients, client)
	} // end lock
	r.Unlock()

	log.Info("New subscriber for query \"%v\" with keys %v", query, q.keys)

	<-client.notify

	r.Lock()
	for i, pubclient := range r.clients {
		if pubclient == client {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			break
		}
	}
	r.Unlock()
}

// Same as MetadataChange, but operates on a known list of keys
// rather than a sMAP message
//TODO: store up queries and do r.EvaluateQuery once each at end. With this current scheme,
//      we can have  multiple queries be re-evaluated twice
func (r *Republisher) MetadataChangeKeys(keys []string) {
	for _, key := range keys {
		for _, query := range r.keyConcern[key] {
			r.EvaluateQuery(query)
		}
	}
}

// We call MetadataChange with an incoming sMAP message that includes
// changes to the metadata of a stream that could affect republish
// subscriptions
//TODO: store up queries and do r.EvaluateQuery once each at end. With this current scheme,
//      we can have  multiple queries be re-evaluated twice
func (r *Republisher) MetadataChange(msg *SmapMessage) {
	if msg.Metadata != nil {
		for key, _ := range msg.Metadata {
			key = "Metadata." + key
			for _, query := range r.keyConcern[key] {
				r.EvaluateQuery(query)
			}
		}
	}
	if msg.Properties != nil {
		for key, _ := range msg.Properties {
			key = "Properties." + key
			for _, query := range r.keyConcern[key] {
				r.EvaluateQuery(query)
			}
		}
	}
	if msg.Actuator != nil {
		for key, _ := range msg.Actuator {
			key = "Actuator." + key
			for _, query := range r.keyConcern[key] {
				r.EvaluateQuery(query)
			}
		}
	}
}

// Given a query hash, reevaluate the associated WHERE clause to get the
// new set of UUIDs that match the query. We now have a set of clients
// attached to a specific query and a set of clients associated with
// each of the UUIDs.
// Compare the previous list of UUIDs with the current list of UUIDs.
// For each UUID that is now in the set, add the list of concerned clients
// to the subscribers. For each UUID that is now NOT in the set, remove
// the list of concerned clients from that subscriber list
func (r *Republisher) EvaluateQuery(qh QueryHash) {
	var query *Query
	var found bool
	var err error
	if query, found = r.queries[qh]; !found {
		return
	}

	// mark UUIDs that match the new query with 'true'. Old UUIDs no longer
	// covered will still be marked as 'false'
	uuids, err := r.a.store.GetUUIDs(query.where)

	if err != nil {
		log.Error("Received error when getting UUIDs for %v: (%v)", query.where, err)
		return
	}

	for _, uuid := range uuids {
		if _, found := query.m_uuids[uuid]; found {
			query.m_uuids[uuid] = SAME
		} else {
			query.m_uuids[uuid] = NEW
		}
	}

	// store our query by its hash
	r.queries[query.hash] = query

	for uuid, status := range query.m_uuids {
		if status == OLD {
			concerned := r.uuidConcern[uuid]
			for i, chash := range concerned {
				if chash == query.hash {
					concerned = append(concerned[:i], concerned[i+1:]...)
					break
				}
			}
			r.uuidConcern[uuid] = concerned
			query.m_uuids[uuid] = DEL
			continue
		}
		if status == NEW {
			r.uuidConcern[uuid] = append(r.uuidConcern[uuid], query.hash)
		}
		query.m_uuids[uuid] = OLD
	}

	for uuid, status := range query.m_uuids {
		if status == DEL {
			delete(query.m_uuids, uuid)
		}
	}
}

// Publish @msg to all clients subscribing to @msg.UUID
func (r *Republisher) Republish(msg *SmapMessage) {
	// for all queries that resolve to this UUID
	if queries, found := r.uuidConcern[msg.UUID]; found {
		for _, hash := range queries {
			// get the list of subscribers for that query and forward the message
			for _, client := range r.queryConcern[hash] {
				go client.subscriber.Send(msg)
			}
		}
	}
}
