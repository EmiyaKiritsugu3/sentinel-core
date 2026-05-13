package graph

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestMigrateAtomicity(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "sentinel-atomicity-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if sqlDB == nil {
		t.Fatal("sqlDB is nil")
	}
	defer func() { _ = sqlDB.Close() }()

	db := &sqlite.DB{Conn: sqlDB}
	testutil.AssertSQLiteDB(t, db, "db")

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("First Migrate failed: %v", err)
	}

	dbPath2 := filepath.Join(tmpDir, "test_rollback.db")
	sqlDB2, err := sql.Open("sqlite", dbPath2)
	if err != nil {
		t.Fatalf("failed to open rollback test db: %v", err)
	}
	if sqlDB2 == nil {
		t.Fatal("sqlDB2 is nil")
	}
	sqlDB2.SetMaxOpenConns(1)
	sqlDB2.SetMaxIdleConns(1)
	defer func() { _ = sqlDB2.Close() }()
	db2 := &sqlite.DB{Conn: sqlDB2}
	testutil.AssertSQLiteDB(t, db2, "db2")

	if _, err = sqlDB2.ExecContext(context.Background(), "PRAGMA query_only = ON;"); err != nil {
		t.Fatalf("failed to enable query_only pragma: %v", err)
	}
	err = Migrate(context.Background(), db2)
	if err == nil {
		t.Errorf("expected Migrate to fail on read-only DB")
	}

	var name string
	err = sqlDB2.QueryRowContext(context.Background(), "SELECT name FROM sqlite_master WHERE type='table' AND name='nodes'").Scan(&name)
	if err != sql.ErrNoRows {
		t.Errorf("expected no tables to be created on failure, but found 'nodes'")
	}
}

func TestMigrateDuplicateColumnHandling(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	sqlDB := db.Conn

	// Pre-create a minimal tasks table with latency_ms to simulate a
	// partially-migrated database where the column already exists.
	_, err := sqlDB.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS tasks (id TEXT PRIMARY KEY);")
	if err != nil {
		t.Fatalf("failed to pre-create tasks table: %v", err)
	}
	_, err = sqlDB.ExecContext(context.Background(), "ALTER TABLE tasks ADD COLUMN latency_ms REAL DEFAULT 0;")
	if err != nil {
		t.Fatalf("failed to pre-add latency_ms: %v", err)
	}

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate failed with existing column: %v", err)
	}

	var tokensUsed int
	err = sqlDB.QueryRowContext(context.Background(), "SELECT tokens_used FROM tasks LIMIT 0").Scan(&tokensUsed)
	if err != nil && err != sql.ErrNoRows {
		t.Errorf("expected tokens_used column to exist: %v", err)
	}
}
