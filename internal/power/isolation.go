package power

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-106/internal/interlock"
)

type Isolator struct {
	service *Service
	barrier *interlock.PowerBarrier
}

func NewIsolator(service *Service, barrier *interlock.PowerBarrier) *Isolator {
	return &Isolator{service: service, barrier: barrier}
}

func (i *Isolator) Trip(ctx context.Context, zoneID string) (State, error) {
	state, err := i.service.Set(ctx, zoneID, false, "gas_trip")
	if err != nil {
		return State{}, fmt.Errorf("trip zone feeder: %w", err)
	}
	i.barrier.ApplyTrip(zoneID, state.Generation)
	return state, nil
}

func (i *Isolator) Restore(ctx context.Context, zoneID string, sequence uint64) (State, error) {
	if !i.barrier.ApplyRestore(zoneID, sequence) {
		return State{}, fmt.Errorf("zone %s remains behind gas safety barrier", zoneID)
	}
	return i.service.Set(ctx, zoneID, true, "manual_clearance")
}
