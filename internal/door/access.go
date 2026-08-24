package door

import "sync"

type AccessController struct {
	mu      sync.Mutex
	blocked map[string]string
}

func NewAccessController() *AccessController {
	return &AccessController{blocked: make(map[string]string)}
}

func (a *AccessController) Block(zoneID, evacuationID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.blocked[zoneID] = evacuationID
}

func (a *AccessController) Release(zoneID, evacuationID string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.blocked[zoneID] != evacuationID {
		return false
	}
	delete(a.blocked, zoneID)
	return true
}

func (a *AccessController) Allowed(zoneID string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.blocked[zoneID] == ""
}
