package verifycase

import (
	"testing"
	"time"

	"github.com/wyw14/cry-106/internal/app"
)

func TestAirwayPlanUsesOneTopologyRevision(t *testing.T) {
	system, err := app.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer system.Close()
	next := system.Topology()
	next.Revision++
	next.Sealed["east"] = true
	delete(next.Edges, "e2")
	done := make(chan error, 1)
	go func() {
		_, err := system.PublishTopology(next)
		done <- err
	}()
	deadline := time.After(300 * time.Millisecond)
	mixed := false
	for {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
			if mixed {
				t.Fatal("reader observed new sealed nodes with the retired edge set")
			}
			return
		case <-deadline:
			t.Fatal("topology publication did not complete")
		default:
			snapshot := system.Topology()
			if snapshot.Revision == next.Revision && snapshot.Sealed["east"] {
				if _, exists := snapshot.Edges["e2"]; exists {
					mixed = true
				}
			}
		}
	}
}
