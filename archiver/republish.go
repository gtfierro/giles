package archiver

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

type RepublishClient struct {
	// the UUIDs we are interested in
	uuids []string
	in    chan []byte
	// a bool is sent on this channel when the client wants to be closed
	notify <-chan bool
	// this is how we handle writes back to the client
	subscriber Subscriber
}

// The Giles Republisher is the core of the sMAP "pub-sub" mechanism. Each of
// the protocol adapters (HTTP, WebSocket, MsgPack, CapnProto, etc) should be
// able to handle pub-sub over their respective protocols by writing a small
// shim to the core Archiver pub-sub API.
type Republisher struct {
	clients     [](*RepublishClient)
	subscribers map[string][](*RepublishClient)
	store       *Store // store is added in archiver.go
}

func NewRepublisher() *Republisher {
	return &Republisher{clients: [](*RepublishClient){}, subscribers: make(map[string][](*RepublishClient))}
}

// This is the Archiver API call. @s is a Subscriber, an interface that allows
// us to agnostically send "published" messages to the subscribed clients.
// Subscriber is wrapped in a RepublishClient internally so that the
// Republisher can keep track of necessary state. @query is the query string
// that describes what the client is subscribing to. This query should be a
// valid sMAP query
func (r *Republisher) HandleSubscriber(s Subscriber, query, apikey string) {
	tokens := tokenize(query)
	where := parseWhere(&tokens)
	uuids, err := r.store.GetUUIDs(where.ToBson())
	if err != nil {
		s.SendError(err)
		return
	}
	client := &RepublishClient{uuids: uuids, notify: s.GetNotify(), subscriber: s}
	r.clients = append(r.clients, client)
	for _, uuid := range uuids {
		r.subscribers[uuid] = append(r.subscribers[uuid], client)
	}
	log.Info("New subscriber for query: %v", query)
	log.Info("Clients: %v", len(r.clients))

	// wait for client to close connection, then tear down client
	<-client.notify
	for i, pubclient := range r.clients {
		if pubclient == client {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			break
		}
	}
	for uuid, clientlist := range r.subscribers {
		for i, pubclient := range clientlist {
			if pubclient == client {
				clientlist = append(clientlist[:i], clientlist[i+1:]...)
			}
		}
		r.subscribers[uuid] = clientlist
	}
}

// Publish @msg to all clients subscribing to @msg.UUID
func (r *Republisher) Republish(msg *SmapMessage) {
	for _, client := range r.subscribers[msg.UUID] {
		go client.subscriber.Send(msg)
	}
}
