package journal

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/wyw14/cry-106/internal/model"
)

type PowerProjection struct {
	ZoneID     string `json:"zone_id"`
	Energized  bool   `json:"energized"`
	Generation uint64 `json:"generation"`
	Reason     string `json:"reason"`
}

type ReplayResult struct {
	Power map[string]PowerProjection `json:"power"`
	Last  uint64                     `json:"last_sequence"`
}

func Replay(events []model.Event) (ReplayResult, error) {
	ordered := append([]model.Event(nil), events...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Sequence < ordered[j].Sequence })
	for index, event := range ordered {
		if event.Sequence == 0 {
			return ReplayResult{}, fmt.Errorf("event at index %d has no committed sequence", index)
		}
		if index > 0 && ordered[index-1].Sequence == event.Sequence {
			return ReplayResult{}, fmt.Errorf("duplicate committed sequence %d", event.Sequence)
		}
	}
	result := ReplayResult{Power: make(map[string]PowerProjection)}
	terminal := make(map[string]uint64)
	for _, event := range ordered {
		if event.Kind != model.EventPowerTarget && event.Kind != model.EventGasTrip {
			continue
		}
		var state PowerProjection
		if err := json.Unmarshal(event.Payload, &state); err != nil {
			return ReplayResult{}, fmt.Errorf("decode power projection: %w", err)
		}
		if event.Kind == model.EventGasTrip {
			terminal[event.ZoneID] = event.Sequence
			state.Energized = false
		}
		if state.Energized && event.Sequence < terminal[event.ZoneID] {
			continue
		}
		current := result.Power[event.ZoneID]
		if current.Generation <= state.Generation || event.Kind == model.EventGasTrip {
			result.Power[event.ZoneID] = state
		}
	}
	if len(ordered) > 0 {
		result.Last = ordered[len(ordered)-1].Sequence
	}
	return result, nil
}
