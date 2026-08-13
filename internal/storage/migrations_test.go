package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestFailedMigrationRollsBackSchemaAndVersionRecord(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "failed-migration.db"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatalf("create migration table: %v", err)
	}
	store := &Store{db: db}
	err = store.applyMigrationSet(context.Background(), []migration{{version: 999, statements: []string{
		`CREATE TABLE migration_should_rollback (id INTEGER PRIMARY KEY)`,
		`INSERT INTO migration_should_rollback(id) VALUES (1)`,
		`THIS IS NOT VALID SQL`,
	}}})
	if err == nil || !strings.Contains(err.Error(), "apply migration 999") {
		t.Fatalf("applyMigrationSet() error = %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'migration_should_rollback'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rolled-back table count = %d, %v", count, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 999`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rolled-back migration record count = %d, %v", count, err)
	}
}
