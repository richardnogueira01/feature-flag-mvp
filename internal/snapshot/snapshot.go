package snapshot

import (
	"encoding/json"
	"sync/atomic"
)

type Flag struct {
	Key     string
	Enabled bool
	Value   json.RawMessage
}
type Snapshot struct {
	Revision uint64
	Flags    map[string]Flag
}
type Store struct{ current atomic.Pointer[Snapshot] }

func NewStore(initial *Snapshot) *Store {
	store := &Store{}
	if initial != nil {
		store.Publish(initial)
	}
	return store
}
func (s *Store) Publish(next *Snapshot) {
	if next == nil {
		return
	}
	flags := make(map[string]Flag, len(next.Flags))
	for key, flag := range next.Flags {
		flag.Value = append(json.RawMessage(nil), flag.Value...)
		flags[key] = flag
	}
	s.current.Store(&Snapshot{Revision: next.Revision, Flags: flags})
}
func (s *Store) Evaluate(key string) (bool, bool, uint64) {
	current := s.current.Load()
	if current == nil {
		return false, false, 0
	}
	flag, found := current.Flags[key]
	if len(flag.Value) > 0 {
		var enabled bool
		if json.Unmarshal(flag.Value, &enabled) == nil {
			return enabled, found, current.Revision
		}
	}
	return flag.Enabled, found, current.Revision
}
func (s *Store) EvaluateValue(key string) (json.RawMessage, bool, uint64) {
	current := s.current.Load()
	if current == nil {
		return nil, false, 0
	}
	flag, found := current.Flags[key]
	value := append(json.RawMessage(nil), flag.Value...)
	if len(value) == 0 {
		value, _ = json.Marshal(flag.Enabled)
	}
	return value, found, current.Revision
}
func (s *Store) Revision() uint64 {
	current := s.current.Load()
	if current == nil {
		return 0
	}
	return current.Revision
}
