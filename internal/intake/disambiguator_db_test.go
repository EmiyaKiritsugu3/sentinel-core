// internal/intake/disambiguator_db_test.go
package intake_test

import (
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/intake"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestVaguenessScore_AnchorPhase2(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.InitAtPath(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	// Insert some nodes to anchor to
	_, err = db.Conn.Exec("INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
		"node1", "AuthService", "struct", "internal/auth/service.go")
	if err != nil {
		t.Fatalf("failed to insert node: %v", err)
	}

	d := intake.NewDisambiguator(db)

	// Case 1: Vague description with no matches in graph
	scoreVague := d.VaguenessScore("improve performance")
	
	// Case 2: Description with keywords that match graph
	scoreAnchored := d.VaguenessScore("improve AuthService performance")

	if scoreAnchored >= scoreVague {
		t.Errorf("want scoreAnchored < scoreVague, got anchored=%.2f vague=%.2f", scoreAnchored, scoreVague)
	}
}

func TestAnalyze_GraphSuggestions(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.InitAtPath(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	// Insert nodes for suggestion
	nodes := []struct {
		id   string
		name string
		path string
	}{
		{"1", "AuthMiddleware", "pkg/auth/middleware.go"},
		{"2", "Authenticator", "pkg/auth/auth.go"},
		{"3", "Authorizer", "pkg/auth/auth.go"},
		{"4", "TokenProvider", "pkg/auth/token.go"},
		{"5", "SessionStore", "pkg/auth/session.go"},
		{"6", "PasswordHasher", "pkg/auth/hash.go"},
	}
	for _, n := range nodes {
		_, _ = db.Conn.Exec("INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
			n.id, n.name, "struct", n.path)
	}

	d := intake.NewDisambiguator(db)
	vague, suggestions := d.Analyze("fix it for auth")

	if !vague {
		t.Error("want vague=true for 'fix it for auth'")
	}
	if len(suggestions) == 0 {
		t.Error("want suggestions, got 0")
	}
	if len(suggestions) > 5 {
		t.Errorf("want max 5 suggestions, got %d", len(suggestions))
	}

	// Check if suggestions are relevant
	found := false
	for _, s := range suggestions {
		if s.NodeName == "AuthMiddleware" {
			found = true
			break
		}
	}
	if !found {
		t.Error("want 'AuthMiddleware' in suggestions")
	}
}
