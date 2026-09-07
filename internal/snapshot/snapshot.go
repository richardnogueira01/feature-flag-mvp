package snapshot

import "sync/atomic"

type Flag struct {
	Key     string
	Enabled bool
}

type Snapshot struct {
	Revision uint64
	Flags    map[string]Flag
}

type Store struct {
	current atomic.Pointer[Snapshot]
}

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
		flags[key] = flag
	}
	s.current.Store(&Snapshot{Revision: next.Revision, Flags: flags})
}

func (s *Store) Evaluate(key string) (enabled bool, found bool, revision uint64) {
	current := s.current.Load()
	if current == nil {
		return false, false, 0
	}
	flag, found := current.Flags[key]
	return flag.Enabled, found, current.Revision
}

func (s *Store) Revision() uint64 {
	current := s.current.Load()
	if current == nil {
		return 0
	}
	return current.Revision
}
