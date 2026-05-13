package graph

import (
	"context"
	"os"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
)

func TestLinker_Integration(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	oldCwd, _ := os.Getwd()
	_ = os.Chdir(t.TempDir())
	defer func() { _ = os.Chdir(oldCwd) }()

	engine, err := NewEngine(db)
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	_ = os.MkdirAll("src/components", 0755)                                                       //nolint:gosec // test fixture
	_ = os.WriteFile("src/app.tsx", []byte("import Button from './components/Button'"), 0644)     //nolint:gosec // test fixture
	_ = os.WriteFile("src/components/Button.tsx", []byte("export const Button = () => {}"), 0644) //nolint:gosec // test fixture

	_ = os.MkdirAll("src/utils", 0755)                                                  //nolint:gosec // test fixture
	_ = os.WriteFile("src/utils/index.ts", []byte("export const log = () => {}"), 0644) //nolint:gosec // test fixture

	_ = os.MkdirAll("pkg/utils", 0755)                                     //nolint:gosec // test fixture
	_ = os.WriteFile("pkg/utils/helper.go", []byte("package utils"), 0644) //nolint:gosec // test fixture
	_ = os.WriteFile("cmd/main.go", []byte("package main"), 0644)          //nolint:gosec // test fixture

	_ = os.MkdirAll("src/deep/a/b/c", 0755)                      //nolint:gosec // test fixture
	_ = os.WriteFile("src/deep/a/b/c/leaf.ts", []byte(""), 0644) //nolint:gosec // test fixture
	_ = os.WriteFile("src/deep/root.ts", []byte(""), 0644)       //nolint:gosec // test fixture

	testNodes := []Node{
		{ID: "file:src/app.tsx", Name: "app.tsx", Type: "file", FilePath: "src/app.tsx"},
		{ID: "file:src/components/Button.tsx", Name: "Button.tsx", Type: "file", FilePath: "src/components/Button.tsx"},
		{ID: "file:src/utils/index.ts", Name: "index.ts", Type: "file", FilePath: "src/utils/index.ts"},
		{ID: "file:cmd/main.go", Name: "main.go", Type: "file", FilePath: "cmd/main.go"},
		{ID: "file:pkg/utils", Name: "utils", Type: "file", FilePath: "pkg/utils"},
		{ID: "file:src/deep/a/b/c/leaf.ts", Name: "leaf.ts", Type: "file", FilePath: "src/deep/a/b/c/leaf.ts"},
		{ID: "file:src/deep/root.ts", Name: "root.ts", Type: "file", FilePath: "src/deep/root.ts"},

		{ID: "import:1", Name: "./components/Button", Type: "unresolved_import", FilePath: "src/app.tsx"},
		{ID: "import:2", Name: "./utils", Type: "unresolved_import", FilePath: "src/app.tsx"},
		{ID: "import:3", Name: "github.com/EmiyaKiritsugu3/sentinel-core/pkg/utils", Type: "unresolved_import", FilePath: "cmd/main.go"},
		{ID: "import:4", Name: "./missing", Type: "unresolved_import", FilePath: "src/app.tsx"},
		{ID: "import:5", Name: "../../../root", Type: "unresolved_import", FilePath: "src/deep/a/b/c/leaf.ts"},
	}

	for _, n := range testNodes {
		query := `INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)`
		_, err := db.Conn.ExecContext(context.Background(), query, n.ID, n.Name, n.Type, n.FilePath)
		if err != nil {
			t.Fatalf("failed to insert test node %s: %v", n.ID, err)
		}
	}

	err = engine.LinkDependencies(context.Background())
	if err != nil {
		t.Fatalf("LinkDependencies failed: %v", err)
	}

	expectedEdges := []struct {
		from, to string
	}{
		{"file:src/app.tsx", "file:src/components/Button.tsx"},
		{"file:src/app.tsx", "file:src/utils/index.ts"},
		{"file:cmd/main.go", "file:pkg/utils"},
		{"file:src/deep/a/b/c/leaf.ts", "file:src/deep/root.ts"},
	}

	for _, ee := range expectedEdges {
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM edges WHERE from_node_id = ? AND to_node_id = ?)"
		err := db.Conn.QueryRowContext(context.Background(), query, ee.from, ee.to).Scan(&exists)
		if err != nil || !exists {
			t.Errorf("expected edge %s -> %s not found", ee.from, ee.to)
		}
	}

	resolvedIDs := []string{"import:1", "import:2", "import:3", "import:5"}
	for _, id := range resolvedIDs {
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM nodes WHERE id = ?)"
		_ = db.Conn.QueryRowContext(context.Background(), query, id).Scan(&exists)
		if exists {
			t.Errorf("expected node %s to be deleted", id)
		}
	}

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM nodes WHERE id = 'import:4')"
	_ = db.Conn.QueryRowContext(context.Background(), query).Scan(&exists)
	if !exists {
		t.Errorf("expected unresolved node import:4 to still exist")
	}
}
