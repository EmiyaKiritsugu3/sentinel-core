package sqlite

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestValidateDB_NilDB(t *testing.T) {
	t.Parallel()
	err := ValidateDB(nil, "caller")
	if err == nil {
		t.Fatal("expected error for nil db")
	}
	if !errors.Is(err, ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
}

func TestValidateDB_NilConn(t *testing.T) {
	t.Parallel()
	db := &DB{Conn: nil}
	err := ValidateDB(db, "engine")
	if err == nil {
		t.Fatal("expected error for nil Conn")
	}
	if !errors.Is(err, ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
}

func TestValidateDB_ValidDB(t *testing.T) {
	t.Parallel()
	db := SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := ValidateDB(db, "engine"); err != nil {
		t.Errorf("expected nil error for valid db, got: %v", err)
	}
}

// SetupTestDB is a local copy for the sqlite package test.
func SetupTestDB(t *testing.T) *DB {
	t.Helper()
	tmpDir := t.TempDir()
	db, err := InitAtPath(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	return db
}
