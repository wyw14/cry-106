package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Snapshot[T any] struct {
	Sequence uint64 `json:"sequence"`
	Value    T      `json:"value"`
}

func WriteSnapshot[T any](path string, snapshot Snapshot[T]) error {
	raw, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create snapshot directory: %w", err)
	}
	temporary := path + ".new"
	if err := os.WriteFile(temporary, append(raw, '\n'), 0o600); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("commit snapshot: %w", err)
	}
	return nil
}

func ReadSnapshot[T any](path string) (Snapshot[T], error) {
	var snapshot Snapshot[T]
	raw, err := os.ReadFile(path)
	if err != nil {
		return snapshot, fmt.Errorf("read snapshot: %w", err)
	}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return snapshot, fmt.Errorf("decode snapshot: %w", err)
	}
	return snapshot, nil
}
