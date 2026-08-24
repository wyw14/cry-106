package interlock

import (
	"fmt"
	"sync"
)

type ZoneRelease struct {
	mu         sync.Mutex
	evacuation map[string]string
	released   map[string]bool
}

func NewZoneRelease() *ZoneRelease {
	return &ZoneRelease{evacuation: make(map[string]string), released: make(map[string]bool)}
}

func (r *ZoneRelease) Bind(zoneID, evacuationID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.evacuation[zoneID] = evacuationID
	r.released[zoneID] = false
}

func (r *ZoneRelease) Release(zoneID, evacuationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.evacuation[zoneID] != evacuationID {
		return fmt.Errorf("air clearance belongs to retired evacuation %s", evacuationID)
	}
	r.released[zoneID] = true
	return nil
}

func (r *ZoneRelease) Released(zoneID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.released[zoneID]
}

func (r *ZoneRelease) Current(zoneID string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	evacuationID, ok := r.evacuation[zoneID]
	return evacuationID, ok
}
