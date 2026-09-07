package syncer

func (s *Syncer) Revision() uint64 {
	return s.store.Revision()
}

func (s *Syncer) SetObserver(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observer = observer
}
