package journal

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wyw14/cry-106/internal/model"
)

var ErrClosed = errors.New("journal is closed")

type Store struct {
	mu   sync.Mutex
	path string
	file *os.File
	next uint64
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create journal directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open journal: %w", err)
	}
	store := &Store{path: path, file: file, next: 1}
	events, err := store.readAllUnlocked()
	if err != nil {
		file.Close()
		return nil, err
	}
	for _, event := range events {
		if event.Sequence >= store.next {
			store.next = event.Sequence + 1
		}
	}
	return store, nil
}

func (s *Store) Append(ctx context.Context, event model.Event) (model.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return model.Event{}, err
	}
	if s.file == nil {
		return model.Event{}, ErrClosed
	}
	event.Sequence = s.next
	line, err := json.Marshal(event)
	if err != nil {
		return model.Event{}, fmt.Errorf("encode event: %w", err)
	}
	if _, err := s.file.Write(append(line, '\n')); err != nil {
		return model.Event{}, fmt.Errorf("append journal: %w", err)
	}
	if err := s.file.Sync(); err != nil {
		return model.Event{}, fmt.Errorf("sync journal: %w", err)
	}
	s.next++
	return event, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return nil
	}
	err := s.file.Close()
	s.file = nil
	return err
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) readAllUnlocked() ([]model.Event, error) {
	if _, err := s.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seek journal: %w", err)
	}
	var events []model.Event
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		var event model.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode journal line: %w", err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan journal: %w", err)
	}
	_, err := s.file.Seek(0, 2)
	return events, err
}
