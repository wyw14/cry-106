package verifycase

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/model"
	"github.com/wyw14/cry-106/internal/sensor"
)

func TestFailedGasFrameCanReplayAfterRestart(t *testing.T) {
	directory := t.TempDir()
	store, err := journal.Open(filepath.Join(directory, "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	cursorPath := filepath.Join(directory, "cursor.json")
	cursor, err := journal.OpenCursor(cursorPath)
	if err != nil {
		t.Fatal(err)
	}
	receiver := sensor.NewReceiver(cursor, journal.NewSampleStore(store))
	now := time.Now().UTC()
	sample := model.GasSample{ID: uuid.NewString(), SensorID: "east-gas-1", ZoneID: "east", Sequence: 1, CalibrationRevision: 1, MethanePercent: 0.8, ObservedAt: now, ReceivedAt: now}
	if err := receiver.Accept(context.Background(), sample); err == nil {
		t.Fatal("closed sample journal unexpectedly accepted frame")
	}
	restarted, err := journal.OpenCursor(cursorPath)
	if err != nil {
		t.Fatal(err)
	}
	if next := restarted.Next(sample.SensorID); next != sample.Sequence {
		t.Fatalf("failed frame was skipped after restart: next=%d frame=%d", next, sample.Sequence)
	}
}
