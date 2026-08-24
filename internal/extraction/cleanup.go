package extraction

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/operation"
)

type CleanupRunner struct {
	owners *operation.Registry
	valves *operation.ValveController
}

func NewCleanupRunner(owners *operation.Registry, commander *operation.Commander) *CleanupRunner {
	return &CleanupRunner{owners: owners, valves: operation.NewValveController(commander)}
}

func (r *CleanupRunner) CloseGroup(retired Session) (bool, error) {
	resourceID := "borehole:" + retired.GroupID
	current, ok := r.owners.Current(resourceID)
	if !ok {
		return false, nil
	}
	if _, err := r.valves.Close(retired.GroupID, current); err != nil {
		return false, fmt.Errorf("close extraction group: %w", err)
	}
	return r.owners.Release(resourceID, current.OperationID, current.Generation), nil
}
