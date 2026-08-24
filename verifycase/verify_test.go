package verifycase

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-106/internal/app"
)

func TestEmergencyDoorsOutliveFanStopTimeout(t *testing.T) {
	system, err := app.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer system.Close()
	system.SetFanStopDelay("aux-west", 50*time.Millisecond)
	result, err := system.Isolate(context.Background(), "east", "aux-west", "east", 5*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Sealed || result.FanError == "" {
		t.Fatalf("isolation result does not preserve both outcomes: %+v", result)
	}
	states := system.DoorStates()
	if len(states) != 2 {
		t.Fatalf("expected two confirmed emergency doors, got %+v", states)
	}
	for _, state := range states {
		if state.Position != "closed" {
			t.Fatalf("door was not sealed: %+v", state)
		}
	}
}
