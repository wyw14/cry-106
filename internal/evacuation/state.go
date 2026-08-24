package evacuation

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/model"
)

type Round struct {
	ID         string    `json:"id"`
	ZoneID     string    `json:"zone_id"`
	Generation uint64    `json:"generation"`
	Phase      string    `json:"phase"`
	Expected   int       `json:"expected"`
	Accounted  int       `json:"accounted"`
	StartedAt  time.Time `json:"started_at"`
	ClearedAt  time.Time `json:"cleared_at,omitempty"`
}

type Coordinator struct {
	mu      sync.Mutex
	current map[string]Round
	history map[string]Round
}

func NewCoordinator() *Coordinator {
	return &Coordinator{current: make(map[string]Round), history: make(map[string]Round)}
}

func (c *Coordinator) Start(zoneID string, expected int) (Round, error) {
	if zoneID == "" || expected < 0 {
		return Round{}, fmt.Errorf("evacuation zone and non-negative headcount are required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	previous := c.current[zoneID]
	if previous.ID != "" {
		c.history[previous.ID] = previous
	}
	round := Round{ID: uuid.NewString(), ZoneID: zoneID, Generation: previous.Generation + 1, Phase: "evacuating", Expected: expected, StartedAt: time.Now().UTC()}
	c.current[zoneID] = round
	return round, nil
}

func (c *Coordinator) Account(zoneID, roundID string, count int) (Round, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.current[zoneID]
	if current.ID != roundID {
		return Round{}, fmt.Errorf("round %s is no longer active", roundID)
	}
	if count < current.Accounted || count > current.Expected {
		return Round{}, fmt.Errorf("invalid evacuation count %d of %d", count, current.Expected)
	}
	current.Accounted = count
	if current.Accounted == current.Expected {
		current.Phase = "air_review"
	}
	c.current[zoneID] = current
	return current, nil
}

func (c *Coordinator) ApplyClearance(clearance model.AirClearance) (Round, error) {
	if err := clearance.Validate(); err != nil {
		return Round{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.current[clearance.ZoneID]
	if current.ID != clearance.EvacuationID {
		return Round{}, fmt.Errorf("clearance belongs to retired evacuation %s", clearance.EvacuationID)
	}
	if !clearance.Approved || current.Phase != "air_review" {
		return Round{}, fmt.Errorf("evacuation %s cannot be cleared", current.ID)
	}
	current.Phase = "clear"
	current.ClearedAt = time.Now().UTC()
	c.current[current.ZoneID] = current
	return current, nil
}

func (c *Coordinator) Rounds() []Round {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]Round, 0, len(c.current))
	for _, round := range c.current {
		result = append(result, round)
	}
	return result
}
