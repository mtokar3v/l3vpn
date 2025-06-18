package connection

import (
	"net"
	"sync"
)

type ConnectionPool struct {
	mu   sync.RWMutex
	pool map[string]net.Conn
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		pool: make(map[string]net.Conn),
	}
}

func (p *ConnectionPool) Get(addr string) (net.Conn, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.pool[addr]
	return val, ok
}

func (p *ConnectionPool) Set(addr string, conn net.Conn) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pool[addr] = conn
	return true
}
