package verifycase

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/wyw14/cry-106/internal/extraction"
	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/operation"
)

func TestExtractionValveWaitsForDurableSession(t *testing.T) {
	store, err := journal.Open(filepath.Join(t.TempDir(), "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	commands := operation.NewCommander()
	service := extraction.NewService(store, operation.NewRegistry(), commands)
	if _, err := service.Start(context.Background(), "B7", "temporary"); err == nil {
		t.Fatal("closed journal unexpectedly accepted extraction session")
	}
	if got := commands.Commands(); len(got) != 0 {
		t.Fatalf("journal failure still emitted valve command: %+v", got)
	}
	if got := service.Sessions(); len(got) != 0 {
		t.Fatalf("failed extraction remained active: %+v", got)
	}
}
