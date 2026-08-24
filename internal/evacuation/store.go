package evacuation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/model"
)

type Store struct {
	events *journal.Store
}

func NewStore(events *journal.Store) *Store {
	return &Store{events: events}
}

func (s *Store) Record(ctx context.Context, round Round) (model.Event, error) {
	event, err := model.NewEvent(uuid.NewString(), model.EventEvacuation, round.ZoneID, round.Generation, round, time.Now())
	if err != nil {
		return model.Event{}, err
	}
	stored, err := s.events.Append(ctx, event)
	if err != nil {
		return model.Event{}, fmt.Errorf("append evacuation state: %w", err)
	}
	return stored, nil
}

func Recover(events []model.Event) (map[string]Round, error) {
	result := make(map[string]Round)
	for _, event := range events {
		if event.Kind != model.EventEvacuation {
			continue
		}
		round, err := model.DecodeEvent[Round](event)
		if err != nil {
			return nil, err
		}
		current := result[round.ZoneID]
		if round.Generation >= current.Generation {
			result[round.ZoneID] = round
		}
	}
	return result, nil
}
