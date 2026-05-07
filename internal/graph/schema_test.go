package graph

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestMigrate(t *testing.T) {
	// Cria um diretório temporário para o teste
	tmpDir, err := os.MkdirTemp("", "sentinel-test-*")
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

	// Executa a migração
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Verifica se as novas tabelas existem
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
