package model

import (
	"fmt"
	"sort"
)

type AirwayNode struct {
	ID     string `json:"id"`
	ZoneID string `json:"zone_id"`
	Name   string `json:"name"`
}

type AirwayEdge struct {
	ID       string `json:"id"`
	From     string `json:"from"`
	To       string `json:"to"`
	DoorID   string `json:"door_id,omitempty"`
	Capacity int    `json:"capacity"`
}

type TopologySnapshot struct {
	Revision uint64                `json:"revision"`
	Nodes    map[string]AirwayNode `json:"nodes"`
	Edges    map[string]AirwayEdge `json:"edges"`
	Sealed   map[string]bool       `json:"sealed"`
}

func (s TopologySnapshot) Validate() error {
	if s.Revision == 0 {
		return fmt.Errorf("topology revision must be positive")
	}
	for id, edge := range s.Edges {
		if id == "" || edge.From == "" || edge.To == "" {
			return fmt.Errorf("topology contains an incomplete edge")
		}
		if _, ok := s.Nodes[edge.From]; !ok {
			return fmt.Errorf("edge %s references missing source %s", id, edge.From)
		}
		if _, ok := s.Nodes[edge.To]; !ok {
			return fmt.Errorf("edge %s references missing destination %s", id, edge.To)
		}
	}
	return nil
}

func (s TopologySnapshot) ActiveEdgeIDs() []string {
	ids := make([]string, 0, len(s.Edges))
	for id, edge := range s.Edges {
		if !s.Sealed[edge.From] && !s.Sealed[edge.To] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
