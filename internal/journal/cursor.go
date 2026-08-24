package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Cursor struct {
	mu   sync.Mutex
	path string
	next map[string]uint64
}

func OpenCursor(path string) (*Cursor, error) {
	cursor := &Cursor{path: path, next: make(map[string]uint64)}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cursor, nil
		}
		return nil, fmt.Errorf("read cursor: %w", err)
	}
	if err := json.Unmarshal(raw, &cursor.next); err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}
	return cursor, nil
}

func (c *Cursor) Next(stream string) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	next := c.next[stream]
	if next == 0 {
		return 1
	}
	return next
}

func (c *Cursor) Advance(stream string, accepted uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.next[stream]
	if current == 0 {
		current = 1
	}
	if accepted != current {
		return fmt.Errorf("cursor %s expected %d, got %d", stream, current, accepted)
	}
	c.next[stream] = accepted + 1
	return c.persistLocked()
}

func (c *Cursor) persistLocked() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return fmt.Errorf("create cursor directory: %w", err)
	}
	raw, err := json.MarshalIndent(c.next, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cursor: %w", err)
	}
	temporary := c.path + ".new"
	if err := os.WriteFile(temporary, append(raw, '\n'), 0o600); err != nil {
		return fmt.Errorf("write cursor: %w", err)
	}
	if err := os.Rename(temporary, c.path); err != nil {
		return fmt.Errorf("replace cursor: %w", err)
	}
	return nil
}
