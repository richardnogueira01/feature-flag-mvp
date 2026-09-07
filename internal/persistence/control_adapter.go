package persistence

import (
	"context"
	"errors"

	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

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
	err := a.store.Delete(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return control.ErrNotFound
	}
	return err
}

func (a *ControlAdapter) Get(ctx context.Context, key string) (control.Flag, error) {
	flag, err := a.store.Get(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return control.Flag{}, control.ErrNotFound
	}
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
