package syncer

func (s *Syncer) Revision() uint64 {
	return s.store.Revision()
}
