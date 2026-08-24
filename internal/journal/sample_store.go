package journal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/model"
)

type SampleStore struct {
	journal *Store
}

func NewSampleStore(store *Store) *SampleStore {
	return &SampleStore{journal: store}
}

func (s *SampleStore) Append(ctx context.Context, sample model.GasSample) (model.Event, error) {
	if err := sample.Validate(); err != nil {
		return model.Event{}, err
	}
	event, err := model.NewEvent(uuid.NewString(), model.EventGasSample, sample.ZoneID, sample.CalibrationRevision, sample, time.Now())
	if err != nil {
		return model.Event{}, err
	}
	stored, err := s.journal.Append(ctx, event)
	if err != nil {
		return model.Event{}, fmt.Errorf("append gas sample failed: %w", err)
	}
	return stored, nil
}

func LoadSamples(path string, sensorID string) ([]model.GasSample, error) {
	events, err := ReadAll(path)
	if err != nil {
		return nil, err
	}
	samples := make([]model.GasSample, 0)
	for _, event := range events {
		if event.Kind != model.EventGasSample {
			continue
		}
		sample, err := model.DecodeEvent[model.GasSample](event)
		if err != nil {
			return nil, err
		}
		if sensorID == "" || sample.SensorID == sensorID {
			samples = append(samples, sample)
		}
	}
	return samples, nil
}
