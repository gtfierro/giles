package main

import (
	"log"
	"net"
	"sync"
	"time"
)

//TODO: benchmark using 'delete' in a map vs setting entry to null

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
		cm.streams[uuid] = conn
		go cm.watchdog(uuid)
	}
}

func (cm *ConnectionMap) watchdog(uuid string) {
	timeout := time.After(time.Duration(cm.keepalive) * time.Second)
	conn := cm.streams[uuid]
	for {
		select {
		case data := <-conn.In:
			_, err := (*conn.conn).Write(*data)
			if err != nil {
				log.Println("Error writing data to ReadingDB", err)
			}
			timeout = time.After(time.Duration(cm.keepalive) * time.Second)
		case <-timeout:
			log.Println("timeout for", uuid)
			cm.Lock()
			(*conn.conn).Close()
			delete(cm.streams, uuid)
			//cm.streams[uuid] = nil
			cm.Unlock()
			break
		}
	}
}

func (cm *ConnectionMap) Stats() {
	log.Println("Live Connections:", len(cm.streams))
}

func test() {
	in := make(chan bool)
	to := time.After(5 * time.Second)
	go func() {
		for {
			select {
			case <-in:
				log.Println("hey")
				to = time.After(5 * time.Second)
			case <-to:
				log.Println("toolate")
			}
		}
	}()
	time.Sleep(3 * time.Second)
	in <- true
	time.Sleep(10 * time.Second)
}
