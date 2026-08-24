package operation

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/model"
)

type Commander struct {
	mu       sync.Mutex
	commands []model.DeviceCommand
	status   map[string]model.DeviceStatus
}

func NewCommander() *Commander {
	return &Commander{status: make(map[string]model.DeviceStatus)}
}

func (c *Commander) Issue(deviceID string, kind model.DeviceKind, target string, owner Owner) (model.DeviceCommand, error) {
	command := model.DeviceCommand{
		ID: uuid.NewString(), DeviceID: deviceID, Kind: kind, Target: target,
		OperationID: owner.OperationID, OwnerGeneration: owner.Generation, IssuedAt: time.Now().UTC(),
	}
	if err := command.Validate(); err != nil {
		return model.DeviceCommand{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	status, exists := c.status[deviceID]
	if exists && !status.Accepts(command) {
		return model.DeviceCommand{}, fmt.Errorf("device %s rejected stale generation %d", deviceID, command.OwnerGeneration)
	}
	c.commands = append(c.commands, command)
	c.status[deviceID] = model.DeviceStatus{DeviceID: deviceID, Kind: kind, Actual: target, OwnerGeneration: owner.Generation, Healthy: true, ObservedAt: time.Now().UTC()}
	return command, nil
}

func (c *Commander) Commands() []model.DeviceCommand {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]model.DeviceCommand, len(c.commands))
	copy(result, c.commands)
	return result
}

func (c *Commander) Statuses() []model.DeviceStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]model.DeviceStatus, 0, len(c.status))
	for _, status := range c.status {
		result = append(result, status)
	}
	return result
}
