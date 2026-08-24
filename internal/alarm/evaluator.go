package alarm

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/interlock"
	"github.com/wyw14/cry-106/internal/model"
)

type Severity string

const (
	SeverityNotice   Severity = "notice"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type Alarm struct {
	ID                  string    `json:"id"`
	ZoneID              string    `json:"zone_id"`
	Severity            Severity  `json:"severity"`
	Message             string    `json:"message"`
	Generation          uint64    `json:"generation"`
	CalibrationRevision uint64    `json:"calibration_revision"`
	Active              bool      `json:"active"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Evaluator struct {
	mu     sync.Mutex
	guard  *interlock.GasGuard
	alarms map[string]Alarm
	epochs *GasEpochStore
}

func NewEvaluator(guard *interlock.GasGuard) *Evaluator {
	return &Evaluator{guard: guard, alarms: make(map[string]Alarm), epochs: NewGasEpochStore()}
}

func (e *Evaluator) Apply(sample model.GasSample) (Alarm, bool, error) {
	decision, changed, err := e.guard.Evaluate(sample)
	if err != nil {
		return Alarm{}, false, err
	}
	changed = e.epochs.Record(GasEpoch{ZoneID: decision.ZoneID, CalibrationRevision: decision.CalibrationRevision, Generation: decision.Generation, MethanePercent: decision.MethanePercent, Tripped: decision.Tripped}) && changed
	e.mu.Lock()
	defer e.mu.Unlock()
	current := e.alarms[sample.ZoneID]
	if !changed && current.ID != "" {
		return current, false, nil
	}
	severity := SeverityNotice
	message := "gas concentration normal"
	if decision.Tripped {
		severity = SeverityCritical
		message = "gas concentration exceeded isolation threshold"
	} else if decision.MethanePercent >= 0.8 {
		severity = SeverityWarning
		message = "gas concentration requires attention"
	}
	alarm := Alarm{ID: current.ID, ZoneID: sample.ZoneID, Severity: severity, Message: message, Generation: decision.Generation, CalibrationRevision: decision.CalibrationRevision, Active: severity != SeverityNotice, UpdatedAt: time.Now().UTC()}
	if alarm.ID == "" {
		alarm.ID = uuid.NewString()
	}
	e.alarms[sample.ZoneID] = alarm
	return alarm, true, nil
}

func (e *Evaluator) Epoch(zoneID string) (GasEpoch, bool) {
	return e.epochs.Current(zoneID)
}

func (e *Evaluator) EpochHistory(zoneID string) []GasEpoch {
	return e.epochs.History(zoneID)
}

func (e *Evaluator) List() []Alarm {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make([]Alarm, 0, len(e.alarms))
	for _, item := range e.alarms {
		result = append(result, item)
	}
	return result
}
