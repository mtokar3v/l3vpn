package nat

import "sync"

const (
	FstPort = 49152
	LstPort = 65535
)

type Socket struct {
	IPAddr string
	Port   uint16
}

type FiveTuple struct {
	Src      Socket
	Dst      Socket
	Protocol string
}

type NatTable struct {
	mu      sync.RWMutex
	table   map[FiveTuple]Socket
	lstPort uint16
}

func NewNatTable() *NatTable {
	return &NatTable{
		table:   make(map[FiveTuple]Socket),
		lstPort: FstPort,
	}
}

func (t *NatTable) Get(key FiveTuple) (Socket, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	val, ok := t.table[key]
	return val, ok
}

func (t *NatTable) Set(ft FiveTuple, org Socket) FiveTuple {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.table[ft] = org
	return ft
}

func (t *NatTable) RentPort() uint16 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lstPort = t.lstPort + 1 // TODO: reuse released ports
	if t.lstPort > LstPort {
		t.lstPort = FstPort
	}
	return t.lstPort
}
