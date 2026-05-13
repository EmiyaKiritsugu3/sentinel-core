package report

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func setupAggregatorDB(t *testing.T) *sqlite.DB {
	t.Helper()
	tmpDir := t.TempDir()
	db, err := sqlite.InitAtPath(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	if err := graph.Migrate(context.Background(), db); err != nil {
		_ = db.Close()
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestFetchStats_EmptyDB(t *testing.T) {
	t.Parallel()
	db := setupAggregatorDB(t)
	defer func() { _ = db.Close() }()

	agg, err := NewAggregator(db)
	if err != nil {
		t.Fatalf("NewAggregator() error: %v", err)
	}
	stats, err := agg.FetchStats(context.Background())
	if err != nil {
		t.Fatalf("FetchStats() error: %v", err)
	}
	if stats.TotalNodes != 0 {
		t.Errorf("TotalNodes = %d, want 0", stats.TotalNodes)
	}
	if stats.TotalTasks != 0 {
		t.Errorf("TotalTasks = %d, want 0", stats.TotalTasks)
	}
	if stats.Tasks != nil {
		t.Errorf("Tasks = %v, want nil", stats.Tasks)
	}
}

func TestFetchStats_WithNodes(t *testing.T) {
	t.Parallel()
	db := setupAggregatorDB(t)
	defer func() { _ = db.Close() }()

	_, err := db.Conn.ExecContext(context.Background(), "INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
		"n1", "main.go", "file", "cmd/main.go")
	if err != nil {
		t.Fatalf("insert node: %v", err)
	}
	_, err = db.Conn.ExecContext(context.Background(), "INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
		"n2", "Handler", "struct", "internal/handler.go")
	if err != nil {
		t.Fatalf("insert node: %v", err)
	}

	agg, err := NewAggregator(db)
	if err != nil {
		t.Fatalf("NewAggregator() error: %v", err)
	}
	stats, err := agg.FetchStats(context.Background())
	if err != nil {
		t.Fatalf("FetchStats() error: %v", err)
	}
	if stats.TotalNodes != 2 {
		t.Errorf("TotalNodes = %d, want 2", stats.TotalNodes)
	}
	if stats.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want 1", stats.TotalFiles)
	}
	if stats.TotalStructs != 1 {
		t.Errorf("TotalStructs = %d, want 1", stats.TotalStructs)
	}
}

func TestFetchStats_WithTasks(t *testing.T) {
	t.Parallel()
	db := setupAggregatorDB(t)
	defer func() { _ = db.Close() }()

	_, err := db.Conn.ExecContext(context.Background(), "INSERT INTO tasks (id, description, status, tier) VALUES (?, ?, ?, ?)",
		"t1", "Add auth", "DONE", "T1")
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}
	_, err = db.Conn.ExecContext(context.Background(), "INSERT INTO tasks (id, description, status, tier, math_delta) VALUES (?, ?, ?, ?, ?)",
		"t2", "Add logging", "DONE", "T2", 5.5)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}
	_, err = db.Conn.ExecContext(context.Background(), "INSERT INTO tasks (id, description, status, tier) VALUES (?, ?, ?, ?)",
		"t3", "Fix bug", "FAILED", "T3")
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	agg, err := NewAggregator(db)
	if err != nil {
		t.Fatalf("NewAggregator() error: %v", err)
	}
	stats, err := agg.FetchStats(context.Background())
	if err != nil {
		t.Fatalf("FetchStats() error: %v", err)
	}
	if stats.TotalTasks != 3 {
		t.Errorf("TotalTasks = %d, want 3", stats.TotalTasks)
	}
	if stats.CompletedTasks != 2 {
		t.Errorf("CompletedTasks = %d, want 2", stats.CompletedTasks)
	}
	if stats.FailedTasks != 1 {
		t.Errorf("FailedTasks = %d, want 1", stats.FailedTasks)
	}
	if diff := stats.SuccessRate - 200.0/3.0; diff > 0.01 || diff < -0.01 {
		t.Errorf("SuccessRate = %f, want ~66.67", stats.SuccessRate)
	}
	if stats.AvgMathDelta != 2.75 {
		t.Errorf("AvgMathDelta = %f, want 2.75", stats.AvgMathDelta)
	}
	if len(stats.Tasks) != 3 {
		t.Errorf("len(Tasks) = %d, want 3", len(stats.Tasks))
	}
}

