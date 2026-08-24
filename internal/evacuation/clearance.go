package evacuation

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/door"
	"github.com/wyw14/cry-106/internal/interlock"
	"github.com/wyw14/cry-106/internal/model"
)

type ClearanceService struct {
	coordinator *Coordinator
	releases    *interlock.ZoneRelease
	access      *door.AccessController
}

func NewClearanceService(coordinator *Coordinator, releases *interlock.ZoneRelease, access *door.AccessController) *ClearanceService {
	return &ClearanceService{coordinator: coordinator, releases: releases, access: access}
}

func (s *ClearanceService) Begin(zoneID string, expected int) (Round, error) {
	round, err := s.coordinator.Start(zoneID, expected)
	if err != nil {
		return Round{}, err
	}
	s.releases.Bind(zoneID, round.ID)
	s.access.Block(zoneID, round.ID)
	return round, nil
}

func (s *ClearanceService) Apply(value model.AirClearance) (Round, error) {
	round, err := s.coordinator.ApplyClearance(value)
	if err != nil {
		return Round{}, err
	}
	if err := s.releases.Release(value.ZoneID, round.ID); err != nil {
		return Round{}, fmt.Errorf("release zone interlock: %w", err)
	}
	if !s.access.Release(value.ZoneID, round.ID) {
		return Round{}, fmt.Errorf("release zone access for evacuation %s", round.ID)
	}
	return round, nil
}
