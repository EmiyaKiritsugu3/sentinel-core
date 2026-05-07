package graph

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestMigrateAtomicity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sentinel-atomicity-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer sqlDB.Close()

	db := &sqlite.DB{Conn: sqlDB}

	// 1. First migration should pass
	if err := Migrate(db); err != nil {
		t.Fatalf("First Migrate failed: %v", err)
	}

	// 2. Test Rollback on failure
	// We'll create a conflict that causes the COMMIT or an intermediate step to fail.
	// Let's create a table that 'schema' wants to create, but with incompatible structure or something
	// or easier: just drop the db and start fresh, then make it fail halfway.
	
	dbPath2 := filepath.Join(tmpDir, "test_rollback.db")
	sqlDB2, _ := sql.Open("sqlite", dbPath2)
	defer sqlDB2.Close()
	db2 := &sqlite.DB{Conn: sqlDB2}

	// Pre-insert a specialist with an ID that Migrate will try to insert, 
	// but make it fail by adding a NOT NULL constraint on a field Migrate doesn't provide? 
	// No, Migrate provides all.
	// Let's force a failure by dropping the table just before Commit if we could.
	// Alternatively, we can use a "bad" migration string if we could inject it.
	
	// Let's use the property that we can't ALTER a table that doesn't exist.
	// If we pre-insert something into specialist_registry but the schema creation fails?
	
	// Actually, let's just verify that if Migrate fails, no tables exist.
	// To make it fail: we can lock the database or use a read-only connection?
	// SQLite specific: "PRAGMA query_only = ON;"
	_, _ = sqlDB2.Exec("PRAGMA query_only = ON;")
	err = Migrate(db2)
	if err == nil {
		t.Errorf("expected Migrate to fail on read-only DB")
	}

	// Verify no tables were created (or at least the first one 'nodes' isn't there)
	var name string
	err = sqlDB2.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='nodes'").Scan(&name)
	if err != sql.ErrNoRows {
		t.Errorf("expected no tables to be created on failure, but found 'nodes'")
	}
}

func TestMigrateDuplicateColumnHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sentinel-duplicate-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer sqlDB.Close()

	db := &sqlite.DB{Conn: sqlDB}

	// Pre-create table and ONE of the columns
	_, err = sqlDB.Exec("CREATE TABLE tasks (id TEXT PRIMARY KEY);")
	if err != nil {
		t.Fatalf("failed to pre-create tasks table: %v", err)
	}
	_, err = sqlDB.Exec("ALTER TABLE tasks ADD COLUMN latency_ms REAL DEFAULT 0;")
	if err != nil {
		t.Fatalf("failed to pre-add latency_ms: %v", err)
	}

	// Run migration - it should NOT fail despite latency_ms already existing
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed with existing column: %v", err)
	}

	// Verify other columns were added
	var tokensUsed int
	err = sqlDB.QueryRow("SELECT tokens_used FROM tasks LIMIT 0").Scan(&tokensUsed)
	// We expect no rows, but the query should not fail if column exists
	if err != nil && err != sql.ErrNoRows {
		t.Errorf("expected tokens_used column to exist: %v", err)
	}
}