func TestFetchStats_ZeroTasks_SuccessRateZero(t *testing.T) {
	t.Parallel()
	db := setupAggregatorDB(t)
	defer func() { _ = db.Close() }()

	agg, err := NewAggregator(db)
	if err != nil {
		t.Fatalf("NewAggregator() error: %v", err)
	}
	stats, err := agg.FetchStats(context.Background())
	if err != nil {
		t.Fatalf("FetchStats() error: %v", err)
	}
	if stats.SuccessRate != 0 {
		t.Errorf("SuccessRate = %f, want 0 when no tasks", stats.SuccessRate)
	}
	if stats.AvgMathDelta != 0 {
		t.Errorf("AvgMathDelta = %f, want 0 when no tasks", stats.AvgMathDelta)
	}
}

func TestNewAggregator(t *testing.T) {
	t.Parallel()
	db := setupAggregatorDB(t)
	defer func() { _ = db.Close() }()

	agg, err := NewAggregator(db)
	if err != nil {
		t.Fatalf("NewAggregator() error: %v", err)
	}
	if agg == nil {
		t.Fatal("NewAggregator returned nil")
	}
	if agg.db != db {
		t.Error("NewAggregator did not store db reference")
	}
}

func TestNewAggregator_NilDB(t *testing.T) {
	t.Parallel()
	agg, err := NewAggregator(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if agg != nil {
		t.Error("expected nil Aggregator for nil db")
	}
}

func TestGenerateMarkdown(t *testing.T) {
	t.Parallel()
	db := setupAggregatorDB(t)
	defer func() { _ = db.Close() }()

	// Insert nodes and tasks so FetchStats returns non-zero stats.
	_, err := db.Conn.ExecContext(context.Background(), "INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
		"n1", "main.go", "file", "cmd/main.go")
	if err != nil {
		t.Fatalf("insert node: %v", err)
	}
	_, err = db.Conn.ExecContext(context.Background(), "INSERT INTO tasks (id, description, status, tier, math_delta) VALUES (?, ?, ?, ?, ?)",
		"t1", "Add auth", "DONE", "T1", 3.14)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	agg, err := NewAggregator(db)
	if err != nil {
		t.Fatalf("NewAggregator() error: %v", err)
	}
	stats, err := agg.FetchStats(context.Background())
	if err != nil {
		t.Fatalf("FetchStats() error: %v", err)
	}

	// GenerateMarkdown writes to a relative path, so chdir to a temp dir.
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir to tmpDir: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := agg.GenerateMarkdown(stats); err != nil {
		t.Fatalf("GenerateMarkdown() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("docs", "process", "COMPLIANCE-DASHBOARD.md"))
	if err != nil {
		t.Fatalf("failed to read generated dashboard: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "Sovereign Math Engine") {
		t.Error("dashboard missing 'Sovereign Math Engine' section")
	}
	if !strings.Contains(content, "+3.14") {
		t.Errorf("dashboard missing Δ value, got: %s", content)
	}
	if !strings.Contains(content, "Engineering Success Rate") {
		t.Error("dashboard missing 'Engineering Success Rate' section")
	}
	if !strings.Contains(content, "`t1`") {
		t.Error("dashboard missing task ID in detailed inventory")
	}
}

func TestFetchStats_WithADRMatch(t *testing.T) {
	t.Parallel()
	db := setupAggregatorDB(t)
	defer func() { _ = db.Close() }()

	taskID := "test-adr-task"
	_, err := db.Conn.ExecContext(context.Background(), "INSERT INTO tasks (id, description, status, tier) VALUES (?, ?, ?, ?)",
		taskID, "ADR test task", "DONE", "T1")
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	tmpDir := t.TempDir()
	adrDir := filepath.Join(tmpDir, "docs", "architecture", "adr")
	if err := os.MkdirAll(adrDir, 0755); err != nil { //nolint:gosec // test fixture
		t.Fatalf("mkdir adr: %v", err)
	}
	adrFile := filepath.Join(adrDir, "ADR-"+taskID+"-auth-decision.md")
	if err := os.WriteFile(adrFile, []byte("# ADR"), 0644); err != nil { //nolint:gosec // test fixture
		t.Fatalf("write adr file: %v", err)
	}

	originalDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	agg, err := NewAggregator(db)
	if err != nil {
		t.Fatalf("NewAggregator() error: %v", err)
	}
	stats, err := agg.FetchStats(context.Background())
	if err != nil {
		t.Fatalf("FetchStats() error: %v", err)
	}

	if len(stats.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(stats.Tasks))
	}
	if stats.Tasks[0].ADRPath == "" {
		t.Error("expected ADRPath to be set when matching ADR file exists")
	}
	expectedADR := filepath.Join("docs", "architecture", "adr", "ADR-"+taskID+"-auth-decision.md")
	if stats.Tasks[0].ADRPath != expectedADR {
		t.Errorf("ADRPath = %q, want %q", stats.Tasks[0].ADRPath, expectedADR)
	}
}
