package monitor

import (
	"sync"
	"time"
)

type Activity struct {
	ID        string         `json:"id"`
	Category  string         `json:"category"`
	Message   string         `json:"message"`
	ZoneID    string         `json:"zone_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	Details   map[string]any `json:"details,omitempty"`
}

type ActivityLog struct {
	mu       sync.Mutex
	capacity int
	items    []Activity
}

func NewActivityLog(capacity int) *ActivityLog {
	if capacity < 1 {
		capacity = 100
	}
	return &ActivityLog{capacity: capacity, items: make([]Activity, 0, capacity)}
}

func (l *ActivityLog) Add(activity Activity) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now().UTC()
	}
	l.items = append(l.items, activity)
	if len(l.items) > l.capacity {
		copy(l.items, l.items[len(l.items)-l.capacity:])
		l.items = l.items[:l.capacity]
	}
}

func (l *ActivityLog) Recent(limit int) []Activity {
	l.mu.Lock()
	defer l.mu.Unlock()
	if limit < 1 || limit > len(l.items) {
		limit = len(l.items)
	}
	start := len(l.items) - limit
	result := make([]Activity, limit)
	copy(result, l.items[start:])
	return result
}

func (l *ActivityLog) ByZone(zoneID string) []Activity {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]Activity, 0)
	for _, activity := range l.items {
		if activity.ZoneID == zoneID {
			result = append(result, activity)
		}
	}
	return result
}
