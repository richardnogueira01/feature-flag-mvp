package control

import (
	"encoding/json"
	"errors"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
	"sort"
	"strings"
	"sync"
)

var (
	ErrInvalidKey   = errors.New("key must contain between 1 and 128 characters")
	ErrExists       = errors.New("flag already exists")
	ErrNotFound     = errors.New("flag not found")
	ErrInvalidValue = errors.New("value must be valid non-null JSON")
)

type Flag struct {
	Key      string          `json:"Key"`
	Enabled  bool            `json:"Enabled"`
	Value    json.RawMessage `json:"Value,omitempty"`
	Revision uint64          `json:"Revision"`
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
	value, _ := json.Marshal(enabled)
	return s.CreateValue(key, value)
}
func (s *Service) Update(key string, enabled bool) (Flag, error) {
	value, _ := json.Marshal(enabled)
	return s.UpdateValue(key, value)
}
func (s *Service) CreateValue(key string, value json.RawMessage) (Flag, error) {
	return s.mutate(key, value, true)
}
func (s *Service) UpdateValue(key string, value json.RawMessage) (Flag, error) {
	return s.mutate(key, value, false)
}
func (s *Service) mutate(key string, value json.RawMessage, create bool) (Flag, error) {
	if err := validateKey(key); err != nil {
		return Flag{}, err
	}
	if err := validateValue(value); err != nil {
		return Flag{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.flags[key]
	if create && exists {
		return Flag{}, ErrExists
	}
	if !create && !exists {
		return Flag{}, ErrNotFound
	}
	s.revision++
	var enabled bool
	_ = json.Unmarshal(value, &enabled)
	flag := Flag{Key: key, Enabled: enabled, Value: append(json.RawMessage(nil), value...), Revision: s.revision}
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
	if _, ok := s.flags[key]; !ok {
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
	f, ok := s.flags[key]
	if !ok {
		return Flag{}, ErrNotFound
	}
	return f, nil
}
func (s *Service) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
func validateKey(key string) error {
	if key == "" || len([]rune(key)) > 128 || strings.TrimSpace(key) != key {
		return ErrInvalidKey
	}
	return nil
}
func validateValue(value json.RawMessage) error {
	if len(value) == 0 || string(value) == "null" {
		return ErrInvalidValue
	}
	var v any
	if json.Unmarshal(value, &v) != nil {
		return ErrInvalidValue
	}
	return nil
}
func (s *Service) publishLocked() {
	if s.data == nil {
		return
	}
	flags := make(map[string]snapshot.Flag, len(s.flags))
	for k, f := range s.flags {
		flags[k] = snapshot.Flag{Key: k, Enabled: f.Enabled, Value: append(json.RawMessage(nil), f.Value...)}
	}
	s.data.Publish(&snapshot.Snapshot{Revision: s.revision, Flags: flags})
}
