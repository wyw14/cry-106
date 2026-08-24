package app

import (
	"context"
	"time"

	"github.com/wyw14/cry-106/internal/model"
	"github.com/wyw14/cry-106/internal/monitor"
	"github.com/wyw14/cry-106/internal/operation"
	"github.com/wyw14/cry-106/internal/power"
)

type Overview struct {
	Health      monitor.Health         `json:"health"`
	Zones       []model.ZoneState      `json:"zones"`
	Devices     []model.DeviceStatus   `json:"devices"`
	Commands    []model.DeviceCommand  `json:"commands"`
	Owners      []operation.Owner      `json:"owners"`
	Topology    model.TopologySnapshot `json:"topology"`
	Power       []power.State          `json:"power"`
	Evacuations int                    `json:"evacuations"`
	Extractions int                    `json:"extractions"`
	Activity    []monitor.Activity     `json:"activity"`
	GeneratedAt time.Time              `json:"generated_at"`
}

func (s *System) Overview() Overview {
	return Overview{
		Health: s.Health(), Zones: s.GasZones(), Devices: s.DeviceStatuses(), Commands: s.Commands(), Owners: s.owners.Snapshot(),
		Topology: s.Topology(), Power: s.power.States(), Evacuations: len(s.Evacuations()),
		Extractions: len(s.ExtractionSessions()), Activity: s.activity.Recent(20), GeneratedAt: time.Now().UTC(),
	}
}

func (s *System) TripPower(ctx context.Context, zoneID string) (power.State, error) {
	state, err := s.isolator.Trip(ctx, zoneID)
	if err != nil {
		return power.State{}, err
	}
	s.recordActivity("power", "zone feeder tripped", zoneID, map[string]any{"generation": state.Generation})
	return state, nil
}

func (s *System) RestorePower(ctx context.Context, zoneID string, sequence uint64) (power.State, error) {
	state, err := s.isolator.Restore(ctx, zoneID, sequence)
	if err != nil {
		return power.State{}, err
	}
	s.recordActivity("power", "zone feeder restored", zoneID, map[string]any{"generation": state.Generation})
	return state, nil
}

func (s *System) recordActivity(category, message, zoneID string, details map[string]any) {
	now := time.Now().UTC()
	s.activity.Add(monitor.Activity{ID: category + ":" + now.Format(time.RFC3339Nano), Category: category, Message: message, ZoneID: zoneID, CreatedAt: now, Details: details})
}

func (s *System) RecentActivity(limit int, zoneID string) []monitor.Activity {
	if zoneID != "" {
		items := s.activity.ByZone(zoneID)
		if limit > 0 && len(items) > limit {
			return items[len(items)-limit:]
		}
		return items
	}
	return s.activity.Recent(limit)
}
