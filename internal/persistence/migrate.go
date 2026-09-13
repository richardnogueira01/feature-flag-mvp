package persistence

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/000000_feature_flags_schema.up.sql migrations/000001_initial.up.sql migrations/000002_configuration_scope.up.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	files := []string{
		"migrations/000000_feature_flags_schema.up.sql",
		"migrations/000001_initial.up.sql",
		"migrations/000002_configuration_scope.up.sql",
	}
	var script string
	for _, name := range files {
		contents, err := migrationFS.ReadFile(name)
		if err != nil {
			return err
		}
		script += string(contents) + "\n"
	}
	// Every file runs inside the same Exec call (one simple-query message, one session), so the
	// search_path set by 000000 stays in effect for 000001 and 000002.
	_, err := pool.Exec(ctx, script)
	return err
}
