package interlock

import (
	"fmt"
	"sync"
)

type FanArbiter struct {
	mu       sync.Mutex
	busy     map[string]string
	terminal map[string]string
}

func NewFanArbiter() *FanArbiter {
	return &FanArbiter{busy: make(map[string]string), terminal: make(map[string]string)}
}

func (a *FanArbiter) Begin(groupID, purpose string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if owner := a.busy[groupID]; owner != "" {
		return fmt.Errorf("fan group %s already performing %s", groupID, owner)
	}
	a.busy[groupID] = purpose
	return nil
}

func (a *FanArbiter) Finish(groupID, purpose, terminal string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.busy[groupID] != purpose {
		return false
	}
	delete(a.busy, groupID)
	a.terminal[groupID] = terminal
	return true
}

func (a *FanArbiter) Current(groupID string) (string, string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.busy[groupID], a.terminal[groupID]
}
