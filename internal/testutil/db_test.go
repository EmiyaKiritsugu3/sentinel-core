package testutil

import (
	"os"
	"os/exec"
	"testing"
)

func TestSetupTestDB(t *testing.T) {
	db := SetupTestDB(t)
	if db == nil {
		t.Fatal("SetupTestDB returned nil db")
	}
	defer db.Close()

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

// TestAssertSQLiteDB_NilDB exercises the t.Fatalf branch in AssertSQLiteDB
// when called with a nil *sqlite.DB. It runs itself as a subprocess because
// t.Fatalf terminates the goroutine, which cannot be asserted from the same process.
func TestAssertSQLiteDB_NilDB(t *testing.T) {
	if os.Getenv("TEST_ASSERT_NIL_DB") == "1" {
		AssertSQLiteDB(t, nil, "db")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestAssertSQLiteDB_NilDB")
	cmd.Env = append(os.Environ(), "TEST_ASSERT_NIL_DB=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for nil db assertion")
	}
}
