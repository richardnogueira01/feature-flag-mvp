package persistence

import (
	"context"
	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

func (a *ControlAdapter) SetEnabled(c context.Context, k string, e bool) (control.Flag, error) {
	f, x := a.store.SetEnabled(c, k, e)
	return control.Flag{Key: f.Key, Enabled: f.Enabled, Value: f.Value, Revision: f.Revision}, x
}
