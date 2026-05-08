package sqlite

import "testing"

func TestValidateDB_NilDB(t *testing.T) {
	err := ValidateDB(nil, "caller")
	if err == nil {
		t.Fatal("expected error for nil db")
	}
	if err.Error() != "caller: nil db" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestValidateDB_NilConn(t *testing.T) {
	db := &DB{Conn: nil}
	err := ValidateDB(db, "engine")
	if err == nil {
		t.Fatal("expected error for nil Conn")
	}
	if err.Error() != "engine: nil db" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestValidateDB_ValidDB(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	if err := ValidateDB(db, "engine"); err != nil {
		t.Errorf("expected nil error for valid db, got: %v", err)
	}
}

// SetupTestDB is a local copy for the sqlite package test.
func SetupTestDB(t *testing.T) *DB {
	t.Helper()
	tmpDir := t.TempDir()
	db, err := InitAtPath(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	return db
}
