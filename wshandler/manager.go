package wshandler

type manager struct {
	// registered connections
	subscribers map[*WSSubscriber]bool

	// new connection request
	initialize chan *WSSubscriber

	// get rid of old connections
	remove chan *WSSubscriber
}

var m = manager{
	subscribers: make(map[*WSSubscriber]bool),
	initialize:  make(chan *WSSubscriber),
	remove:      make(chan *WSSubscriber),
}

func (m *manager) start() {
	for {
		select {
		case s := <-m.initialize:
			m.subscribers[s] = true
		case s := <-m.remove:
			if _, found := m.subscribers[s]; found {
				delete(m.subscribers, s)
				close(s.outbound)
			}
		}
	}
}
