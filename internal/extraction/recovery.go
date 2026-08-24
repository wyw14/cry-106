package extraction

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/model"
)

func Recover(events []model.Event) (map[string]Session, error) {
	sessions := make(map[string]Session)
	for _, event := range events {
		switch event.Kind {
		case model.EventExtractionStart:
			session, err := model.DecodeEvent[Session](event)
			if err != nil {
				return nil, err
			}
			sessions[session.ID] = session
		case model.EventExtractionStop:
			var stopped struct {
				SessionID string `json:"session_id"`
			}
			stopped, err := model.DecodeEvent[struct {
				SessionID string `json:"session_id"`
			}](event)
			if err != nil {
				return nil, err
			}
			session, exists := sessions[stopped.SessionID]
			if !exists {
				return nil, fmt.Errorf("stop references unknown extraction session %s", stopped.SessionID)
			}
			session.Active = false
			sessions[session.ID] = session
		}
	}
	return sessions, nil
}
