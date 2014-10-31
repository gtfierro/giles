package main

import (
	"net"
	"sync"
	"time"
)

type Connection struct {
	conn *net.Conn
	In   chan *[]byte
}

type ConnectionMap struct {
	sync.Mutex
	streams   map[string]*Connection
	keepalive int
}

func (cm *ConnectionMap) Add(uuid string, data *[]byte) {
	cm.Lock()
	defer cm.Unlock()
	if conn := cm.streams[uuid]; conn != nil {
		conn.In <- data
	} else {
		log.Notice("new conn for %v", uuid)
		// start new watchdog
		c, err := tsdb.GetConnection()
		if err != nil {
			log.Panic("Error connecting to TSDB")
		}
		conn = &Connection{conn: &c, In: make(chan *[]byte)}
		if _, found := cm.streams[uuid]; !found {
			cm.streams[uuid] = conn
			go cm.watchdog(uuid)
		}
	}
}

func (cm *ConnectionMap) watchdog(uuid string) {
	var timeout <-chan time.Time
	timer := time.NewTimer(0)
	conn := cm.streams[uuid]
	for {
		if conn == nil {
			return
		}
		select {
		case data := <-conn.In:
			timer.Reset(time.Duration(cm.keepalive) * time.Second)
			timeout = timer.C
			pendingwritescounter.Mark()
			_, err := (*conn.conn).Write(*data)
			if err != nil {
				log.Error("Error writing data to ReadingDB", err)
			}
		case <-timeout:
			log.Notice("timeout for %v", uuid)
			cm.Lock()
			timeout = nil
			(*conn.conn).Close()
			close(conn.In)
			delete(cm.streams, uuid)
			cm.Unlock()
			return
		}
	}
}

func (cm *ConnectionMap) LiveConnections() int {
	return len(cm.streams)
}
