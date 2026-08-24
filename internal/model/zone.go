package model

import (
	"fmt"
	"time"
)

type ZonePhase string

const (
	ZoneNormal     ZonePhase = "normal"
	ZoneRestricted ZonePhase = "restricted"
	ZoneIsolated   ZonePhase = "isolated"
	ZoneEvacuating ZonePhase = "evacuating"
	ZoneRecovery   ZonePhase = "recovery"
)

type ZoneState struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Phase      ZonePhase `json:"phase"`
	Generation uint64    `json:"generation"`
	GasPercent float64   `json:"gas_percent"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (z ZoneState) Validate() error {
	if z.ID == "" || z.Name == "" {
		return fmt.Errorf("zone identity is required")
	}
	if z.Generation == 0 {
		return fmt.Errorf("zone %s has zero generation", z.ID)
	}
	if z.GasPercent < 0 || z.GasPercent > 100 {
		return fmt.Errorf("zone %s has invalid gas percentage %.3f", z.ID, z.GasPercent)
	}
	switch z.Phase {
	case ZoneNormal, ZoneRestricted, ZoneIsolated, ZoneEvacuating, ZoneRecovery:
		return nil
	default:
		return fmt.Errorf("zone %s has invalid phase %q", z.ID, z.Phase)
	}
}

func (z ZoneState) Advance(next ZonePhase, gas float64, at time.Time) (ZoneState, error) {
	updated := z
	updated.Phase = next
	updated.GasPercent = gas
	updated.Generation++
	updated.UpdatedAt = at.UTC()
	if err := updated.Validate(); err != nil {
		return ZoneState{}, err
	}
	return updated, nil
}
