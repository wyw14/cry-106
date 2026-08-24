package ventilation

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/model"
)

type Plan struct {
	Revision uint64   `json:"revision"`
	EdgeIDs  []string `json:"edge_ids"`
}

type Planner struct{}

func NewPlanner() *Planner {
	return &Planner{}
}

func (p *Planner) Build(snapshot model.TopologySnapshot) (Plan, error) {
	if err := snapshot.Validate(); err != nil {
		return Plan{}, err
	}
	edges := snapshot.ActiveEdgeIDs()
	if len(edges) == 0 {
		return Plan{}, fmt.Errorf("topology revision %d has no active airflow route", snapshot.Revision)
	}
	plan := Plan{Revision: snapshot.Revision, EdgeIDs: edges}
	return plan, nil
}
