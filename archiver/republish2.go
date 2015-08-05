package archiver

import (
	"fmt"
	"strings"
)

type QueryChangeSet struct {
	// new messages (streams) that match this query
	New map[string]*SmapMessage
	// list of streams that no longer match this query
	Del map[string]struct{}
}

func NewQueryChangeSet() *QueryChangeSet {
	return &QueryChangeSet{
		New: make(map[string]*SmapMessage),
		Del: make(map[string]struct{}),
	}
}

func (cs *QueryChangeSet) NewStream(uuid string, msg *SmapMessage) {
	cs.New[uuid] = msg
}

func (cs *QueryChangeSet) DelStream(uuid string) {
	cs.Del[uuid] = struct{}{}
}

func (cs *QueryChangeSet) AddMsg(msg *SmapMessage) {
	if _, found := cs.New[msg.UUID]; found {
		fmt.Println("add message!")
		cs.New[msg.UUID] = msg
	}
}

func (r *Republisher) HandleQuery2(query string) (*Query, error) {
	var (
		q      *Query
		myhash QueryHash
	)

	// parse the query string
	lex := r.a.qp.Parse(query)

	// verify that query is SELECT, APPLY or DATA.
	// All other query types cannot be subscribed to
	switch lex.query.qtype {
	case SELECT_TYPE, APPLY_TYPE, DATA_TYPE:
		break
	default:
		return q, fmt.Errorf("Query (%s) is not of type SELECT, APPLY or DATA", query)
	}

	// calculate the hash of this query so we can see if it has already
	// been created
	myhash = QueryHash(strings.Join(lex.tokens, ""))

	r.queriesLock.RLock()
	prev_query, found := r.queries[myhash]
	r.queriesLock.RUnlock()

	if found {
		// we found the query, and return i
		q = prev_query
	} else {
		// this is a new query, so we fill in fields on the Query struct
		q = new(Query)
		q.hash = myhash
		q.m_uuids = make(map[string]UUIDSTATE)
		q.where = lex.query.WhereBson()
		q.keys = lex.keys
		q.target = lex.query.Contents
		q.querytype = lex.query.qtype

		// resolve the query where clause to a set of UUIDs
		uuids, err := r.a.store.GetUUIDs(q.where)
		if err != nil {
			return q, err
		}

		// associate the matched UUIDs with this query
		for _, uuid := range uuids {
			q.m_uuids[uuid] = OLD
		}

		// store this query in the republisher map
		r.queriesLock.Lock()
		r.queries[q.hash] = q
		r.queriesLock.Unlock()

		// for each matched UUID, store a reference to this query
		r.uuidConcernLock.Lock()
		for uuid, _ := range q.m_uuids {
			if list, found := r.uuidConcern[uuid]; found {
				list = append(list, q.hash)
				r.uuidConcern[uuid] = list
			} else {
				r.uuidConcern[uuid] = []QueryHash{q.hash}
			}
		}
		r.uuidConcernLock.Unlock()

		// for each key in the query where clause, store a reference to this query
		r.keyConcernLock.Lock()
		for _, key := range q.keys {
			if queries, found := r.keyConcern[key]; found {
				queries = append(queries, q.hash)
				r.keyConcern[key] = queries
			} else {
				r.keyConcern[key] = []QueryHash{q.hash}
			}
		}
		r.keyConcernLock.Unlock()
	}
	return q, nil
}

