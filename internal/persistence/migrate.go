package persistence

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/000000_feature_flags_schema.up.sql migrations/000001_initial.up.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	schema, err := migrationFS.ReadFile("migrations/000000_feature_flags_schema.up.sql")
	if err != nil {
		return err
	}
	initial, err := migrationFS.ReadFile("migrations/000001_initial.up.sql")
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, string(schema)+"\n"+string(initial))
	return err
}
