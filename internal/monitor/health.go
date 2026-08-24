package monitor

import (
	"sync"
	"time"
)

type Health struct {
	Status    string            `json:"status"`
	StartedAt time.Time         `json:"started_at"`
	Checks    map[string]string `json:"checks"`
}

type HealthMonitor struct {
	mu        sync.Mutex
	startedAt time.Time
	checks    map[string]string
}

func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{startedAt: time.Now().UTC(), checks: map[string]string{"journal": "ok", "control": "ok", "console": "ok"}}
}

func (m *HealthMonitor) Set(component, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checks[component] = status
}

func (m *HealthMonitor) Snapshot() Health {
	m.mu.Lock()
	defer m.mu.Unlock()
	status := "ok"
	checks := make(map[string]string, len(m.checks))
	for name, value := range m.checks {
		checks[name] = value
		if value != "ok" {
			status = "degraded"
		}
	}
	return Health{Status: status, StartedAt: m.startedAt, Checks: checks}
}
