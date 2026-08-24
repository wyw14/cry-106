package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/app"
	"github.com/wyw14/cry-106/internal/model"
	"github.com/wyw14/cry-106/internal/ventilation"
)

type Server struct {
	system *app.System
	router chi.Router
}

func NewServer(system *app.System) *Server {
	s := &Server{system: system}
	r := chi.NewRouter()
	r.Get("/healthz", s.health)
	r.Get("/api/overview", s.overview)
	r.Get("/api/activity", s.activity)
	r.Get("/api/topology", s.topology)
	r.Put("/api/topology", s.publishTopology)
	r.Get("/api/ventilation", s.ventilation)
	r.Post("/api/ventilation/operations", s.startFanOperation)
	r.Post("/api/ventilation/operations/{operation}/complete", s.completeFanOperation)
	r.Get("/api/gas-zones", s.gasZones)
	r.Post("/api/gas-zones/{sensor}/samples", s.submitSample)
	r.Get("/api/extraction", s.extraction)
	r.Post("/api/extraction/sessions", s.startExtraction)
	r.Get("/api/incidents", s.incidents)
	r.Post("/api/power/trip", s.powerTrip)
	r.Post("/api/power/restore", s.powerRestore)
	r.Post("/api/incidents/evacuations", s.startEvacuation)
	r.Post("/api/incidents/evacuations/{zone}/account", s.accountEvacuation)
	r.Post("/api/incidents/clearance", s.applyClearance)
	r.Mount("/", consoleRoutes())
	s.router = r
	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) ListenAndServe(ctx context.Context, address string) error {
	server := &http.Server{Addr: address, Handler: s.router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.system.Health())
}

func (s *Server) ventilation(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"targets": s.system.FanTargets(), "active_operations": s.system.ActiveFanOperations(), "commands": s.system.Commands(), "devices": s.system.DeviceStatuses(), "topology": s.system.Topology()})
}

func (s *Server) completeFanOperation(w http.ResponseWriter, r *http.Request) {
	operationID := chi.URLParam(r, "operation")
	if err := s.system.CompleteFanOperation(operationID); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"operation_id": operationID, "status": "complete"})
}

func (s *Server) startFanOperation(w http.ResponseWriter, r *http.Request) {
	var request ventilation.TransferRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	operationID, err := s.system.StartFanOperation(r.Context(), request)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"operation_id": operationID})
}

func (s *Server) gasZones(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"zones": s.system.GasZones(), "epochs": s.system.GasEpochs(), "epoch_history": s.system.GasEpochHistory(), "alarms": s.system.Alarms(), "events": s.system.AlarmEvents()})
}

func (s *Server) submitSample(w http.ResponseWriter, r *http.Request) {
	sensorID := chi.URLParam(r, "sensor")
	var sample model.GasSample
	if err := decodeJSON(w, r, &sample); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	sample.SensorID = sensorID
	value, err := s.system.SubmitGas(r.Context(), sample)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(w, http.StatusAccepted, value)
}

func (s *Server) extraction(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"sessions": s.system.ExtractionSessions(), "commands": s.system.Commands()})
}

func (s *Server) startExtraction(w http.ResponseWriter, r *http.Request) {
	var request struct {
		GroupID string `json:"group_id"`
		Mode    string `json:"mode"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session, err := s.system.StartExtraction(r.Context(), request.GroupID, request.Mode)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, session)
}

func (s *Server) incidents(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"evacuations": s.system.Evacuations(), "door_states": s.system.DoorStates()})
}

func (s *Server) startEvacuation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ZoneID   string `json:"zone_id"`
		Expected int    `json:"expected"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	round, err := s.system.BeginEvacuation(r.Context(), request.ZoneID, request.Expected)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, round)
}

func (s *Server) accountEvacuation(w http.ResponseWriter, r *http.Request) {
	zoneID := chi.URLParam(r, "zone")
	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var request struct {
		RoundID string `json:"round_id"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	round, err := s.system.AccountEvacuation(r.Context(), zoneID, request.RoundID, count)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, round)
}

func (s *Server) applyClearance(w http.ResponseWriter, r *http.Request) {
	var clearance model.AirClearance
	if err := decodeJSON(w, r, &clearance); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if clearance.ID == "" {
		clearance.ID = uuid.NewString()
	}
	if clearance.ObservedAt.IsZero() {
		clearance.ObservedAt = time.Now().UTC()
	}
	round, err := s.system.ApplyClearance(r.Context(), clearance)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, round)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
