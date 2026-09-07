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

func NewStore(i *Snapshot) *Store {
	s := &Store{}
	if i != nil {
		s.Publish(i)
	}
	return s
}
func (s *Store) Publish(n *Snapshot) {
	if n == nil {
		return
	}
	m := make(map[string]Flag, len(n.Flags))
	for k, f := range n.Flags {
		f.Value = append(json.RawMessage(nil), f.Value...)
		m[k] = f
	}
	s.current.Store(&Snapshot{Revision: n.Revision, Flags: m})
}
func (s *Store) Evaluate(k string) (bool, bool, uint64) {
	c := s.current.Load()
	if c == nil {
		return false, false, 0
	}
	f, ok := c.Flags[k]
	return f.Enabled, ok, c.Revision
}
func (s *Store) EvaluateValue(k string) (json.RawMessage, bool, uint64) {
	c := s.current.Load()
	if c == nil {
		return nil, false, 0
	}
	f, ok := c.Flags[k]
	v := append(json.RawMessage(nil), f.Value...)
	if len(v) == 0 {
		v, _ = json.Marshal(f.Enabled)
	}
	return v, ok, c.Revision
}
func (s *Store) Revision() uint64 {
	c := s.current.Load()
	if c == nil {
		return 0
	}
	return c.Revision
}
