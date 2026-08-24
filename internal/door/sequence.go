package door

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type State struct {
	DoorID      string    `json:"door_id"`
	Position    string    `json:"position"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

type Controller struct {
	mu     sync.Mutex
	groups map[string][]string
	state  map[string]State
	delay  time.Duration
}

func NewController(groups map[string][]string) *Controller {
	return &Controller{groups: groups, state: make(map[string]State), delay: time.Millisecond}
}

func (c *Controller) SetDelay(delay time.Duration) {
	c.mu.Lock()
	c.delay = delay
	c.mu.Unlock()
}

func (c *Controller) Seal(ctx context.Context, groupID string) error {
	c.mu.Lock()
	doors, ok := c.groups[groupID]
	delay := c.delay
	c.mu.Unlock()
	if !ok || len(doors) == 0 {
		return fmt.Errorf("unknown emergency door group %s", groupID)
	}
	for _, doorID := range doors {
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			c.mu.Lock()
			c.state[doorID] = State{DoorID: doorID, Position: "closed", ConfirmedAt: time.Now().UTC()}
			c.mu.Unlock()
		}
	}
	return nil
}

func (c *Controller) States() []State {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]State, 0, len(c.state))
	for _, state := range c.state {
		result = append(result, state)
	}
	return result
}
