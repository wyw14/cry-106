package operation

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/model"
)

type ValveController struct {
	commands *Commander
}

func NewValveController(commands *Commander) *ValveController {
	return &ValveController{commands: commands}
}

func (v *ValveController) Open(groupID string, owner Owner) (model.DeviceCommand, error) {
	command, err := v.commands.Issue(groupID, model.DeviceValve, "open", owner)
	if err != nil {
		return model.DeviceCommand{}, fmt.Errorf("open extraction valves: %w", err)
	}
	return command, nil
}

func (v *ValveController) Close(groupID string, owner Owner) (model.DeviceCommand, error) {
	command, err := v.commands.Issue(groupID, model.DeviceValve, "closed", owner)
	if err != nil {
		return model.DeviceCommand{}, fmt.Errorf("close extraction valves: %w", err)
	}
	return command, nil
}
