package graph

import (
	"database/sql"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
)

func TestMigrate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	sqlDB := db.Conn

	tables := []string{
		"specialist_registry",
		"sub_tasks",
		"performance_logs",
		"agent_trust",
	}

	for _, table := range tables {
		var name string
		err := sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("table %s was not created", table)
			} else {
				t.Errorf("failed to query sqlite_master for table %s: %v", table, err)
			}
		}
	}
}
