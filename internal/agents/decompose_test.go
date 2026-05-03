package agents

import (
	"context"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestDecomposeTool(t *testing.T) {
	db, err := sqlite.Init()
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()

	// Ensure tables exist for the test
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	mgr := state.NewManager(db)
	taskID, err := mgr.CreateTask("Test Goal", "T3", "go test")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	if err := mgr.StartTask(taskID); err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}

	tool := &DecomposeTool{db: db}
	args := map[string]interface{}{
		"subtasks": []interface{}{
			map[string]interface{}{
				"description":  "SubTask 1",
				"capabilities": []interface{}{"go"},
				"branch_name":  "test-branch-1",
			},
			map[string]interface{}{
				"description":  "SubTask 2",
				"capabilities": []interface{}{"md"},
				"branch_name":  "test-branch-2",
			},
		},
	}

	ctx := context.Background()
	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	t.Logf("Result: %s", result)

	// Verify DB
	var count int
	err = db.Conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM sub_tasks WHERE parent_task_id = ?", taskID).Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 sub-tasks, got %d", count)
	}
}
