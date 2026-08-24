package interlock

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-106/internal/model"
)

type GasDecision struct {
	ZoneID              string  `json:"zone_id"`
	CalibrationRevision uint64  `json:"calibration_revision"`
	MethanePercent      float64 `json:"methane_percent"`
	Tripped             bool    `json:"tripped"`
	Generation          uint64  `json:"generation"`
}

type GasGuard struct {
	mu        sync.Mutex
	threshold float64
	states    map[string]GasDecision
}

func NewGasGuard(threshold float64) *GasGuard {
	return &GasGuard{threshold: threshold, states: make(map[string]GasDecision)}
}

func (g *GasGuard) Evaluate(sample model.GasSample) (GasDecision, bool, error) {
	if err := sample.Validate(); err != nil {
		return GasDecision{}, false, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	current, exists := g.states[sample.ZoneID]
	if exists && sample.CalibrationRevision < current.CalibrationRevision {
		return current, false, nil
	}
	next := GasDecision{ZoneID: sample.ZoneID, CalibrationRevision: sample.CalibrationRevision, MethanePercent: sample.MethanePercent, Tripped: sample.MethanePercent >= g.threshold, Generation: current.Generation}
	if !exists || next.Tripped != current.Tripped || next.CalibrationRevision != current.CalibrationRevision {
		next.Generation++
	}
	if next.Generation == 0 {
		return GasDecision{}, false, fmt.Errorf("gas decision generation did not advance")
	}
	g.states[sample.ZoneID] = next
	return next, !exists || next != current, nil
}

func (g *GasGuard) State(zoneID string) (GasDecision, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	state, ok := g.states[zoneID]
	return state, ok
}
