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
	Value    json.RawMessage
	Revision uint64
}
type Store struct{ pool *pgxpool.Pool }

func NewStore(p *pgxpool.Pool) *Store { return &Store{pool: p} }
func (s *Store) Apply(c context.Context, k string, e bool) (Flag, error) {
	v, _ := json.Marshal(e)
	return s.ApplyValue(c, k, v)
}
func (s *Store) ApplyValue(c context.Context, k string, v json.RawMessage) (Flag, error) {
	tx, e := s.pool.BeginTx(c, pgx.TxOptions{})
	if e != nil {
		return Flag{}, e
	}
	defer func() { _ = tx.Rollback(c) }()
	var rev uint64
	if e = tx.QueryRow(c, `UPDATE revision_counter SET revision=revision+1 WHERE id=1 RETURNING revision`).Scan(&rev); e != nil {
		return Flag{}, e
	}
	enabled := true
	var boolValue bool
	if json.Unmarshal(v, &boolValue) == nil {
		enabled = boolValue
	}
	var f Flag
	if e = tx.QueryRow(c, `INSERT INTO feature_flags(key,enabled,value,revision) VALUES($1,$2,$3,$4) ON CONFLICT(key) DO UPDATE SET enabled=EXCLUDED.enabled,value=EXCLUDED.value,revision=EXCLUDED.revision,updated_at=now() RETURNING key,enabled,value,revision`, k, enabled, v, rev).Scan(&f.Key, &f.Enabled, &f.Value, &f.Revision); e != nil {
		return Flag{}, e
	}
	history, _ := json.Marshal(f)
	if _, e = tx.Exec(c, `INSERT INTO flag_history(revision,key,operation,payload) VALUES($1,$2,'upsert',$3)`, rev, k, history); e != nil {
		return Flag{}, e
	}
	payload, e := snapshotPayload(c, tx, rev)
	if e != nil {
		return Flag{}, e
	}
	if _, e = tx.Exec(c, `INSERT INTO outbox_events(revision,event_type,payload) VALUES($1,'flag.updated',$2)`, rev, payload); e != nil {
		return Flag{}, e
	}
	e = tx.Commit(c)
	return f, e
}
func (s *Store) Delete(c context.Context, k string) error {
	tx, e := s.pool.BeginTx(c, pgx.TxOptions{})
	if e != nil {
		return e
	}
	defer func() { _ = tx.Rollback(c) }()
	var rev uint64
	if e = tx.QueryRow(c, `UPDATE revision_counter SET revision=revision+1 WHERE id=1 RETURNING revision`).Scan(&rev); e != nil {
		return e
	}
	r, e := tx.Exec(c, `DELETE FROM feature_flags WHERE key=$1`, k)
	if e != nil {
		return e
	}
	if r.RowsAffected() == 0 {
		return ErrNotFound
	}
	p, _ := json.Marshal(map[string]any{"key": k, "revision": rev})
	if _, e = tx.Exec(c, `INSERT INTO flag_history(revision,key,operation,payload) VALUES($1,$2,'delete',$3)`, rev, k, p); e != nil {
		return e
	}
	sp, e := snapshotPayload(c, tx, rev)
	if e != nil {
		return e
	}
	if _, e = tx.Exec(c, `INSERT INTO outbox_events(revision,event_type,payload) VALUES($1,'flag.deleted',$2)`, rev, sp); e != nil {
		return e
	}
	return tx.Commit(c)
}
func snapshotPayload(c context.Context, tx pgx.Tx, rev uint64) ([]byte, error) {
	rows, e := tx.Query(c, `SELECT key,enabled,value FROM feature_flags ORDER BY key`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	m := map[string]snapshot.Flag{}
	for rows.Next() {
		var k string
		var en bool
		var v []byte
		if e = rows.Scan(&k, &en, &v); e != nil {
			return nil, e
		}
		m[k] = snapshot.Flag{Key: k, Enabled: en, Value: v}
	}
	return json.Marshal(snapshot.Snapshot{Revision: rev, Flags: m})
}
func (s *Store) Snapshot(c context.Context) (*snapshot.Snapshot, error) {
	var rev uint64
	if e := s.pool.QueryRow(c, `SELECT revision FROM revision_counter WHERE id=1`).Scan(&rev); e != nil {
		return nil, e
	}
	rows, e := s.pool.Query(c, `SELECT key,enabled,value FROM feature_flags ORDER BY key`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	m := map[string]snapshot.Flag{}
	for rows.Next() {
		var k string
		var en bool
		var v []byte
		if e = rows.Scan(&k, &en, &v); e != nil {
			return nil, e
		}
		m[k] = snapshot.Flag{Key: k, Enabled: en, Value: v}
	}
	return &snapshot.Snapshot{Revision: rev, Flags: m}, rows.Err()
}

var _ = errors.New
