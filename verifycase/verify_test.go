package verifycase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/app"
	"github.com/wyw14/cry-106/internal/model"
)

func TestOldAirClearanceCannotReleaseNewEvacuation(t *testing.T) {
	system, err := app.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer system.Close()
	first, err := system.BeginEvacuation(context.Background(), "north", 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := system.AccountEvacuation(context.Background(), "north", first.ID, 1); err != nil {
		t.Fatal(err)
	}
	second, err := system.BeginEvacuation(context.Background(), "north", 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := system.AccountEvacuation(context.Background(), "north", second.ID, 1); err != nil {
		t.Fatal(err)
	}
	clearance := model.AirClearance{ID: uuid.NewString(), ZoneID: "north", EvacuationID: first.ID, Approved: true, MethanePPM: 200, ObservedAt: time.Now().UTC()}
	if round, err := system.ApplyClearance(context.Background(), clearance); err == nil {
		t.Fatalf("retired clearance released active round: %+v", round)
	}
	if system.AccessAllowed("north") {
		t.Fatal("retired clearance reopened access for successor evacuation")
	}
}
