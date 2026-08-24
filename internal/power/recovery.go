package power

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/interlock"
	"github.com/wyw14/cry-106/internal/model"
)

type Recovered struct {
	States  map[string]State `json:"states"`
	Ignored []uint64         `json:"ignored_sequences"`
}

func Recover(events []model.Event, barrier *interlock.PowerBarrier) (Recovered, error) {
	result := Recovered{States: make(map[string]State)}
	for _, event := range events {
		if event.Kind != model.EventPowerTarget && event.Kind != model.EventGasTrip {
			continue
		}
		state, err := model.DecodeEvent[State](event)
		if err != nil {
			return Recovered{}, fmt.Errorf("decode power recovery event: %w", err)
		}
		if event.Kind == model.EventGasTrip {
			barrier.ApplyTrip(event.ZoneID, event.Sequence)
			state.Energized = false
			result.States[event.ZoneID] = state
			continue
		}
		if state.Energized && !barrier.ApplyRestore(event.ZoneID, event.Sequence) {
			result.Ignored = append(result.Ignored, event.Sequence)
			continue
		}
		current := result.States[event.ZoneID]
		if current.Generation <= state.Generation {
			result.States[event.ZoneID] = state
		}
	}
	return result, nil
}
