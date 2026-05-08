// Package testutil provides shared test helpers for database setup and validation.
package testutil

import (
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// SetupTestDB creates a temporary SQLite database and returns a ready-to-use
// *sqlite.DB. The caller is responsible for running migrations and calling db.Close().
func SetupTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sqlite.InitAtPath(dbPath)
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	return db
}

// AssertSQLiteDB fails the test immediately if the database or its connection is nil.
func AssertSQLiteDB(t *testing.T, db *sqlite.DB, name string) {
	t.Helper()
	if db == nil || db.Conn == nil {
		t.Fatalf("%s or %s.Conn is nil", name, name)
	}
}
