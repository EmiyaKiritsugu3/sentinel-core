package intake_test

import (
	"context"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/intake"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
)

func TestVaguenessScore_AnchorPhase2(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	_, err := db.Conn.Exec("INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
		"node1", "AuthService", "struct", "internal/auth/service.go")
	if err != nil {
		t.Fatalf("failed to insert node: %v", err)
	}

	d := intake.NewDisambiguator(db)

	scoreVague := d.VaguenessScore(context.Background(), "improve performance")
	scoreAnchored := d.VaguenessScore(context.Background(), "improve AuthService performance")

	if scoreAnchored >= scoreVague {
		t.Errorf("want scoreAnchored < scoreVague, got anchored=%.2f vague=%.2f", scoreAnchored, scoreVague)
	}
}

func TestAnalyze_GraphSuggestions(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

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
		if _, err := db.Conn.Exec("INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
			n.id, n.name, "struct", n.path); err != nil {
			t.Fatalf("failed to insert node %s: %v", n.name, err)
		}
	}

	d := intake.NewDisambiguator(db)
	vague, suggestions := d.Analyze(context.Background(), "fix it for auth")

	if !vague {
		t.Error("want vague=true for 'fix it for auth'")
	}
	if len(suggestions) == 0 {
		t.Error("want suggestions, got 0")
	}
	if len(suggestions) > 5 {
		t.Errorf("want max 5 suggestions, got %d", len(suggestions))
	}

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
