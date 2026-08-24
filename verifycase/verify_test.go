package verifycase

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wyw14/cry-106/internal/api"
	"github.com/wyw14/cry-106/internal/app"
)

func TestGasAlarmPublicationDoesNotBlockZoneReads(t *testing.T) {
	system, err := app.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer system.Close()
	request := httptest.NewRequest(http.MethodPost, "/api/gas-zones/east-gas-1/samples", bytes.NewBufferString(`{"zone_id":"east","sequence":1,"calibration_revision":1,"methane_percent":1.8,"observed_at":"2026-08-24T10:00:00Z","received_at":"2026-08-24T10:00:01Z"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		api.NewServer(system).Router().ServeHTTP(response, request)
		close(done)
	}()
	select {
	case <-done:
		if response.Code != http.StatusAccepted {
			t.Fatalf("gas alarm request returned %d: %s", response.Code, response.Body.String())
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("gas alarm publication blocked the zone read path")
	}
	if len(system.AlarmEvents()) != 1 {
		t.Fatalf("alarm listener did not publish one zone snapshot: %+v", system.AlarmEvents())
	}
}
