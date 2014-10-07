package main

import (
	"log"
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
	if conn := cm.streams[uuid]; conn != nil {
		conn.In <- data
	} else {
		log.Println("new conn for", uuid)
		// start new watchdog
		c, err := tsdb.GetConnection()
		if err != nil {
			log.Panic("Error connecting to TSDB")
		}
		conn = &Connection{conn: &c, In: make(chan *[]byte)}
		cm.Lock()
		cm.streams[uuid] = conn
		go cm.watchdog(uuid)
		cm.Unlock()
	}
}

func (cm *ConnectionMap) watchdog(uuid string) {
	timeout := time.After(time.Duration(cm.keepalive) * time.Second)
	conn := cm.streams[uuid]
	for {
		if conn == nil {
			return
		}
		select {
		case data := <-conn.In:
			timeout = time.After(time.Duration(cm.keepalive) * time.Second)
			pendingwritescounter.Mark()
			_, err := (*conn.conn).Write(*data)
			if err != nil {
				log.Println("Error writing data to ReadingDB", err)
			}
		case <-timeout:
			log.Println("timeout for", uuid)
			cm.Lock()
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
