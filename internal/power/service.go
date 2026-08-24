package power

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/model"
	"github.com/wyw14/cry-106/internal/operation"
)

type State struct {
	ZoneID     string    `json:"zone_id"`
	Energized  bool      `json:"energized"`
	Generation uint64    `json:"generation"`
	Reason     string    `json:"reason"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Service struct {
	mu        sync.Mutex
	journal   *journal.Store
	owners    *operation.Registry
	commander *operation.Commander
	states    map[string]State
}

func NewService(events *journal.Store, owners *operation.Registry, commander *operation.Commander) *Service {
	return &Service{journal: events, owners: owners, commander: commander, states: make(map[string]State)}
}

func (s *Service) Set(ctx context.Context, zoneID string, energized bool, reason string) (State, error) {
	operationID := uuid.NewString()
	owner, err := s.owners.Acquire("feeder:"+zoneID, operationID, "power")
	if err != nil {
		return State{}, err
	}
	s.mu.Lock()
	current := s.states[zoneID]
	next := State{ZoneID: zoneID, Energized: energized, Generation: current.Generation + 1, Reason: reason, UpdatedAt: time.Now().UTC()}
	s.mu.Unlock()
	eventKind := model.EventPowerTarget
	if !energized && reason == "gas_trip" {
		eventKind = model.EventGasTrip
	}
	event, err := model.NewEvent(uuid.NewString(), eventKind, zoneID, next.Generation, next, time.Now())
	if err != nil {
		s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation)
		return State{}, err
	}
	if _, err := s.journal.Append(ctx, event); err != nil {
		s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation)
		return State{}, fmt.Errorf("append power target: %w", err)
	}
	target := "open"
	if energized {
		target = "closed"
	}
	if _, err := s.commander.Issue("feeder-"+zoneID, model.DeviceFeeder, target, owner); err != nil {
		s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation)
		return State{}, err
	}
	s.mu.Lock()
	s.states[zoneID] = next
	s.mu.Unlock()
	s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation)
	return next, nil
}

func (s *Service) States() []State {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]State, 0, len(s.states))
	for _, state := range s.states {
		result = append(result, state)
	}
	return result
}
