package persistence

import (
	"context"

	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

// ControlAdapter exposes Store through the control-plane persistence port.
type ControlAdapter struct {
	store *Store
}

func NewControlAdapter(store *Store) *ControlAdapter {
	return &ControlAdapter{store: store}
}

func (a *ControlAdapter) Apply(ctx context.Context, key string, enabled bool) (control.Flag, error) {
	flag, err := a.store.Apply(ctx, key, enabled)
	if err != nil {
		return control.Flag{}, err
	}
	return control.Flag{Key: flag.Key, Enabled: flag.Enabled, Revision: flag.Revision}, nil
}

func (a *ControlAdapter) Delete(ctx context.Context, key string) error {
	return a.store.Delete(ctx, key)
}

func (a *ControlAdapter) Get(ctx context.Context, key string) (control.Flag, error) {
	flag, err := a.store.Get(ctx, key)
	if err != nil {
		return control.Flag{}, err
	}
	return control.Flag{Key: flag.Key, Enabled: flag.Enabled, Revision: flag.Revision}, nil
}

func (a *ControlAdapter) List(ctx context.Context) ([]control.Flag, error) {
	flags, err := a.store.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]control.Flag, len(flags))
	for i, flag := range flags {
		result[i] = control.Flag{Key: flag.Key, Enabled: flag.Enabled, Revision: flag.Revision}
	}
	return result, nil
}

var _ control.Repository = (*ControlAdapter)(nil)
