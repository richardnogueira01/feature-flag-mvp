package control

import (
	"context"

	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

// PersistentService adapts the control-plane operations to a Repository.
// The repository owns the transaction; this service owns domain validation and snapshot refresh.
type PersistentService struct {
	repo Repository
	data *snapshot.Store
}

func NewPersistentService(repo Repository, data *snapshot.Store) *PersistentService {
	return &PersistentService{repo: repo, data: data}
}

func (s *PersistentService) Create(key string, enabled bool) (Flag, error) {
	if err := validateKey(key); err != nil {
		return Flag{}, err
	}
	if _, err := s.repo.Get(context.Background(), key); err == nil {
		return Flag{}, ErrExists
	} else if err != ErrNotFound {
		return Flag{}, err
	}
	flag, err := s.repo.Apply(context.Background(), key, enabled)
	if err != nil {
		return Flag{}, err
	}
	s.refreshSnapshot()
	return flag, nil
}

func (s *PersistentService) Update(key string, enabled bool) (Flag, error) {
	if err := validateKey(key); err != nil {
		return Flag{}, err
	}
	if _, err := s.repo.Get(context.Background(), key); err != nil {
		if err == ErrNotFound {
			return Flag{}, ErrNotFound
		}
		return Flag{}, err
	}
	flag, err := s.repo.Apply(context.Background(), key, enabled)
	if err != nil {
		return Flag{}, err
	}
	s.refreshSnapshot()
	return flag, nil
}

func (s *PersistentService) Delete(key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := s.repo.Delete(context.Background(), key); err != nil {
		return err
	}
	s.refreshSnapshot()
	return nil
}

func (s *PersistentService) Get(key string) (Flag, error) {
	return s.repo.Get(context.Background(), key)
}

func (s *PersistentService) List() []Flag {
	flags, err := s.repo.List(context.Background())
	if err != nil {
		return nil
	}
	return flags
}

func (s *PersistentService) refreshSnapshot() {
	if s.data == nil {
		return
	}
	flags := s.repo.List(context.Background())
	if len(flags) == 0 {
		return
	}
	snapshotFlags := make(map[string]snapshot.Flag, len(flags))
	var revision uint64
	for _, flag := range flags {
		snapshotFlags[flag.Key] = snapshot.Flag{Key: flag.Key, Enabled: flag.Enabled}
		if flag.Revision > revision {
			revision = flag.Revision
		}
	}
	s.data.Publish(&snapshot.Snapshot{Revision: revision, Flags: snapshotFlags})
}
