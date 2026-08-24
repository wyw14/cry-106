package interlock

import "sync"

type PowerBarrier struct {
	mu         sync.Mutex
	terminalAt map[string]uint64
	tripped    map[string]bool
}

func NewPowerBarrier() *PowerBarrier {
	return &PowerBarrier{terminalAt: make(map[string]uint64), tripped: make(map[string]bool)}
}

func (b *PowerBarrier) ApplyTrip(zoneID string, sequence uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if sequence >= b.terminalAt[zoneID] {
		b.terminalAt[zoneID] = sequence
		b.tripped[zoneID] = true
	}
}

func (b *PowerBarrier) ApplyRestore(zoneID string, sequence uint64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.tripped[zoneID] && sequence < b.terminalAt[zoneID] {
		return false
	}
	b.tripped[zoneID] = false
	return true
}

func (b *PowerBarrier) Tripped(zoneID string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.tripped[zoneID]
}
