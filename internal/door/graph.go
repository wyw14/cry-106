package door

import (
	"sync"

	"github.com/wyw14/cry-106/internal/model"
)

type Graph struct {
	mu       sync.RWMutex
	revision uint64
	edges    map[string]model.AirwayEdge
}

func NewGraph() *Graph {
	return &Graph{edges: make(map[string]model.AirwayEdge)}
}

func (g *Graph) Apply(snapshot model.TopologySnapshot) {
	edges := make(map[string]model.AirwayEdge, len(snapshot.Edges))
	for id, edge := range snapshot.Edges {
		edges[id] = edge
	}
	g.mu.Lock()
	g.edges = edges
	g.revision = snapshot.Revision
	g.mu.Unlock()
}

func (g *Graph) Edges() (uint64, []model.AirwayEdge) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]model.AirwayEdge, 0, len(g.edges))
	for _, edge := range g.edges {
		result = append(result, edge)
	}
	return g.revision, result
}
