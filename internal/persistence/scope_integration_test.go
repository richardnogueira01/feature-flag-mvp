package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

// TestScopeStoreAgainstRealPostgres exercises every Store method backing control.ScopeService
// against a live PostgreSQL instance: the SQL itself (column counts, ON CONFLICT targets,
// transactions), not just the in-memory fake used by internal/control's unit tests.
//
// Skipped unless TEST_DATABASE_URL points at a scratch PostgreSQL database (never a real one).
func TestScopeStoreAgainstRealPostgres(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping real-PostgreSQL scope test")
	}
	ctx := context.Background()
	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS feature_flags CASCADE`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	store := NewStore(pool)
	service := control.NewScopeService(store)

	if _, err := service.CreateApplication(ctx, "atlas"); err != nil {
		t.Fatalf("create application: %v", err)
	}
	if _, err := service.CreateApplication(ctx, "atlas"); !errors.Is(err, control.ErrScopeExists) {
		t.Fatalf("expected ErrScopeExists on duplicate application, got %v", err)
	}
	if _, err := service.CreateEnvironment(ctx, "atlas", "production"); err != nil {
		t.Fatalf("create environment: %v", err)
	}
	if _, err := service.CreateConfiguration(ctx, "atlas", "production", "mobile-menu"); err != nil {
		t.Fatalf("create configuration: %v", err)
	}
	if _, err := service.CreateConfiguration(ctx, "atlas", "production", "checkout"); err != nil {
		t.Fatalf("create second configuration: %v", err)
	}

	trueValue, _ := json.Marshal(true)
	falseValue, _ := json.Marshal(false)
	if _, err := service.CreateFlag(ctx, "atlas", "production", "mobile-menu", "new-ui", trueValue); err != nil {
		t.Fatalf("create flag in mobile-menu: %v", err)
	}
	if _, err := service.CreateFlag(ctx, "atlas", "production", "checkout", "new-ui", falseValue); err != nil {
		t.Fatalf("create same key in checkout: %v", err)
	}

	a, err := service.GetFlag(ctx, "atlas", "production", "mobile-menu", "new-ui")
	if err != nil || !a.Enabled {
		t.Fatalf("mobile-menu/new-ui = %+v, err=%v", a, err)
	}
	b, err := service.GetFlag(ctx, "atlas", "production", "checkout", "new-ui")
	if err != nil || b.Enabled {
		t.Fatalf("checkout/new-ui = %+v, err=%v", b, err)
	}

	toggled, err := service.SetFlagEnabled(ctx, "atlas", "production", "mobile-menu", "new-ui", false)
	if err != nil || toggled.Enabled || toggled.Revision != 2 {
		t.Fatalf("toggle: flag=%+v err=%v", toggled, err)
	}

	snapshot, err := service.GetConfiguration(ctx, "atlas", "production", "mobile-menu")
	if err != nil || snapshot.Revision != 2 || len(snapshot.Flags) != 1 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}

	if err := service.DeleteFlag(ctx, "atlas", "production", "mobile-menu", "new-ui"); err != nil {
		t.Fatalf("delete flag: %v", err)
	}
	if _, err := service.GetFlag(ctx, "atlas", "production", "mobile-menu", "new-ui"); !errors.Is(err, control.ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound after delete, got %v", err)
	}
	// checkout/new-ui must be untouched by deleting mobile-menu/new-ui.
	if _, err := service.GetFlag(ctx, "atlas", "production", "checkout", "new-ui"); err != nil {
		t.Fatalf("checkout/new-ui should still exist: %v", err)
	}

	if _, err := service.CreateConfiguration(ctx, "atlas", "staging", "mobile-menu"); !errors.Is(err, control.ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound for missing environment, got %v", err)
	}
}
