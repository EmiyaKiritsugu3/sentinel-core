package agents

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestDispatcher_ReconcileEvents(t *testing.T) {
	db, err := sqlite.Init()
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()
	defer os.Remove(".sentinel/graph.db")

	// Use graph.Migrate to ensure schema is correct (Issue fix for robust testing)
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	// Insert needed data for the test
	_, err = db.Conn.Exec("INSERT INTO tasks (id, description, status) VALUES ('parent-1', 'Parent Task', 'IN_PROGRESS')")
	if err != nil {
		t.Fatalf("failed to insert parent task: %v", err)
	}
	_, err = db.Conn.Exec("INSERT INTO sub_tasks (id, parent_task_id, description, status) VALUES ('task-1', 'parent-1', 'Test Subtask', 'PENDING')")
	if err != nil {
		t.Fatalf("failed to insert sub-task: %v", err)
	}

	eventDir := ".sentinel/events"
	os.MkdirAll(eventDir, 0755)
	defer os.RemoveAll(".sentinel")

	eventFile := filepath.Join(eventDir, "task-1.json")
	eventData := map[string]string{
		"sub_task_id": "task-1",
		"status":      "DONE",
	}
	bytes, err := json.Marshal(eventData)
	if err != nil {
		t.Fatalf("failed to marshal event data: %v", err)
	}
	err = os.WriteFile(eventFile, bytes, 0644)
	if err != nil {
		t.Fatalf("failed to write event file: %v", err)
	}

	d := NewDispatcher(nil, nil, db)
	err = d.ReconcileEvents(context.Background())
	if err != nil {
		t.Fatalf("reconciliation failed: %v", err)
	}

	var status string
	err = db.Conn.QueryRow("SELECT status FROM sub_tasks WHERE id = 'task-1'").Scan(&status)
	if err != nil {
		t.Fatalf("failed to query status: %v", err)
	}

	if status != "DONE" {
		t.Errorf("expected status DONE, got %s", status)
	}

	if _, err := os.Stat(eventFile); !os.IsNotExist(err) {
		t.Error("event file was not deleted after reconciliation")
	}
}
