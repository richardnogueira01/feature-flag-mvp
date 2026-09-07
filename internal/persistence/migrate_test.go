package persistence

import "testing"

func TestEmbeddedMigrationIsPresent(t *testing.T) {
	migration, err := migrationFS.ReadFile("migrations/000001_initial.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(migration) == 0 {
		t.Fatal("embedded migration is empty")
	}
}
