package extraction

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-106/internal/operation"
)

type Handover struct {
	owners *operation.Registry
}

func NewHandover(owners *operation.Registry) *Handover {
	return &Handover{owners: owners}
}

func (h *Handover) Commit(current Session, successorID, mode string) (Session, error) {
	if !current.Active {
		return Session{}, fmt.Errorf("session %s is not active", current.ID)
	}
	owner, err := h.owners.Transfer("borehole:"+current.GroupID, current.ID, successorID, "extraction")
	if err != nil {
		return Session{}, err
	}
	return Session{ID: successorID, GroupID: current.GroupID, Mode: mode, OwnerGeneration: owner.Generation, Active: true, StartedAt: time.Now().UTC()}, nil
}
