package agents

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestDispatcher_ReconcileEvents(t *testing.T) {
	db, err := sqlite.Init()
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()
	defer os.Remove(".sentinel/graph.db")

	// Criar tabela de sub_tasks para o teste (Simulando migração)
	_, _ = db.Conn.Exec("CREATE TABLE sub_tasks (id TEXT PRIMARY KEY, status TEXT, updated_at TIMESTAMP)")
	_, _ = db.Conn.Exec("INSERT INTO sub_tasks (id, status) VALUES ('task-1', 'PENDING')")

	eventDir := ".sentinel/events"
	os.MkdirAll(eventDir, 0755)
	defer os.RemoveAll(".sentinel")

	eventFile := filepath.Join(eventDir, "task-1.json")
	eventData := map[string]string{
		"sub_task_id": "task-1",
		"status":      "DONE",
	}
	bytes, _ := json.Marshal(eventData)
	os.WriteFile(eventFile, bytes, 0644)

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
