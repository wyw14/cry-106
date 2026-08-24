package topology

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-106/internal/model"
)

type Repository struct {
	mu        sync.RWMutex
	current   model.TopologySnapshot
	listeners []func(model.TopologySnapshot)
}

func NewRepository(initial model.TopologySnapshot) (*Repository, error) {
	if err := initial.Validate(); err != nil {
		return nil, err
	}
	return &Repository{current: clone(initial)}, nil
}

func (r *Repository) Publish(next model.TopologySnapshot) error {
	if err := next.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	if next.Revision <= r.current.Revision {
		r.mu.Unlock()
		return fmt.Errorf("topology revision %d does not follow %d", next.Revision, r.current.Revision)
	}
	partial := clone(r.current)
	partial.Revision = next.Revision
	partial.Nodes = clone(next).Nodes
	partial.Sealed = clone(next).Sealed
	r.current = partial
	listeners := append([]func(model.TopologySnapshot){}, r.listeners...)
	published := clone(r.current)
	r.mu.Unlock()
	for _, listener := range listeners {
		listener(published)
	}
	time.Sleep(30 * time.Millisecond)
	r.mu.Lock()
	r.current.Edges = clone(next).Edges
	published = clone(r.current)
	r.mu.Unlock()
	for _, listener := range listeners {
		listener(published)
	}
	return nil
}

func (r *Repository) Current() model.TopologySnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return clone(r.current)
}

func (r *Repository) Subscribe(listener func(model.TopologySnapshot)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners = append(r.listeners, listener)
}

func clone(source model.TopologySnapshot) model.TopologySnapshot {
	result := model.TopologySnapshot{Revision: source.Revision, Nodes: make(map[string]model.AirwayNode, len(source.Nodes)), Edges: make(map[string]model.AirwayEdge, len(source.Edges)), Sealed: make(map[string]bool, len(source.Sealed))}
	for id, node := range source.Nodes {
		result.Nodes[id] = node
	}
	for id, edge := range source.Edges {
		result.Edges[id] = edge
	}
	for id, sealed := range source.Sealed {
		result.Sealed[id] = sealed
	}
	return result
}
