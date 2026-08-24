package alarm

import "sync"

type GasEpoch struct {
	ZoneID              string  `json:"zone_id"`
	CalibrationRevision uint64  `json:"calibration_revision"`
	Generation          uint64  `json:"generation"`
	MethanePercent      float64 `json:"methane_percent"`
	Tripped             bool    `json:"tripped"`
}

type GasEpochStore struct {
	mu      sync.Mutex
	current map[string]GasEpoch
	history map[string][]GasEpoch
}

func NewGasEpochStore() *GasEpochStore {
	return &GasEpochStore{current: make(map[string]GasEpoch), history: make(map[string][]GasEpoch)}
}

func (s *GasEpochStore) Record(epoch GasEpoch) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, exists := s.current[epoch.ZoneID]
	s.history[epoch.ZoneID] = append(s.history[epoch.ZoneID], epoch)
	if exists && epoch.CalibrationRevision < current.CalibrationRevision {
		return false
	}
	s.current[epoch.ZoneID] = epoch
	return !exists || epoch != current
}

func (s *GasEpochStore) Current(zoneID string) (GasEpoch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	epoch, ok := s.current[zoneID]
	return epoch, ok
}

func (s *GasEpochStore) History(zoneID string) []GasEpoch {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]GasEpoch, len(s.history[zoneID]))
	copy(result, s.history[zoneID])
	return result
}
