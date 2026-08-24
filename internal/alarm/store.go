package alarm

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/model"
)

type Store struct {
	journal *journal.Store
}

func NewStore(events *journal.Store) *Store {
	return &Store{journal: events}
}

func (s *Store) Record(ctx context.Context, value Alarm) (model.Event, error) {
	event, err := model.NewEvent(uuid.NewString(), model.EventAlarm, value.ZoneID, value.Generation, value, time.Now())
	if err != nil {
		return model.Event{}, err
	}
	stored, err := s.journal.Append(ctx, event)
	if err != nil {
		return model.Event{}, fmt.Errorf("append alarm state: %w", err)
	}
	return stored, nil
}
