package control

import (
	"context"
	"encoding/json"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

type PersistentService struct {
	repo Repository
	data *snapshot.Store
}

func NewPersistentService(r Repository, d *snapshot.Store) *PersistentService {
	return &PersistentService{repo: r, data: d}
}
func (s *PersistentService) Create(k string, e bool) (Flag, error) {
	v, _ := json.Marshal(e)
	return s.CreateValue(k, v)
}
func (s *PersistentService) Update(k string, e bool) (Flag, error) {
	v, _ := json.Marshal(e)
	return s.UpdateValue(k, v)
}
func (s *PersistentService) CreateValue(k string, v json.RawMessage) (Flag, error) {
	if err := validateKey(k); err != nil {
		return Flag{}, err
	}
	if err := validateValue(v); err != nil {
		return Flag{}, err
	}
	if _, e := s.repo.Get(context.Background(), k); e == nil {
		return Flag{}, ErrExists
	} else if e != ErrNotFound {
		return Flag{}, e
	}
	f, e := s.repo.ApplyValue(context.Background(), k, v)
	if e == nil {
		s.refreshSnapshot()
	}
	return f, e
}
func (s *PersistentService) UpdateValue(k string, v json.RawMessage) (Flag, error) {
	if err := validateKey(k); err != nil {
		return Flag{}, err
	}
	if err := validateValue(v); err != nil {
		return Flag{}, err
	}
	if _, e := s.repo.Get(context.Background(), k); e != nil {
		return Flag{}, e
	}
	f, e := s.repo.ApplyValue(context.Background(), k, v)
	if e == nil {
		s.refreshSnapshot()
	}
	return f, e
}
func (s *PersistentService) Delete(k string) error {
	if e := validateKey(k); e != nil {
		return e
	}
	e := s.repo.Delete(context.Background(), k)
	if e == nil {
		s.refreshSnapshot()
	}
	return e
}
func (s *PersistentService) Get(k string) (Flag, error) { return s.repo.Get(context.Background(), k) }
func (s *PersistentService) List() []Flag {
	f, e := s.repo.List(context.Background())
	if e != nil {
		return nil
	}
	return f
}
func (s *PersistentService) refreshSnapshot() {
	if s.data == nil {
		return
	}
	fs, e := s.repo.List(context.Background())
	if e != nil {
		return
	}
	m := make(map[string]snapshot.Flag, len(fs))
	var rev uint64
	for _, f := range fs {
		m[f.Key] = snapshot.Flag{Key: f.Key, Enabled: f.Enabled, Value: f.Value}
		if f.Revision > rev {
			rev = f.Revision
		}
	}
	s.data.Publish(&snapshot.Snapshot{Revision: rev, Flags: m})
}
