package syncer

import (
	"context"
	"errors"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
	"sync"
)

var ErrRevisionGap = errors.New("snapshot revision gap")

type Fetcher func(context.Context) (*snapshot.Snapshot, error)
type Observer interface {
	ObserveGap()
	ObserveResync(bool)
}
type Syncer struct {
	mu           sync.Mutex
	store        *snapshot.Store
	fetch        Fetcher
	observer     Observer
	synchronized bool
	resyncing    bool
}

func New(store *snapshot.Store, fetch Fetcher) *Syncer { return NewWithObserver(store, fetch, nil) }
func NewWithObserver(store *snapshot.Store, fetch Fetcher, observer Observer) *Syncer {
	return &Syncer{store: store, fetch: fetch, observer: observer}
}

func (s *Syncer) Initialize(next *snapshot.Snapshot) error {
	if next == nil {
		return errors.New("snapshot is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store.Publish(next)
	s.synchronized = true
	return nil
}
func (s *Syncer) Apply(ctx context.Context, next *snapshot.Snapshot) error {
	if next == nil {
		return errors.New("snapshot is nil")
	}
	s.mu.Lock()
	current := s.store.Revision()
	if next.Revision <= current {
		s.mu.Unlock()
		return nil
	}
	if current != 0 && next.Revision != current+1 {
		if s.observer != nil {
			s.observer.ObserveGap()
		}
		s.mu.Unlock()
		return s.resync(ctx)
	}
	s.store.Publish(next)
	s.synchronized = true
	s.mu.Unlock()
	return nil
}
func (s *Syncer) Resync(ctx context.Context) error { return s.resync(ctx) }
func (s *Syncer) resync(ctx context.Context) error {
	s.mu.Lock()
	if s.resyncing {
		s.mu.Unlock()
		return ErrRevisionGap
	}
	s.resyncing = true
	s.mu.Unlock()
	success := false
	defer func() {
		s.mu.Lock()
		s.resyncing = false
		s.mu.Unlock()
		if s.observer != nil {
			s.observer.ObserveResync(success)
		}
	}()
	if s.fetch == nil {
		return errors.New("snapshot fetcher is not configured")
	}
	next, err := s.fetch(ctx)
	if err != nil {
		return err
	}
	if next == nil {
		return errors.New("snapshot fetcher returned nil")
	}
	s.mu.Lock()
	s.store.Publish(next)
	s.synchronized = true
	s.mu.Unlock()
	success = true
	return nil
}
func (s *Syncer) Ready() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.synchronized }
