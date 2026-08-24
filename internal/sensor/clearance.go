package sensor

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/interlock"
	"github.com/wyw14/cry-106/internal/model"
)

type ClearanceDispatcher struct {
	releases *interlock.ZoneRelease
}

func NewClearanceDispatcher(releases *interlock.ZoneRelease) *ClearanceDispatcher {
	return &ClearanceDispatcher{releases: releases}
}

func (d *ClearanceDispatcher) Deliver(clearance model.AirClearance) error {
	if err := clearance.Validate(); err != nil {
		return err
	}
	if !clearance.Approved {
		return fmt.Errorf("air clearance %s was not approved", clearance.ID)
	}
	current, ok := d.releases.Current(clearance.ZoneID)
	if !ok {
		return fmt.Errorf("zone %s has no active evacuation", clearance.ZoneID)
	}
	if err := d.releases.Release(clearance.ZoneID, current); err != nil {
		return fmt.Errorf("release gas interlock: %w", err)
	}
	return nil
}
