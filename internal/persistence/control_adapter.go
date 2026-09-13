package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

type ControlAdapter struct{ store *Store }

func NewControlAdapter(s *Store) *ControlAdapter { return &ControlAdapter{store: s} }
func (a *ControlAdapter) Apply(c context.Context, k string, e bool) (control.Flag, error) {
	v, _ := json.Marshal(e)
	return a.ApplyValue(c, k, v)
}
func (a *ControlAdapter) ApplyValue(c context.Context, k string, v json.RawMessage) (control.Flag, error) {
	f, e := a.store.ApplyValue(c, k, v)
	return control.Flag{Key: f.Key, Enabled: f.Enabled, Value: f.Value, Revision: f.Revision}, e
}
func (a *ControlAdapter) Delete(c context.Context, k string) error {
	e := a.store.Delete(c, k)
	if errors.Is(e, ErrNotFound) {
		return control.ErrNotFound
	}
	return e
}
func (a *ControlAdapter) Get(c context.Context, k string) (control.Flag, error) {
	f, e := a.store.Get(c, k)
	if errors.Is(e, ErrNotFound) {
		return control.Flag{}, control.ErrNotFound
	}
	return control.Flag{Key: f.Key, Enabled: f.Enabled, Value: f.Value, Revision: f.Revision}, e
}
func (a *ControlAdapter) List(c context.Context) ([]control.Flag, error) {
	fs, e := a.store.List(c)
	if e != nil {
		return nil, e
	}
	out := make([]control.Flag, len(fs))
	for i, f := range fs {
		out[i] = control.Flag{Key: f.Key, Enabled: f.Enabled, Value: f.Value, Revision: f.Revision}
	}
	return out, nil
}
