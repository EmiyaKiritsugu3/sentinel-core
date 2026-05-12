package commands

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/patterns"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func setupCmdStore(t *testing.T) (*patterns.PatternStore, *sqlite.DB) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sqlite.InitAtPath(dbPath)
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := patterns.NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}
	return store, db
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = old
	out, _ := io.ReadAll(r)
	r.Close()
	return string(out)
}

func TestRunBackfillCognitiveDNA_Success(t *testing.T) {
	store, _ := setupCmdStore(t)
	stderr := captureStderr(t, func() {
		runBackfillCognitiveDNA(store, findProjectRoot())
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestRunBackfillCognitiveDNA_Error(t *testing.T) {
	store, _ := setupCmdStore(t)
	stderr := captureStderr(t, func() {
		runBackfillCognitiveDNA(store, "/nonexistent/path")
	})
	if stderr == "" {
		t.Fatal("expected warning on stderr for nonexistent path")
	}
}

func TestRunBackfillEvolutionInsights_Success(t *testing.T) {
	store, _ := setupCmdStore(t)
	stderr := captureStderr(t, func() {
		runBackfillEvolutionInsights(store, findProjectRoot())
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestRunBackfillEvolutionInsights_Error(t *testing.T) {
	store, _ := setupCmdStore(t)
	stderr := captureStderr(t, func() {
		runBackfillEvolutionInsights(store, "/nonexistent/path")
	})
	if stderr == "" {
		t.Fatal("expected warning on stderr for nonexistent path")
	}
}

func TestRunBackfillSentinelLog_Success(t *testing.T) {
	store, _ := setupCmdStore(t)

	dir := t.TempDir()
	docDir := filepath.Join(dir, "docs", "process")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := "# Log\n- Filtro A aplicado no roteamento de módulos críticos\n"
	if err := os.WriteFile(filepath.Join(docDir, "sentinel-log.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stderr := captureStderr(t, func() {
		runBackfillSentinelLog(store, dir)
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestRunBackfillSentinelLog_Error(t *testing.T) {
	store, _ := setupCmdStore(t)
	stderr := captureStderr(t, func() {
		runBackfillSentinelLog(store, "/nonexistent/path")
	})
	if stderr == "" {
		t.Fatal("expected warning on stderr for nonexistent path")
	}
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
