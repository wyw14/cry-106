package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventKind string

const (
	EventGasSample       EventKind = "gas.sample"
	EventGasTrip         EventKind = "gas.trip"
	EventPowerTarget     EventKind = "power.target"
	EventExtractionStart EventKind = "extraction.start"
	EventExtractionStop  EventKind = "extraction.stop"
	EventEvacuation      EventKind = "evacuation.state"
	EventTopology        EventKind = "topology.revision"
	EventAlarm           EventKind = "alarm.state"
)

type Event struct {
	ID         string          `json:"id"`
	Sequence   uint64          `json:"sequence"`
	Kind       EventKind       `json:"kind"`
	ZoneID     string          `json:"zone_id,omitempty"`
	Generation uint64          `json:"generation,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	Payload    json.RawMessage `json:"payload"`
}

func NewEvent(id string, kind EventKind, zoneID string, generation uint64, payload any, at time.Time) (Event, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("marshal %s event: %w", kind, err)
	}
	event := Event{ID: id, Kind: kind, ZoneID: zoneID, Generation: generation, CreatedAt: at.UTC(), Payload: raw}
	if err := event.Validate(); err != nil {
		return Event{}, err
	}
	return event, nil
}

func (e Event) Validate() error {
	if e.ID == "" || e.Kind == "" || e.CreatedAt.IsZero() {
		return fmt.Errorf("event identity, kind, and timestamp are required")
	}
	if !json.Valid(e.Payload) {
		return fmt.Errorf("event %s payload is not valid JSON", e.ID)
	}
	return nil
}

func DecodeEvent[T any](event Event) (T, error) {
	var value T
	if err := json.Unmarshal(event.Payload, &value); err != nil {
		return value, fmt.Errorf("decode %s event %s: %w", event.Kind, event.ID, err)
	}
	return value, nil
}
