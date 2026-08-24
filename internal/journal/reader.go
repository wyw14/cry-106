package journal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wyw14/cry-106/internal/model"
)

func ReadAll(path string) ([]model.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Event{}, nil
		}
		return nil, fmt.Errorf("open event log: %w", err)
	}
	defer file.Close()
	return Decode(file)
}

func Decode(file *os.File) ([]model.Event, error) {
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	events := make([]model.Event, 0, 64)
	var previous uint64
	for scanner.Scan() {
		var event model.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode event stream: %w", err)
		}
		if event.Sequence <= previous {
			return nil, fmt.Errorf("event sequence %d does not follow %d", event.Sequence, previous)
		}
		if err := event.Validate(); err != nil {
			return nil, fmt.Errorf("validate event %d: %w", event.Sequence, err)
		}
		previous = event.Sequence
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan event stream: %w", err)
	}
	return events, nil
}
