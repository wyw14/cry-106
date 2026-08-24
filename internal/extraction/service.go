package extraction

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/model"
	"github.com/wyw14/cry-106/internal/operation"
)

type Session struct {
	ID              string    `json:"id"`
	GroupID         string    `json:"group_id"`
	Mode            string    `json:"mode"`
	OwnerGeneration uint64    `json:"owner_generation"`
	Active          bool      `json:"active"`
	StartedAt       time.Time `json:"started_at"`
}

type Service struct {
	mu       sync.Mutex
	journal  *journal.SessionStore
	owners   *operation.Registry
	valves   *operation.ValveController
	sessions map[string]Session
}

func NewService(events *journal.Store, owners *operation.Registry, commander *operation.Commander) *Service {
	return &Service{journal: journal.NewSessionStore(events), owners: owners, valves: operation.NewValveController(commander), sessions: make(map[string]Session)}
}

func (s *Service) Start(ctx context.Context, groupID, mode string) (Session, error) {
	sessionID := uuid.NewString()
	owner, err := s.owners.Acquire("borehole:"+groupID, sessionID, "extraction")
	if err != nil {
		return Session{}, err
	}
	session := Session{ID: sessionID, GroupID: groupID, Mode: mode, OwnerGeneration: owner.Generation, Active: true, StartedAt: time.Now().UTC()}
	event, err := model.NewEvent(uuid.NewString(), model.EventExtractionStart, groupID, owner.Generation, session, time.Now())
	if err != nil {
		s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation)
		return Session{}, err
	}
	if _, err := s.journal.AppendStart(ctx, event); err != nil {
		s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation)
		return Session{}, err
	}
	if _, err := s.valves.Open(groupID, owner); err != nil {
		s.retireFailedStart(ctx, session, owner)
		return Session{}, err
	}
	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()
	return session, nil
}

// retireFailedStart invalidates a session whose start event was already
// persisted to the journal but whose valves then failed to open. A failed open
// leaves device state untouched (the commander rejects the command before
// recording anything), so no close command is issued here. The persisted start
// is paired with an extraction.stop event so the next recovery does not
// resurrect an active session bound to a borehole group that no operation
// owns, and ownership is released. Rollback errors are swallowed so the caller
// surfaces the original failure rather than a partial-rollback artifact.
func (s *Service) retireFailedStart(ctx context.Context, session Session, owner operation.Owner) {
	stopEvent, err := model.NewEvent(uuid.NewString(), model.EventExtractionStop, session.GroupID, owner.Generation, stopPayload{SessionID: session.ID}, time.Now())
	if err == nil {
		_, _ = s.journal.AppendStop(ctx, stopEvent)
	}
	s.owners.Release(owner.ResourceID, owner.OperationID, owner.Generation)
}

type stopPayload struct {
	SessionID string `json:"session_id"`
}

func (s *Service) Sessions() []Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		result = append(result, session)
	}
	return result
}

func (s *Service) Session(id string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	return session, ok
}
