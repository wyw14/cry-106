package ventilation

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type StopController struct {
	mu     sync.Mutex
	delays map[string]time.Duration
}

func NewStopController() *StopController {
	return &StopController{delays: make(map[string]time.Duration)}
}

func (c *StopController) SetDelay(fanID string, delay time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delays[fanID] = delay
}

func (c *StopController) Stop(ctx context.Context, fanID string) error {
	c.mu.Lock()
	delay := c.delays[fanID]
	c.mu.Unlock()
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("fan stop deadline exceeded: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}
