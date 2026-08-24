package ventilation

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/interlock"
	"github.com/wyw14/cry-106/internal/model"
	"github.com/wyw14/cry-106/internal/operation"
)

type Service struct {
	mu        sync.Mutex
	arbiter   *interlock.FanArbiter
	owners    *operation.Registry
	commander *operation.Commander
	targets   map[string]string
	active    map[string]operation.Owner
}

func NewService(arbiter *interlock.FanArbiter, owners *operation.Registry, commander *operation.Commander) *Service {
	return &Service{arbiter: arbiter, owners: owners, commander: commander, targets: make(map[string]string), active: make(map[string]operation.Owner)}
}

func (s *Service) Transfer(groupID, targetFan string) (operation.Owner, error) {
	return s.start(groupID, "transfer", targetFan+":forward-ramp")
}

func (s *Service) Reverse(groupID string) (operation.Owner, error) {
	return s.start(groupID, "reversal", groupID+":reverse-prepare")
}

func (s *Service) start(groupID, purpose, target string) (operation.Owner, error) {
	operationID := uuid.NewString()
	// Switching and reversal must arbitrate on the same fan-group ownership
	// state. If the coordination key embeds the purpose, transfer and reverse
	// acquire disjoint entries and both proceed to issue conflicting commands
	// to the same group (e.g. forward-ramp and reverse-prepare on V2), which
	// trips the coupling guard and stops the whole group. Key on the group
	// alone so at most one operation can hold the group at any time.
	coordinationKey := groupID
	if err := s.arbiter.Begin(coordinationKey, purpose); err != nil {
		return operation.Owner{}, err
	}
	owner, err := s.owners.Acquire("fan-group:"+coordinationKey, operationID, purpose)
	if err != nil {
		s.arbiter.Finish(coordinationKey, purpose, "rejected")
		return operation.Owner{}, err
	}
	if _, err := s.commander.Issue(groupID, model.DeviceFan, target, owner); err != nil {
		s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation)
		s.arbiter.Finish(coordinationKey, purpose, "failed")
		return operation.Owner{}, fmt.Errorf("issue fan target: %w", err)
	}
	s.mu.Lock()
	s.targets[groupID] = target
	s.active[owner.OperationID] = owner
	s.mu.Unlock()
	return owner, nil
}

func (s *Service) Complete(groupID string, owner operation.Owner) bool {
	if !s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation) {
		return false
	}
	finished := s.arbiter.Finish(groupID, owner.Purpose, "complete")
	if finished {
		s.mu.Lock()
		delete(s.active, owner.OperationID)
		s.mu.Unlock()
	}
	return finished
}

func (s *Service) CompleteOperation(operationID string) error {
	s.mu.Lock()
	owner, exists := s.active[operationID]
	s.mu.Unlock()
	if !exists {
		return fmt.Errorf("unknown active fan operation %s", operationID)
	}
	groupID := owner.ResourceID
	const prefix = "fan-group:"
	if len(groupID) <= len(prefix) || groupID[:len(prefix)] != prefix {
		return fmt.Errorf("operation %s has invalid fan resource %s", operationID, groupID)
	}
	if !s.Complete(groupID[len(prefix):], owner) {
		return fmt.Errorf("operation %s no longer owns fan group", operationID)
	}
	return nil
}

func (s *Service) ActiveOperations() []operation.Owner {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]operation.Owner, 0, len(s.active))
	for _, owner := range s.active {
		result = append(result, owner)
	}
	return result
}

func (s *Service) Targets() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]string, len(s.targets))
	for key, value := range s.targets {
		result[key] = value
	}
	return result
}
