package archiver

import (
	"net"
	"sync/atomic"
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
	conn   net.Conn
	closed bool
}

func (c *TSDBConn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *TSDBConn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *TSDBConn) Close() error {
	c.closed = true
	return c.conn.Close()
}

func (c *TSDBConn) IsClosed() bool {
	return c.closed
}

type ConnectionPool struct {
	pool chan *TSDBConn
	// ConnectionPool will call this function when it needs a new connection
	newConn func() *TSDBConn
	count   int64
}

func NewConnectionPool(newConn func() *TSDBConn, maxConnections int) *ConnectionPool {
	pool := &ConnectionPool{newConn: newConn, pool: make(chan *TSDBConn, maxConnections), count: 0}
	return pool
}

func (pool *ConnectionPool) Get() *TSDBConn {
	var c *TSDBConn
	select {
	case c = <-pool.pool:
	default:
		c = pool.newConn()
		atomic.AddInt64(&pool.count, 1)
		log.Info("Creating new connection in pool %v, %v", &c.conn, pool.count)
	}
	return c
}

func (pool *ConnectionPool) Put(c *TSDBConn) {
	if c.IsClosed() {
		atomic.AddInt64(&pool.count, -1)
		return
	}
	select {
	case pool.pool <- c:
	default:
		c.Close()
		atomic.AddInt64(&pool.count, -1)
		log.Info("Releasing connection in pool, now %v", pool.count)
	}
}
