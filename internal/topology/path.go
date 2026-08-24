package topology

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/model"
)

func FindPath(snapshot model.TopologySnapshot, from, to string) ([]string, error) {
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}
	if snapshot.Sealed[from] || snapshot.Sealed[to] {
		return nil, fmt.Errorf("path endpoint is sealed")
	}
	type step struct {
		node string
		path []string
	}
	queue := []step{{node: from}}
	seen := map[string]bool{from: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.node == to {
			return current.path, nil
		}
		for id, edge := range snapshot.Edges {
			if edge.From != current.node || snapshot.Sealed[edge.To] || seen[edge.To] {
				continue
			}
			seen[edge.To] = true
			nextPath := append(append([]string(nil), current.path...), id)
			queue = append(queue, step{node: edge.To, path: nextPath})
		}
	}
	return nil, fmt.Errorf("no active route from %s to %s", from, to)
}

func DefaultSnapshot() model.TopologySnapshot {
	return model.TopologySnapshot{
		Revision: 1,
		Nodes: map[string]model.AirwayNode{
			"intake": {ID: "intake", ZoneID: "surface", Name: "Main intake"},
			"east":   {ID: "east", ZoneID: "east", Name: "East wing"},
			"north":  {ID: "north", ZoneID: "north", Name: "North wing"},
			"return": {ID: "return", ZoneID: "surface", Name: "Return shaft"},
		},
		Edges: map[string]model.AirwayEdge{
			"e1": {ID: "e1", From: "intake", To: "east", DoorID: "D1", Capacity: 120},
			"e2": {ID: "e2", From: "east", To: "north", DoorID: "D2", Capacity: 80},
			"e3": {ID: "e3", From: "north", To: "return", DoorID: "D3", Capacity: 110},
		},
		Sealed: map[string]bool{},
	}
}
