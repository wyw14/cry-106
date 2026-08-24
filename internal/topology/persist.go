package topology

import (
	"fmt"
	"path/filepath"

	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/model"
)

type SnapshotStore struct {
	path string
}

func NewSnapshotStore(dataDir string) *SnapshotStore {
	return &SnapshotStore{path: filepath.Join(dataDir, "topology.json")}
}

func (s *SnapshotStore) Save(snapshot model.TopologySnapshot) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	if err := journal.WriteSnapshot(s.path, journal.Snapshot[model.TopologySnapshot]{Sequence: snapshot.Revision, Value: snapshot}); err != nil {
		return fmt.Errorf("save topology: %w", err)
	}
	return nil
}

func (s *SnapshotStore) Load(fallback model.TopologySnapshot) (model.TopologySnapshot, error) {
	snapshot, err := journal.ReadSnapshot[model.TopologySnapshot](s.path)
	if err != nil {
		return fallback, nil
	}
	if snapshot.Sequence != snapshot.Value.Revision {
		return model.TopologySnapshot{}, fmt.Errorf("topology snapshot sequence does not match revision")
	}
	return snapshot.Value, snapshot.Value.Validate()
}
