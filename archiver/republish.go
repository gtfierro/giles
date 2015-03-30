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
	uuids []string
}

// Given a query string, tokenize and parse the query, also keeping
// track of what keys are mentioned in the query.
func HandleQuery(querystring string) *Query {
	q := &Query{}
	tokens := tokenize(querystring)
	where := parseWhere(&tokens)
	q.where = where.ToBson()
	q.keys = where.GetKeys()
	q.hash = QueryHash(strings.Join(tokens, ""))
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
	//TODO: maybe add a client ping to test for health?
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
//
// The basic approach will be to create a normalization for the metadata
// queries that are used to describe subscriptions to data. Once we have a
// hashable normalization, we can use them as keys in maps.
//
// In order to get efficient lookup of who to forward an incoming message to,
// we use a map of UUID -> list of clients. The UUID is provided in each
// message, so this is a simple lookup. TODO: how well does this scale?
//
// To facilitate efficient reevaluation of queries, the key is to quickly
// identify the subset of queries that need to be redone. When a change to the
// metadata occurs, the Republisher.MetadataChange method should be called with
// the relevant message.  This message will contain the new metadata that is to
// be considered.  During the parsing of queries (or perhaps in conjunction
// with the parsing), there should be a method that keeps track of the metadata
// keys that are mentioned in the query. In addition, we need a technique that
// normalizes a query and creates a perfect hash -- this is to get around
// complications of having queries either represented as flexible strings or as
// non-hashable maps.
//
// This gives the ability to have two helpful maps. The first is a simple map
// from query -> list of clients that have made that query. The second is a map
// from metadata key -> list of queries that include that key. What this allows
// us to do is on the receipt of a metadata change, we can use the keys involved
// in the change to look up a list of queries that could be affected, reevaluate
// each of those queries, and then adjust the subscription list of UUIDs on each
// of the clients associated with that query (if needed -- it could be the case
// that a metadata change does not affect a query or any clients).
type Republisher struct {
	sync.RWMutex

	// list of all republish clients (unique)
	clients [](*RepublishClient)

	// reference to the metadata store (should be added by archiver.go)
	store MetadataStore

	// stores hash -> query object
	queries map[QueryHash]*Query

	// query -> list of clients
	queryConcern map[QueryHash][](*RepublishClient)

	// key -> list of queries
	keyConcern map[string][]QueryHash

	// uuid -> queries concerning uuid
	uuidConcern map[string][]QueryHash
}

func NewRepublisher() *Republisher {
	return &Republisher{clients: [](*RepublishClient){},
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
	var err error
	q := HandleQuery(query)
	if prev_q, found := r.queries[q.hash]; found {
		// this query has already been done
		q = prev_q
	} else {
		// add it to the cache of queries
		q.uuids, err = r.store.GetUUIDs(q.where)
		if err != nil {
			s.SendError(err)
			return
		}
		r.queries[q.hash] = q

		// for each matched UUID, store the query that matched it
		for _, uuid := range q.uuids {
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
	//TODO: fixup removing client
	for i, pubclient := range r.clients {
		if pubclient == client {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			break
		}
	}
	r.Unlock()
}

// We call MetadataChange with an incoming sMAP message that includes
// changes to the metadata of a stream that could affect republish
// subscriptions
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
	// store old set of UUIDs
	olduuids := query.uuids
	// get new set of UUIDs
	query.uuids, err = r.store.GetUUIDs(query.where)
	if err != nil {
		log.Error("Received error when getting UUIDs for %v: (%v)", query.where, err)
		return
	}
	// store our query by its hash
	r.queries[query.hash] = query
	// list of UUIDs to remove clients from
	to_remove := []string{}
	// list of UUIDs to add clients to
	to_add := []string{}

	// remove duplicates between olduuids and new uuids
	newuuids := query.uuids
	for i, newuuid := range newuuids {
		for j, olduuid := range olduuids {
			if newuuid == olduuid {
				newuuids = append(newuuids[:i], newuuids[i+1:]...)
				olduuids = append(olduuids[:j], olduuids[j+1:]...)
				break
			}
		}
	}

	// add UUIDs to to_remove and to_add as necessary
	for _, newuuid := range newuuids {
		found := false
		for _, olduuid := range olduuids {
			if newuuid == olduuid {
				found = true
				break
			}
		}
		if !found {
			to_add = append(to_add, newuuid)
		}
	}

	// TODO: do this more intelligently
	for _, olduuid := range olduuids {
		found := false
		for _, newuuid := range newuuids {
			if newuuid == olduuid {
				found = true
				break
			}
		}
		if !found {
			to_remove = append(to_remove, olduuid)
		}
	}

	if len(to_add) > 0 || len(to_remove) > 0 {
		log.Debug("to add %v", to_add)
		log.Debug("to remove %v", to_remove)
	}

	for _, uuid := range to_remove {
		concerned := r.uuidConcern[uuid]
		for i, chash := range concerned {
			if chash == query.hash {
				concerned = append(concerned[:i], concerned[i+1:]...)
				break
			}
		}
		r.uuidConcern[uuid] = concerned
	}

	for _, uuid := range to_add {
		r.uuidConcern[uuid] = append(r.uuidConcern[uuid], query.hash)
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
