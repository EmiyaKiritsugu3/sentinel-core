package testutil

import (
	"testing"
)

func TestSetupTestDB(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	if db == nil {
		t.Fatal("SetupTestDB returned nil db")
	}
	if db.Conn == nil {
		t.Fatal("SetupTestDB returned db with nil Conn")
	}
	if err := db.Conn.Ping(); err != nil {
		t.Fatalf("test db ping failed: %v", err)
	}
}

func TestAssertSQLiteDB_ValidDB(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	AssertSQLiteDB(t, db, "db")
}
