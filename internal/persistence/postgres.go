package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

type Flag struct {
	Key      string
	Enabled  bool
	Revision uint64
}
type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Apply(ctx context.Context, key string, enabled bool) (Flag, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Flag{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var revision uint64
	if err := tx.QueryRow(ctx, `UPDATE revision_counter SET revision = revision + 1 WHERE id = 1 RETURNING revision`).Scan(&revision); err != nil {
		return Flag{}, err
	}
	var flag Flag
	if err := tx.QueryRow(ctx, `INSERT INTO feature_flags (key, enabled, revision) VALUES ($1, $2, $3) ON CONFLICT (key) DO UPDATE SET enabled = EXCLUDED.enabled, revision = EXCLUDED.revision, updated_at = now() RETURNING key, enabled, revision`, key, enabled, revision).Scan(&flag.Key, &flag.Enabled, &flag.Revision); err != nil {
		return Flag{}, err
	}
	if err := insertHistoryAndOutbox(ctx, tx, flag, `upsert`); err != nil {
		return Flag{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Flag{}, err
	}
	return flag, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var revision uint64
	if err := tx.QueryRow(ctx, `UPDATE revision_counter SET revision = revision + 1 WHERE id = 1 RETURNING revision`).Scan(&revision); err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `DELETE FROM feature_flags WHERE key = $1`, key)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New(`flag not found`)
	}
	historyPayload, err := json.Marshal(map[string]any{`key`: key, `revision`: revision})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO flag_history (revision, key, operation, payload) VALUES ($1, $2, 'delete', $3)`, revision, key, historyPayload); err != nil {
		return err
	}
	eventPayload, err := snapshotPayload(ctx, tx, revision)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO outbox_events (revision, event_type, payload) VALUES ($1, 'flag.deleted', $2)`, revision, eventPayload); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func insertHistoryAndOutbox(ctx context.Context, tx pgx.Tx, flag Flag, operation string) error {
	historyPayload, err := json.Marshal(flag)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO flag_history (revision, key, operation, payload) VALUES ($1, $2, $3, $4)`, flag.Revision, flag.Key, operation, historyPayload); err != nil {
		return err
	}
	eventPayload, err := snapshotPayload(ctx, tx, flag.Revision)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox_events (revision, event_type, payload) VALUES ($1, 'flag.updated', $2)`, flag.Revision, eventPayload)
	return err
}

func snapshotPayload(ctx context.Context, tx pgx.Tx, revision uint64) ([]byte, error) {
	rows, err := tx.Query(ctx, `SELECT key, enabled FROM feature_flags ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flags := make(map[string]snapshot.Flag)
	for rows.Next() {
		var key string
		var enabled bool
		if err := rows.Scan(&key, &enabled); err != nil {
			return nil, err
		}
		flags[key] = snapshot.Flag{Key: key, Enabled: enabled}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return json.Marshal(snapshot.Snapshot{Revision: revision, Flags: flags})
}

func (s *Store) Snapshot(ctx context.Context) (*snapshot.Snapshot, error) {
	var revision uint64
	if err := s.pool.QueryRow(ctx, `SELECT revision FROM revision_counter WHERE id = 1`).Scan(&revision); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT key, enabled FROM feature_flags ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flags := make(map[string]snapshot.Flag)
	for rows.Next() {
		var key string
		var enabled bool
		if err := rows.Scan(&key, &enabled); err != nil {
			return nil, err
		}
		flags[key] = snapshot.Flag{Key: key, Enabled: enabled}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &snapshot.Snapshot{Revision: revision, Flags: flags}, nil
}
