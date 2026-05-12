package commands

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/patterns"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func setupCmdDB(t *testing.T) *sqlite.DB {
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
	return db
}

func setupCmdStore(t *testing.T) (*patterns.PatternStore, *sqlite.DB) {
	t.Helper()
	db := setupCmdDB(t)
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

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	r.Close()
	return string(out)
}

func resetAddFlags() {
	addTitle = ""
	addDesc = ""
	addCategory = ""
	addSource = ""
	addSourceRef = ""
	addTags = ""
	addImpact = ""
	addForce = false
}

func resetListFlags() {
	listCategory = ""
	listSource = ""
	listImpact = ""
}

func resetBackfillFlags() {
	backfillSource = ""
	backfillAll = false
}

func TestPatternAddCmd(t *testing.T) {
	db := setupCmdDB(t)
	cmd := NewPatternCmd(db)
	resetAddFlags()

	cmd.SetArgs([]string{"add", "--title", "Test Pattern", "--desc", "A test", "--category", "anti-pattern", "--tags", "test,go", "--impact", "high"})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("pattern add failed: %v", err)
		}
	})
	if !strings.Contains(out, "PATTERN CAPTURED") {
		t.Fatalf("expected capture output, got: %s", out)
	}
}

func TestPatternAddCmd_Duplicate(t *testing.T) {
	db := setupCmdDB(t)
	cmd := NewPatternCmd(db)

	resetAddFlags()
	cmd.SetArgs([]string{"add", "--title", "Dup Test", "--desc", "first", "--category", "anti-pattern"})
	captureStdout(t, func() { cmd.Execute() })

	resetAddFlags()
	cmd2 := NewPatternCmd(db)
	cmd2.SetArgs([]string{"add", "--title", "Dup Test", "--desc", "duplicate", "--category", "anti-pattern"})
	out := captureStdout(t, func() {
		if err := cmd2.Execute(); err != nil {
			t.Fatalf("second add failed: %v", err)
		}
	})
	if !strings.Contains(out, "Similar pattern found") {
		t.Fatalf("expected dedup warning, got: %s", out)
	}
}

func TestPatternAddCmd_ForceSkipDedup(t *testing.T) {
	db := setupCmdDB(t)
	cmd := NewPatternCmd(db)

	resetAddFlags()
	cmd.SetArgs([]string{"add", "--title", "Force Test", "--desc", "first", "--category", "anti-pattern"})
	captureStdout(t, func() { cmd.Execute() })

	resetAddFlags()
	cmd2 := NewPatternCmd(db)
	cmd2.SetArgs([]string{"add", "--title", "Force Test", "--desc", "forced", "--category", "anti-pattern", "--force"})
	out := captureStdout(t, func() {
		if err := cmd2.Execute(); err != nil {
			t.Fatalf("force add failed: %v", err)
		}
	})
	if !strings.Contains(out, "PATTERN CAPTURED") {
		t.Fatalf("expected capture output with --force, got: %s", out)
	}
}

func TestPatternListCmd(t *testing.T) {
	db := setupCmdDB(t)

	resetAddFlags()
	addCmd := NewPatternCmd(db)
	addCmd.SetArgs([]string{"add", "--title", "List Test", "--desc", "for listing", "--category", "cognitive-pattern"})
	captureStdout(t, func() { addCmd.Execute() })

	resetListFlags()
	cmd := NewPatternCmd(db)
	cmd.SetArgs([]string{"list"})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("pattern list failed: %v", err)
		}
	})
	if !strings.Contains(out, "List Test") {
		t.Fatalf("expected pattern in list, got: %s", out)
	}
}

func TestPatternListCmd_Empty(t *testing.T) {
	db := setupCmdDB(t)
	resetListFlags()
	cmd := NewPatternCmd(db)
	cmd.SetArgs([]string{"list"})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("pattern list failed: %v", err)
		}
	})
	if !strings.Contains(out, "No patterns found") {
		t.Fatalf("expected empty message, got: %s", out)
	}
}

func TestPatternSearchCmd(t *testing.T) {
	db := setupCmdDB(t)

	resetAddFlags()
	addCmd := NewPatternCmd(db)
	addCmd.SetArgs([]string{"add", "--title", "Searchable Pattern", "--desc", "for search test", "--category", "structural-principle"})
	captureStdout(t, func() { addCmd.Execute() })

	cmd := NewPatternCmd(db)
	cmd.SetArgs([]string{"search", "Searchable"})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("pattern search failed: %v", err)
		}
	})
	if !strings.Contains(out, "Searchable Pattern") {
		t.Fatalf("expected search result, got: %s", out)
	}
}

func TestPatternSearchCmd_NoMatch(t *testing.T) {
	db := setupCmdDB(t)
	cmd := NewPatternCmd(db)
	cmd.SetArgs([]string{"search", "xyznonexistent"})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("pattern search failed: %v", err)
		}
	})
	if !strings.Contains(out, "No patterns found matching") {
		t.Fatalf("expected no match message, got: %s", out)
	}
}

func TestPatternGetCmd(t *testing.T) {
	db := setupCmdDB(t)

	resetAddFlags()
	addCmd := NewPatternCmd(db)
	addCmd.SetArgs([]string{"add", "--title", "Get Test", "--desc", "for get", "--category", "routing-principle"})
	addOut := captureStdout(t, func() { addCmd.Execute() })

	idx := strings.Index(addOut, "[ID: ")
	if idx == -1 {
		t.Fatal("add command did not output ID")
	}
	idStart := idx + len("[ID: ")
	idEnd := strings.Index(addOut[idStart:], "]")
	if idEnd == -1 {
		t.Fatal("could not parse ID from add output")
	}
	patternID := strings.TrimSpace(addOut[idStart : idStart+idEnd])

	cmd := NewPatternCmd(db)
	cmd.SetArgs([]string{"get", patternID})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("pattern get failed: %v", err)
		}
	})
	if !strings.Contains(out, "Get Test") {
		t.Fatalf("expected pattern detail, got: %s", out)
	}
}

func TestPatternBackfillAllCmd(t *testing.T) {
	db := setupCmdDB(t)
	resetBackfillFlags()
	cmd := NewPatternCmd(db)
	cmd.SetArgs([]string{"backfill", "--all"})
	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("backfill --all failed: %v", err)
		}
	})
	if !strings.Contains(stderr, "cognitive-dna backfill") {
		t.Fatalf("expected cognitive-dna backfill attempt, got: %s", stderr)
	}
}

func TestPatternBackfillSourceCmd(t *testing.T) {
	db := setupCmdDB(t)
	resetBackfillFlags()
	cmd := NewPatternCmd(db)
	cmd.SetArgs([]string{"backfill", "--source", "sentinel-log"})
	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("backfill --source sentinel-log failed: %v", err)
		}
	})
	if !strings.Contains(stderr, "sentinel-log backfill") {
		t.Fatalf("expected sentinel-log backfill attempt, got: %s", stderr)
	}
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

// CG-02: NewPatternCmd com nil DB deve retornar ErrNilDB na execução

func TestNewPatternCmd_NilDB(t *testing.T) {
	cmd := NewPatternCmd(nil)
	err := cmd.Execute()
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}
