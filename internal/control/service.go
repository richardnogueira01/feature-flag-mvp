package control

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

var (
	ErrInvalidKey = errors.New("key must contain between 1 and 128 characters")
	ErrExists     = errors.New("flag already exists")
	ErrNotFound   = errors.New("flag not found")
)

type Flag struct {
	Key      string
	Enabled  bool
	Revision uint64
}

type Service struct {
	mu       sync.RWMutex
	flags    map[string]Flag
	revision uint64
	data     *snapshot.Store
}

func NewService(data *snapshot.Store) *Service {
	return &Service{flags: make(map[string]Flag), data: data}
}

func (s *Service) Create(key string, enabled bool) (Flag, error) {
	if err := validateKey(key); err != nil {
		return Flag{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[key]; exists {
		return Flag{}, ErrExists
	}
	flag := s.nextFlag(key, enabled)
	s.flags[key] = flag
	s.publishLocked()
	return flag, nil
}

func (s *Service) Update(key string, enabled bool) (Flag, error) {
	if err := validateKey(key); err != nil {
		return Flag{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[key]; !exists {
		return Flag{}, ErrNotFound
	}
	flag := s.nextFlag(key, enabled)
	s.flags[key] = flag
	s.publishLocked()
	return flag, nil
}

func (s *Service) Delete(key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[key]; !exists {
		return ErrNotFound
	}
	delete(s.flags, key)
	s.revision++
	s.publishLocked()
	return nil
}

func (s *Service) Get(key string) (Flag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	flag, exists := s.flags[key]
	if !exists {
		return Flag{}, ErrNotFound
	}
	return flag, nil
}

func (s *Service) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	flags := make([]Flag, 0, len(s.flags))
	for _, flag := range s.flags {
		flags = append(flags, flag)
	}
	sort.Slice(flags, func(i, j int) bool { return flags[i].Key < flags[j].Key })
	return flags
}

func validateKey(key string) error {
	if key == "" || len([]rune(key)) > 128 || strings.TrimSpace(key) != key {
		return ErrInvalidKey
	}
	return nil
}

func (s *Service) nextFlag(key string, enabled bool) Flag {
	s.revision++
	return Flag{Key: key, Enabled: enabled, Revision: s.revision}
}

func (s *Service) publishLocked() {
	if s.data == nil {
		return
	}
	flags := make(map[string]snapshot.Flag, len(s.flags))
	for key, flag := range s.flags {
		flags[key] = snapshot.Flag{Key: key, Enabled: flag.Enabled}
	}
	s.data.Publish(&snapshot.Snapshot{Revision: s.revision, Flags: flags})
}
