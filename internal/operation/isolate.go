package operation

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type FanStopper interface {
	Stop(context.Context, string) error
}

type DoorSealer interface {
	Seal(context.Context, string) error
}

type IsolationResult struct {
	ZoneID   string    `json:"zone_id"`
	FanError string    `json:"fan_error,omitempty"`
	Sealed   bool      `json:"sealed"`
	EndedAt  time.Time `json:"ended_at"`
}

func Isolate(ctx context.Context, zoneID, fanID, doorGroup string, stopTimeout time.Duration, fans FanStopper, doors DoorSealer) (IsolationResult, error) {
	stopContext, cancel := context.WithTimeout(ctx, stopTimeout)
	stopErr := fans.Stop(stopContext, fanID)
	cancel()
	sealErr := doors.Seal(ctx, doorGroup)
	result := IsolationResult{ZoneID: zoneID, Sealed: sealErr == nil, EndedAt: time.Now().UTC()}
	if stopErr != nil {
		result.FanError = stopErr.Error()
	}
	if sealErr != nil {
		return result, fmt.Errorf("seal emergency doors: %w", sealErr)
	}
	if stopErr != nil && !errors.Is(stopErr, context.DeadlineExceeded) {
		return result, fmt.Errorf("stop fan: %w", stopErr)
	}
	return result, nil
}
