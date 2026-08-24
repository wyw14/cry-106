package sensor

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-106/internal/alarm"
	"github.com/wyw14/cry-106/internal/model"
)

type Publisher interface {
	Publish(alarm.Alarm)
}

type ZoneState struct {
	mu        sync.RWMutex
	zones     map[string]model.ZoneState
	evaluator *alarm.Evaluator
	publisher Publisher
}

func NewZoneState(zones []model.ZoneState, evaluator *alarm.Evaluator, publisher Publisher) *ZoneState {
	state := &ZoneState{zones: make(map[string]model.ZoneState), evaluator: evaluator, publisher: publisher}
	for _, zone := range zones {
		state.zones[zone.ID] = zone
	}
	return state
}

func (s *ZoneState) ApplySample(sample model.GasSample) (alarm.Alarm, bool, error) {
	value, changed, err := s.evaluator.Apply(sample)
	if err != nil {
		return alarm.Alarm{}, false, err
	}
	if sample.CalibrationRevision < value.CalibrationRevision {
		return value, false, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	zone, ok := s.zones[sample.ZoneID]
	if !ok {
		return alarm.Alarm{}, false, fmt.Errorf("unknown gas zone %s", sample.ZoneID)
	}
	zone.GasPercent = sample.MethanePercent
	zone.UpdatedAt = time.Now().UTC()
	if value.Severity == alarm.SeverityCritical {
		zone, err = zone.Advance(model.ZoneIsolated, sample.MethanePercent, time.Now())
		if err != nil {
			return alarm.Alarm{}, false, err
		}
	}
	s.zones[zone.ID] = zone
	if changed && s.publisher != nil {
		s.publisher.Publish(value)
	}
	return value, changed, nil
}

func (s *ZoneState) ReadZone(zoneID string) (alarm.ZoneSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	zone, ok := s.zones[zoneID]
	return alarm.ZoneSnapshot{ZoneID: zone.ID, Phase: string(zone.Phase), Generation: zone.Generation}, ok
}

func (s *ZoneState) Zones() []model.ZoneState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.ZoneState, 0, len(s.zones))
	for _, zone := range s.zones {
		result = append(result, zone)
	}
	return result
}
