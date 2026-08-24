package verifycase

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/wyw14/cry-106/internal/api"
	"github.com/wyw14/cry-106/internal/app"
)

func TestFanTransferAndReversalHaveOneOwner(t *testing.T) {
	system, err := app.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer system.Close()
	server := httptest.NewServer(api.NewServer(system).Router())
	defer server.Close()
	bodies := []string{
		`{"group_id":"main","target_id":"V2","mode":"transfer"}`,
		`{"group_id":"main","mode":"reverse"}`,
	}
	start := make(chan struct{})
	statuses := make(chan int, 2)
	var wait sync.WaitGroup
	for _, body := range bodies {
		wait.Add(1)
		go func(payload string) {
			defer wait.Done()
			<-start
			response, err := http.Post(server.URL+"/api/ventilation/operations", "application/json", bytes.NewBufferString(payload))
			if err != nil {
				statuses <- 0
				return
			}
			defer response.Body.Close()
			statuses <- response.StatusCode
		}(body)
	}
	close(start)
	wait.Wait()
	close(statuses)
	accepted, rejected := 0, 0
	for status := range statuses {
		switch status {
		case http.StatusAccepted:
			accepted++
		case http.StatusConflict:
			rejected++
		}
	}
	if accepted != 1 || rejected != 1 {
		t.Fatalf("accepted=%d rejected=%d commands=%v", accepted, rejected, system.Commands())
	}
	if len(system.Commands()) != 1 {
		t.Fatalf("conflicting operations emitted %d device commands", len(system.Commands()))
	}
}
