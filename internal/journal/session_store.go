package journal

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-106/internal/model"
)

type SessionStore struct {
	events *Store
}

func NewSessionStore(events *Store) *SessionStore {
	return &SessionStore{events: events}
}

func (s *SessionStore) AppendStart(ctx context.Context, event model.Event) (model.Event, error) {
	if event.Kind != model.EventExtractionStart {
		return model.Event{}, fmt.Errorf("session start store rejected event kind %s", event.Kind)
	}
	stored, err := s.events.Append(ctx, event)
	if err != nil {
		return model.Event{}, fmt.Errorf("append extraction session failed: %w", err)
	}
	return stored, nil
}

func (s *SessionStore) AppendStop(ctx context.Context, event model.Event) (model.Event, error) {
	if event.Kind != model.EventExtractionStop {
		return model.Event{}, fmt.Errorf("session stop store rejected event kind %s", event.Kind)
	}
	stored, err := s.events.Append(ctx, event)
	if err != nil {
		return model.Event{}, fmt.Errorf("append extraction stop failed: %w", err)
	}
	return stored, nil
}
