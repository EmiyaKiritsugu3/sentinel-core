package agents

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
)

func TestRegistryManager_SelectBest(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	// Clear seeded specialists so test data is the only source of matches.
	if _, err := db.Conn.Exec("DELETE FROM specialist_registry"); err != nil {
		t.Fatalf("failed to clear seeded specialists: %v", err)
	}

	caps1, _ := json.Marshal([]string{"go", "sqlite"})
	caps2, _ := json.Marshal([]string{"go", "react", "typescript"})
	caps3, _ := json.Marshal([]string{"rust", "wasm"})

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
