package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("flag not found")

func (s *Store) Get(c context.Context, k string) (Flag, error) {
	var f Flag
	var v []byte
	e := s.pool.QueryRow(c, `SELECT key, enabled, value, revision FROM feature_flags WHERE key=$1`, k).Scan(&f.Key, &f.Enabled, &v, &f.Revision)
	if errors.Is(e, pgx.ErrNoRows) {
		return Flag{}, ErrNotFound
	}
	f.Value = json.RawMessage(v)
	return f, e
}
func (s *Store) List(c context.Context) ([]Flag, error) {
	rows, e := s.pool.Query(c, `SELECT key, enabled, value, revision FROM feature_flags ORDER BY key`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Flag
	for rows.Next() {
		var f Flag
		var v []byte
		if e = rows.Scan(&f.Key, &f.Enabled, &v, &f.Revision); e != nil {
			return nil, e
		}
		f.Value = json.RawMessage(v)
		out = append(out, f)
	}
	return out, rows.Err()
}
