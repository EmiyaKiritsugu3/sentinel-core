package report

import (
	"path/filepath"
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
	if err := graph.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestFetchStats_EmptyDB(t *testing.T) {
	db := setupAggregatorDB(t)
	defer db.Close()

	agg := NewAggregator(db)
	stats, err := agg.FetchStats()
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
	db := setupAggregatorDB(t)
	defer db.Close()

	_, err := db.Conn.Exec("INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
		"n1", "main.go", "file", "cmd/main.go")
	if err != nil {
		t.Fatalf("insert node: %v", err)
	}
	_, err = db.Conn.Exec("INSERT INTO nodes (id, name, type, file_path) VALUES (?, ?, ?, ?)",
		"n2", "Handler", "struct", "internal/handler.go")
	if err != nil {
		t.Fatalf("insert node: %v", err)
	}

	agg := NewAggregator(db)
	stats, err := agg.FetchStats()
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
	db := setupAggregatorDB(t)
	defer db.Close()

	_, err := db.Conn.Exec("INSERT INTO tasks (id, description, status, tier) VALUES (?, ?, ?, ?)",
		"t1", "Add auth", "DONE", "T1")
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}
	_, err = db.Conn.Exec("INSERT INTO tasks (id, description, status, tier, math_delta) VALUES (?, ?, ?, ?, ?)",
		"t2", "Add logging", "DONE", "T2", 5.5)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}
	_, err = db.Conn.Exec("INSERT INTO tasks (id, description, status, tier) VALUES (?, ?, ?, ?)",
		"t3", "Fix bug", "FAILED", "T3")
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	agg := NewAggregator(db)
	stats, err := agg.FetchStats()
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
	db := setupAggregatorDB(t)
	defer db.Close()

	agg := NewAggregator(db)
	stats, err := agg.FetchStats()
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
	db := setupAggregatorDB(t)
	defer db.Close()

	agg := NewAggregator(db)
	if agg == nil {
		t.Fatal("NewAggregator returned nil")
	}
	if agg.db != db {
		t.Error("NewAggregator did not store db reference")
	}
}
