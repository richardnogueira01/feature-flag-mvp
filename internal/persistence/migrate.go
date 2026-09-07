package persistence

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/000001_initial.up.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	migration, err := migrationFS.ReadFile("migrations/000001_initial.up.sql")
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, string(migration))
	return err
}
