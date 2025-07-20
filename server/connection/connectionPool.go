package connection

import (
	"net"
	"sync"
)

type Pool struct {
	mu   sync.RWMutex
	pool map[string]net.TCPConn
}

func NewConnectionPool() *Pool {
	return &Pool{
		pool: make(map[string]net.TCPConn),
	}
}

func (p *Pool) Get(addr string) (*net.TCPConn, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.pool[addr]
	return &val, ok
}

func (p *Pool) Set(addr string, conn *net.TCPConn) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pool[addr] = *conn
	return true
}
