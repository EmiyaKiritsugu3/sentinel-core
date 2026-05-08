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

// TestMigrate_ColumnMigration covers the ALTER TABLE migration path.
// It creates a legacy schema (tasks table without metric columns, agent_trust with specialist_id)
// then runs Migrate to exercise the column-add and column-rename code paths.
func TestMigrate_ColumnMigration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	_, err := db.Conn.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		description TEXT NOT NULL,
		status TEXT NOT NULL,
		tier TEXT,
		verification_command TEXT,
		commit_hash TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("failed to create legacy tasks table: %v", err)
	}

	_, err = db.Conn.Exec(`CREATE TABLE IF NOT EXISTS agent_trust (
		specialist_id TEXT PRIMARY KEY,
		successes INTEGER NOT NULL DEFAULT 0,
		total INTEGER NOT NULL DEFAULT 0,
		trust_score REAL NOT NULL DEFAULT 0.5,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("failed to create legacy agent_trust table: %v", err)
	}

	_, err = db.Conn.Exec(`INSERT INTO agent_trust (specialist_id, successes, total, trust_score) VALUES ('test-agent', 5, 10, 0.5)`)
	if err != nil {
		t.Fatalf("failed to seed legacy agent_trust: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed on legacy schema: %v", err)
	}

	var count int
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name IN ('latency_ms', 'tokens_used', 'api_cost', 'math_delta')").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query pragma_table_info: %v", err)
	}
	if count != 4 {
		t.Errorf("expected 4 metric columns in tasks, got %d", count)
	}

	var agentName string
	err = db.Conn.QueryRow("SELECT agent_name FROM agent_trust WHERE agent_name = 'test-agent'").Scan(&agentName)
	if err != nil {
		t.Errorf("expected agent_name column to exist after rename, got error: %v", err)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	for i := 0; i < 2; i++ {
		if err := Migrate(db); err != nil {
			t.Fatalf("Migrate pass %d failed: %v", i+1, err)
		}
	}
}
