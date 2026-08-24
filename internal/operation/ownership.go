package operation

import (
	"fmt"
	"sync"
	"time"
)

type Owner struct {
	ResourceID  string    `json:"resource_id"`
	OperationID string    `json:"operation_id"`
	Generation  uint64    `json:"generation"`
	Purpose     string    `json:"purpose"`
	AcquiredAt  time.Time `json:"acquired_at"`
}

type Registry struct {
	mu      sync.Mutex
	owners  map[string]Owner
	nextGen map[string]uint64
}

func NewRegistry() *Registry {
	return &Registry{owners: make(map[string]Owner), nextGen: make(map[string]uint64)}
}

func (r *Registry) Acquire(resourceID, operationID, purpose string) (Owner, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if current, exists := r.owners[resourceID]; exists {
		return Owner{}, fmt.Errorf("resource %s is owned by %s for %s", resourceID, current.OperationID, current.Purpose)
	}
	r.nextGen[resourceID]++
	owner := Owner{ResourceID: resourceID, OperationID: operationID, Generation: r.nextGen[resourceID], Purpose: purpose, AcquiredAt: time.Now().UTC()}
	r.owners[resourceID] = owner
	return owner, nil
}

func (r *Registry) Transfer(resourceID, fromOperationID, toOperationID, purpose string) (Owner, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, exists := r.owners[resourceID]
	if !exists || current.OperationID != fromOperationID {
		return Owner{}, fmt.Errorf("operation %s no longer owns %s", fromOperationID, resourceID)
	}
	r.nextGen[resourceID]++
	next := Owner{ResourceID: resourceID, OperationID: toOperationID, Generation: r.nextGen[resourceID], Purpose: purpose, AcquiredAt: time.Now().UTC()}
	r.owners[resourceID] = next
	return next, nil
}

func (r *Registry) Release(resourceID, operationID string, generation uint64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, exists := r.owners[resourceID]
	if !exists || current.OperationID != operationID || current.Generation != generation {
		return false
	}
	delete(r.owners, resourceID)
	return true
}

func (r *Registry) Current(resourceID string) (Owner, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	owner, ok := r.owners[resourceID]
	return owner, ok
}

func (r *Registry) Snapshot() []Owner {
	r.mu.Lock()
	defer r.mu.Unlock()
	owners := make([]Owner, 0, len(r.owners))
	for _, owner := range r.owners {
		owners = append(owners, owner)
	}
	return owners
}
