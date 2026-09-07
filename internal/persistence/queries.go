package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("flag not found")

func (s *Store) Get(ctx context.Context, key string) (Flag, error) {
	var flag Flag
	err := s.pool.QueryRow(ctx, `SELECT key, enabled, revision FROM feature_flags WHERE key = $1`, key).
		Scan(&flag.Key, &flag.Enabled, &flag.Revision)
	if errors.Is(err, pgx.ErrNoRows) {
		return Flag{}, ErrNotFound
	}
	if err != nil {
		return Flag{}, err
	}
	return flag, nil
}

func (s *Store) List(ctx context.Context) ([]Flag, error) {
	rows, err := s.pool.Query(ctx, `SELECT key, enabled, revision FROM feature_flags ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []Flag
	for rows.Next() {
		var flag Flag
		if err := rows.Scan(&flag.Key, &flag.Enabled, &flag.Revision); err != nil {
			return nil, err
		}
		flags = append(flags, flag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return flags, nil
}