func (r *Republisher) HandleSubscriber2(s Subscriber, query, apikey string) {
	// create or get reference to the parsed query
	q, err := r.HandleQuery2(query)
	if err != nil {
		s.SendError(err)
		return
	}

	// create new instance of a client
	client := &RepublishClient{notify: s.GetNotify(), subscriber: s, query: q}

	// add the client to the relevant lists
	r.Lock()
	{ // begin lock
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

	// wait for client to quit
	<-client.notify

	// remove client
	r.Lock()
	for i, pubclient := range r.clients {
		if pubclient == client {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			break
		}
	}
	//TODO: remove client from queryConcern
	r.Unlock()
}

// When we receive a new set of readings, we make a set (map[Queryhash]struct{}) of which
// queries could potentially be affected by the incoming data. We then reevaluate each of
// these queries, and keep track of which changed (true return value on ReevaluateQuery).
// For each of the clients for the changed queries, we send the updates.
// We look up which clients to send to based on what their where-clause is (looking at
// republisher.keyconcern for each key mentioned in msg.{Metadata, Properties, Actuator}
func (r *Republisher) ChangeSubscriptions(readings map[string]*SmapMessage) []QueryHash {
	var (
		reeval    map[QueryHash]struct{}
		changed   = []QueryHash{}
		changeset = NewQueryChangeSet()
	)
	r.keyConcernLock.RLock()
	for _, msg := range readings {
		reeval = make(map[QueryHash]struct{})

		if msg.Metadata != nil {
			for key, _ := range msg.Metadata {
				for _, query := range r.keyConcern["Metadata."+key] {
					if _, found := reeval[query]; !found {
						reeval[query] = struct{}{}
					}
				}
			}
		}
		if msg.Properties != nil {
			for key, _ := range msg.Properties {
				for _, query := range r.keyConcern["Properties."+key] {
					if _, found := reeval[query]; !found {
						reeval[query] = struct{}{}
					}
				}
			}
		}
		if msg.Actuator != nil {
			for key, _ := range msg.Actuator {
				for _, query := range r.keyConcern["Actuator."+key] {
					if _, found := reeval[query]; !found {
						reeval[query] = struct{}{}
					}
				}
			}
		}

		// reevaluate the queries
		for queryhash, _ := range reeval {
			if r.ReevaluateQuery(queryhash, changeset) {
				fmt.Println("changeset add", msg, msg.UUID)
				changeset.AddMsg(msg)
				changed = append(changed, queryhash)
			}
		}
	}
	r.keyConcernLock.RUnlock()

	for uuid, newmsg := range changeset.New {
		fmt.Println("NEW", uuid, newmsg)
	}

	for uuid, _ := range changeset.Del {
		fmt.Println("Del", uuid)
	}

	return changed
}

// reevaluate the query corresponding to the given QueryHash. Return true
// if the results of the query changed (streams add or remove)
func (r *Republisher) ReevaluateQuery(qh QueryHash, cs *QueryChangeSet) bool {
	var (
		query   *Query
		found   bool
		changed = false
	)
	// fetch the query corresponding to this hash
	r.queriesLock.RLock()
	query, found = r.queries[qh]
	r.queriesLock.RUnlock()

	if !found {
		return changed // did not find the query, so we cannot reevaluate anything
	}

	// find UUIDs that match the where clause
	uuids, err := r.a.store.GetUUIDs(query.where)
	if err != nil {
		log.Error("Received error when getting UUIDs for %v: (%v)", query.where, err)
		return changed
	}

	// go through the matched UUIDs and mark as SAME or NEW
	// the ones to be deleted will be marked as OLD
	for _, uuid := range uuids {
		if _, found := query.m_uuids[uuid]; found {
			query.m_uuids[uuid] = SAME
		} else {
			//TODO: notify each repub client that this UUID is new
			query.m_uuids[uuid] = NEW
			cs.NewStream(uuid, nil)
			changed = true
		}
	}

	// remove UUIDs marked as OLD because they no longer match
	r.uuidConcernLock.Lock()
	for uuid, status := range query.m_uuids {
		if status == OLD {
			// get list of queries concerned with this UUID
			concerned := r.uuidConcern[uuid]

			// remove this query from that list
			for i, hash := range concerned {
				if hash == query.hash {
					concerned = append(concerned[:i], concerned[i+1:]...)
					break // can't have duplicates, so can stop early
				}
			}
			r.uuidConcern[uuid] = concerned
			changed = true
			query.m_uuids[uuid] = DEL
			cs.DelStream(uuid)
			continue // do not mark as old
		} else if status == NEW {
			r.uuidConcern[uuid] = append(r.uuidConcern[uuid], query.hash)
		}

		// mark as old if status is NEW or SAME
		query.m_uuids[uuid] = OLD
	}
	r.uuidConcernLock.Unlock()

	// remove UUIDs marked for deletion
	for uuid, status := range query.m_uuids {
		if status == DEL {
			delete(query.m_uuids, uuid)
		}
	}

	// store our query by its hash
	r.queriesLock.Lock()
	r.queries[query.hash] = query
	r.queriesLock.Unlock()

	return changed
}

// When a metadata change comes in from somewhere other than a smap message, we
// calculate the subscription changes and notify subscribers
func (r *Republisher) RepublishKeyChanges(keys []string) []QueryHash {
	var (
		reeval    = make(map[QueryHash]struct{})
		changed   = []QueryHash{}
		changeset = NewQueryChangeSet()
	)

	// create the set of affected queries
	for _, key := range keys {
		for _, query := range r.keyConcern[key] {
			if _, found := reeval[query]; !found {
				reeval[query] = struct{}{}
			}
		}
	}

	// reevaluate each of the affected queries, and keep track
	// of which actually changed
	for queryhash, _ := range reeval {
		fmt.Println("examine", queryhash)
		if r.ReevaluateQuery(queryhash, changeset) {
			changed = append(changed, queryhash)
		}
	}

	for uuid, newmsg := range changeset.New {
		fmt.Println("NEW", uuid, newmsg)
	}

	for uuid, _ := range changeset.Del {
		fmt.Println("Del", uuid)
	}

	for _, query := range changed {
		fmt.Println("changed", query)
	}

	// notify each relevant subscriber of a changed query

	return changed
}

// We receive a new message from a client, and want to send it out to the subscribers.
// A subscriber is interested in 1 of 3 things: * (all metadata), data before now (most recent
// data point) or a list of metadata tags.
func (r *Republisher) RepublishReadings(readings map[string]*SmapMessage) {
	//var (
	//	queries []QueryHash
	//	found	bool
	//)
	// get list of queries affected by these readings
	affected_queries := r.ChangeSubscriptions(readings)

	for _, query := range affected_queries {
		fmt.Println("changed", query)
	}

	// find all queries that match this UUID
	//r.uuidConcernLock.RLock()
	//queries, found = r.uuidConcern[msg.UUID]
	//r.uuidConcernLock.RUnlock()

	//if !found {
	//	return
	//}

	//for _, _  = range queries {
	//}

}
