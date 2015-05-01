package archiver

import (
	"net"
)

// For handling connections to the TSDB, we want to have a pool of long-lived
// connection objects that escape Go's garbage collection cycles. Objects
// stored in sync.Pool that are not referenced are garbage collected by Go, so
// we want to use a buffered channel to maintain non GC-able references to
// connections. When a coalesced buffer wants to write to a TSDB, it can grab a
// connection from the channel, and when it is finished, it can return it to
// the channel. Buffered channels give us a way to place a maximum number of
// connections as well, which is nice.

type TSDBConn struct {
	conn net.Conn
}

func (c *TSDBConn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *TSDBConn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *TSDBConn) Close() error {
	return c.conn.Close()
}

type ConnectionPool struct {
	pool chan *TSDBConn
	// ConnectionPool will call this function when it needs a new connection
	newConn func() *TSDBConn
}

func NewConnectionPool(newConn func() *TSDBConn, maxConnections int) *ConnectionPool {
	return &ConnectionPool{newConn: newConn, pool: make(chan *TSDBConn, maxConnections)}
}

func (pool *ConnectionPool) Get() *TSDBConn {
	var c *TSDBConn
	select {
	case c = <-pool.pool:
	default:
		c = pool.newConn()
		log.Info("Creating new connection in pool %v", &c.conn)
	}
	return c
}

func (pool *ConnectionPool) Put(c *TSDBConn) {
	select {
	case pool.pool <- c:
	default:
		log.Info("Releasing connection in pool")
	}
}
