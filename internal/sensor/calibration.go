package sensor

import (
	"fmt"
	"sync"
	"time"
)

type Calibration struct {
	SensorID string    `json:"sensor_id"`
	Revision uint64    `json:"revision"`
	Offset   float64   `json:"offset"`
	Applied  time.Time `json:"applied"`
}

type Calibrations struct {
	mu      sync.Mutex
	current map[string]Calibration
}

func NewCalibrations() *Calibrations {
	return &Calibrations{current: make(map[string]Calibration)}
}

func (c *Calibrations) Publish(sensorID string, offset float64) (Calibration, error) {
	if sensorID == "" {
		return Calibration{}, fmt.Errorf("sensor ID is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.current[sensorID]
	next := Calibration{SensorID: sensorID, Revision: current.Revision + 1, Offset: offset, Applied: time.Now().UTC()}
	c.current[sensorID] = next
	return next, nil
}

func (c *Calibrations) Current(sensorID string) (Calibration, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.current[sensorID]
	return value, ok
}
