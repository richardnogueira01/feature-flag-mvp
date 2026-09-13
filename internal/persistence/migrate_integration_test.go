package persistence

import (
	"context"
	"os"
	"testing"
)

// TestMigrateBackfillsLegacyConfiguration exercises the real migration against a live
// PostgreSQL instance. It seeds a pre-013 schema (no configuration hierarchy) with an existing
// flag, runs Migrate, and checks that the flag is backfilled into an implicit legacy/default
// configuration without data loss — the scenario a real upgrade goes through.
//
// Skipped unless TEST_DATABASE_URL points at a scratch PostgreSQL database (never a real one).
func TestMigrateBackfillsLegacyConfiguration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping real-PostgreSQL migration test")
	}
	ctx := context.Background()
	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Reset to a clean slate, then recreate the pre-013 shape by hand (schema + first table
	// migration only) so this test proves the upgrade path, not just a fresh install.
	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS feature_flags CASCADE`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	preMigration := `
CREATE SCHEMA IF NOT EXISTS feature_flags;
CREATE TABLE revision_counter (id integer PRIMARY KEY CHECK(id=1),revision bigint NOT NULL);
INSERT INTO revision_counter(id,revision) VALUES(1,0);
CREATE TABLE feature_flags (key text PRIMARY KEY CHECK(length(key) BETWEEN 1 AND 128),enabled boolean NOT NULL,value jsonb NOT NULL DEFAULT 'false'::jsonb,revision bigint NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE flag_history(id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,revision bigint NOT NULL,key text NOT NULL,operation text NOT NULL,payload jsonb NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE outbox_events(id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,revision bigint NOT NULL,event_type text NOT NULL,payload jsonb NOT NULL,attempts integer NOT NULL DEFAULT 0,published_at timestamptz,created_at timestamptz NOT NULL DEFAULT now());
INSERT INTO feature_flags(key,enabled,value,revision) VALUES ('legacy-flag', true, 'true'::jsonb, 1);
`
	if _, err := pool.Exec(ctx, preMigration); err != nil {
		t.Fatalf("seed pre-013 state: %v", err)
	}

	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Deliberately acquire a *different* connection than the one Migrate() and the seed step
	// used, to prove the fix in NewPool (AfterConnect) — not just that the lucky first
	// connection happens to have the right search_path.
	var total, withoutScope int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM feature_flags`).Scan(&total); err != nil {
		t.Fatalf("count flags: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected the pre-existing flag to survive the migration, got %d rows", total)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM feature_flags WHERE configuration_id IS NULL`).Scan(&withoutScope); err != nil {
		t.Fatalf("count unscoped flags: %v", err)
	}
	if withoutScope != 0 {
		t.Fatalf("expected every flag to be backfilled into a configuration, got %d without one", withoutScope)
	}

	var app, env, cfg string
	err = pool.QueryRow(ctx, `
		SELECT a.name, e.name, c.name
		FROM feature_flags f
		JOIN configurations c ON c.id = f.configuration_id
		JOIN environments e ON e.id = c.environment_id
		JOIN applications a ON a.id = e.application_id
		WHERE f.key = 'legacy-flag'
	`).Scan(&app, &env, &cfg)
	if err != nil {
		t.Fatalf("resolve backfilled scope: %v", err)
	}
	if app != "legacy" || env != "default" || cfg != "default" {
		t.Fatalf("expected legacy/default/default, got %s/%s/%s", app, env, cfg)
	}

	// Re-running the migration must stay a no-op (idempotent), matching how AUTO_MIGRATE runs it
	// on every startup.
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("second Migrate call must be idempotent: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("third Migrate call must be idempotent: %v", err)
	}
}
