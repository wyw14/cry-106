package monitor

import (
	"sync"
	"time"

	"github.com/wyw14/cry-106/internal/alarm"
	"github.com/wyw14/cry-106/internal/model"
)

type SnapshotReader struct {
	mu      sync.RWMutex
	zones   map[string]model.ZoneState
	devices map[string]model.DeviceStatus
	updated time.Time
}

func NewSnapshotReader(zones []model.ZoneState) *SnapshotReader {
	reader := &SnapshotReader{zones: make(map[string]model.ZoneState), devices: make(map[string]model.DeviceStatus)}
	for _, zone := range zones {
		reader.zones[zone.ID] = zone
	}
	return reader
}

func (r *SnapshotReader) ReadZone(zoneID string) (alarm.ZoneSnapshot, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	zone, ok := r.zones[zoneID]
	return alarm.ZoneSnapshot{ZoneID: zone.ID, Phase: string(zone.Phase), Generation: zone.Generation}, ok
}

func (r *SnapshotReader) UpdateZone(zone model.ZoneState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.zones[zone.ID] = zone
	r.updated = time.Now().UTC()
}

func (r *SnapshotReader) UpdateDevice(status model.DeviceStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[status.DeviceID] = status
	r.updated = time.Now().UTC()
}

func (r *SnapshotReader) Zones() []model.ZoneState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.ZoneState, 0, len(r.zones))
	for _, zone := range r.zones {
		result = append(result, zone)
	}
	return result
}

func (r *SnapshotReader) Devices() []model.DeviceStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.DeviceStatus, 0, len(r.devices))
	for _, device := range r.devices {
		result = append(result, device)
	}
	return result
}
