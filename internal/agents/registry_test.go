package agents

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestRegistryManager_SelectBest(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_registry.db")
	db, err := sqlite.InitAtPath(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS specialist_registry (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		base_persona TEXT NOT NULL,
		current_persona_path TEXT NOT NULL,
		reliability_score REAL DEFAULT 1.0,
		capabilities TEXT
	);`
	if _, err := db.Conn.Exec(schema); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	caps1, _ := json.Marshal([]string{"go", "sqlite"})
	caps2, _ := json.Marshal([]string{"go", "react", "typescript"})
	caps3, _ := json.Marshal([]string{"rust", "wasm"})

	// Insert test data
	inserts := []struct {
		id   string
		name string
		rel  float64
		caps string
	}{
		{"s1", "Go Expert", 0.95, string(caps1)},
		{"s2", "Fullstack Expert", 0.98, string(caps2)},
		{"s3", "Rust Expert", 0.90, string(caps3)},
	}

	for _, ins := range inserts {
		_, err := db.Conn.Exec("INSERT INTO specialist_registry (id, name, base_persona, current_persona_path, reliability_score, capabilities) VALUES (?, ?, ?, ?, ?, ?)",
			ins.id, ins.name, "persona", "path", ins.rel, ins.caps)
		if err != nil {
			t.Fatalf("failed to insert data: %v", err)
		}
	}

	mgr := NewRegistryManager(db)
	ctx := context.Background()

	t.Run("Match Go and SQLite", func(t *testing.T) {
		s, err := mgr.SelectBest(ctx, []string{"go", "sqlite"})
		if err != nil {
			t.Fatalf("expected specialist, got error: %v", err)
		}
		if s.ID != "s1" {
			t.Errorf("expected s1, got %s", s.ID)
		}
	})

	t.Run("Match Go (Highest reliability)", func(t *testing.T) {
		// Both s1 and s2 have "go", but s2 has higher reliability (0.98 vs 0.95)
		s, err := mgr.SelectBest(ctx, []string{"go"})
		if err != nil {
			t.Fatalf("expected specialist, got error: %v", err)
		}
		if s.ID != "s2" {
			t.Errorf("expected s2, got %s", s.ID)
		}
	})

	t.Run("No Match", func(t *testing.T) {
		_, err := mgr.SelectBest(ctx, []string{"python"})
		if err == nil {
			t.Error("expected error for no match, got nil")
		}
	})
}
