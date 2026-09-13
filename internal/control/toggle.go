package control

import "context"

func (s *Service) SetEnabled(k string, e bool) (Flag, error) {
	if x := validateKey(k); x != nil {
		return Flag{}, x
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.flags[k]
	if !ok {
		return Flag{}, ErrNotFound
	}
	s.revision++
	f.Enabled = e
	f.Revision = s.revision
	s.flags[k] = f
	s.publishLocked()
	return f, nil
}
func (s *PersistentService) SetEnabled(k string, e bool) (Flag, error) {
	if x := validateKey(k); x != nil {
		return Flag{}, x
	}
	t, ok := s.repo.(interface {
		SetEnabled(context.Context, string, bool) (Flag, error)
	})
	if !ok {
		return Flag{}, ErrInvalidValue
	}
	f, x := t.SetEnabled(context.Background(), k, e)
	if x == nil {
		s.refreshSnapshot()
	}
	return f, x
}
