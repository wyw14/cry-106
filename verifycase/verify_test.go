package verifycase

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/extraction"
	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/operation"
)

func TestRetiredExtractionCleanupCannotCloseSuccessor(t *testing.T) {
	store, err := journal.Open(filepath.Join(t.TempDir(), "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	owners := operation.NewRegistry()
	commands := operation.NewCommander()
	service := extraction.NewService(store, owners, commands)
	retired, err := service.Start(context.Background(), "B12", "temporary")
	if err != nil {
		t.Fatal(err)
	}
	handover := extraction.NewHandover(owners)
	successor, err := handover.Commit(retired, uuid.NewString(), "long_running")
	if err != nil {
		t.Fatal(err)
	}
	closed, err := extraction.NewCleanupRunner(owners, commands).CloseGroup(retired)
	if err != nil {
		t.Fatal(err)
	}
	if closed {
		t.Fatalf("retired session closed successor %+v", successor)
	}
	current, ok := owners.Current("borehole:B12")
	if !ok || current.OperationID != successor.ID || current.Generation != successor.OwnerGeneration {
		t.Fatalf("successor ownership was lost: %+v", current)
	}
	issued := commands.Commands()
	if issued[len(issued)-1].Target != "open" {
		t.Fatalf("retired cleanup emitted closing target: %+v", issued)
	}
}
