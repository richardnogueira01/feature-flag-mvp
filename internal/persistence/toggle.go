package persistence

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

func (s *Store) SetEnabled(c context.Context, k string, e bool) (control.Flag, error) {
	tx, x := s.pool.BeginTx(c, pgx.TxOptions{})
	if x != nil {
		return control.Flag{}, x
	}
	defer func() { _ = tx.Rollback(c) }()
	var rev uint64
	if x = tx.QueryRow(c, `UPDATE revision_counter SET revision=revision+1 WHERE id=1 RETURNING revision`).Scan(&rev); x != nil {
		return control.Flag{}, x
	}
	var f Flag
	if x = tx.QueryRow(c, `UPDATE feature_flags SET enabled=$2,revision=$3,updated_at=now() WHERE key=$1 RETURNING key,enabled,value,revision`, k, e, rev).Scan(&f.Key, &f.Enabled, &f.Value, &f.Revision); x == pgx.ErrNoRows {
		return control.Flag{}, ErrNotFound
	}
	if x != nil {
		return control.Flag{}, x
	}
	h, _ := json.Marshal(f)
	if _, x = tx.Exec(c, `INSERT INTO flag_history(revision,key,operation,payload) VALUES($1,$2,'toggle',$3)`, rev, k, h); x != nil {
		return control.Flag{}, x
	}
	p, x := snapshotPayload(c, tx, rev)
	if x != nil {
		return control.Flag{}, x
	}
	if _, x = tx.Exec(c, `INSERT INTO outbox_events(revision,event_type,payload) VALUES($1,'flag.updated',$2)`, rev, p); x != nil {
		return control.Flag{}, x
	}
	if x = tx.Commit(c); x != nil {
		return control.Flag{}, x
	}
	return control.Flag{Key: f.Key, Enabled: f.Enabled, Value: f.Value, Revision: f.Revision}, nil
}
