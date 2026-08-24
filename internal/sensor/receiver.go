package sensor

import (
	"context"
	"fmt"
	"sync"

	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/model"
)

type Receiver struct {
	mu     sync.Mutex
	cursor *journal.Cursor
	store  *journal.SampleStore
}

func NewReceiver(cursor *journal.Cursor, store *journal.SampleStore) *Receiver {
	return &Receiver{cursor: cursor, store: store}
}

func (r *Receiver) Accept(ctx context.Context, sample model.GasSample) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	next := r.cursor.Next(sample.SensorID)
	if sample.Sequence < next {
		return nil
	}
	if sample.Sequence > next {
		return fmt.Errorf("sensor %s expected frame %d, got %d", sample.SensorID, next, sample.Sequence)
	}
	if err := r.cursor.Advance(sample.SensorID, sample.Sequence); err != nil {
		return fmt.Errorf("advance gas cursor: %w", err)
	}
	if _, err := r.store.Append(ctx, sample); err != nil {
		return err
	}
	return nil
}
