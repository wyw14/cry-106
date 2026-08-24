package api

import (
	"net/http"
	"strconv"

	"github.com/wyw14/cry-106/internal/model"
)

func (s *Server) activity(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			writeError(w, http.StatusBadRequest, strconv.ErrSyntax)
			return
		}
		limit = parsed
	}
	writeJSON(w, http.StatusOK, map[string]any{"activity": s.system.RecentActivity(limit, r.URL.Query().Get("zone"))})
}

func (s *Server) overview(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.system.Overview())
}

func (s *Server) topology(w http.ResponseWriter, request *http.Request) {
	from, to := request.URL.Query().Get("from"), request.URL.Query().Get("to")
	var path []string
	var pathError string
	if from != "" || to != "" {
		var err error
		path, err = s.system.TopologyPath(from, to)
		if err != nil {
			pathError = err.Error()
		}
	}
	revision, doors := s.system.DoorGraph()
	writeJSON(w, http.StatusOK, map[string]any{
		"snapshot":            s.system.Topology(),
		"door_graph_revision": revision,
		"door_edges":          doors,
		"path":                path,
		"path_error":          pathError,
	})
}

func (s *Server) publishTopology(w http.ResponseWriter, r *http.Request) {
	var snapshot model.TopologySnapshot
	if err := decodeJSON(w, r, &snapshot); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	plan, err := s.system.PublishTopology(snapshot)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, plan)
}

func (s *Server) powerTrip(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ZoneID string `json:"zone_id"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	state, err := s.system.TripPower(r.Context(), request.ZoneID)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, state)
}

func (s *Server) powerRestore(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ZoneID   string `json:"zone_id"`
		Sequence uint64 `json:"sequence"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	state, err := s.system.RestorePower(r.Context(), request.ZoneID, request.Sequence)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, state)
}
