package agents

import (
	"context"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
	"github.com/google/generative-ai-go/genai"
)

// --- readPriorTrust ---

func TestReadPriorTrust_NoRow(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	successes, total, trust, err := readPriorTrust(db, "nonexistent-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if successes != 0 || total != 0 {
		t.Errorf("expected (0, 0), got (%d, %d)", successes, total)
	}
	if trust <= 0 || trust >= 1 {
		t.Errorf("trust for new agent should be ~0.5 (Laplace prior), got %.4f", trust)
	}
}

func TestReadPriorTrust_ExistingRow(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	_, err := db.Conn.Exec(
		"INSERT INTO agent_trust (agent_name, successes, total, trust_score) VALUES (?, ?, ?, ?)",
		"test-agent", 7, 10, 0.7273,
	)
	if err != nil {
		t.Fatalf("failed to seed trust data: %v", err)
	}

	successes, total, trust, err := readPriorTrust(db, "test-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if successes != 7 || total != 10 {
		t.Errorf("expected (7, 10), got (%d, %d)", successes, total)
	}
	if trust <= 0 {
		t.Errorf("trust should be positive, got %.4f", trust)
	}
}

func TestReadPriorTrust_ClosedDB(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	db.Close()

	_, _, _, err := readPriorTrust(db, "any-agent")
	if err == nil {
		t.Fatal("expected error for closed DB, got nil")
	}
}

// --- persistTrust ---

func TestPersistTrust_NewAgent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	err := persistTrust(db, "new-agent", 0, 0, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var successes, total int
	var trustScore float64
	err = db.Conn.QueryRow(
		"SELECT successes, total, trust_score FROM agent_trust WHERE agent_name = ?", "new-agent",
	).Scan(&successes, &total, &trustScore)
	if err != nil {
		t.Fatalf("failed to query trust: %v", err)
	}
	if successes != 1 || total != 1 {
		t.Errorf("expected (1, 1), got (%d, %d)", successes, total)
	}
}

func TestPersistTrust_ExistingAgentFailure(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	_, err := db.Conn.Exec(
		"INSERT INTO agent_trust (agent_name, successes, total, trust_score) VALUES (?, ?, ?, ?)",
		"existing-agent", 5, 10, 0.5455,
	)
	if err != nil {
		t.Fatalf("failed to seed trust data: %v", err)
	}

	err = persistTrust(db, "existing-agent", 5, 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var successes, total int
	err = db.Conn.QueryRow(
		"SELECT successes, total FROM agent_trust WHERE agent_name = ?", "existing-agent",
	).Scan(&successes, &total)
	if err != nil {
		t.Fatalf("failed to query trust: %v", err)
	}
	if successes != 5 || total != 11 {
		t.Errorf("expected (5, 11) after failure, got (%d, %d)", successes, total)
	}
}

func TestPersistTrust_ClosedDB(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	db.Close()

	err := persistTrust(db, "any-agent", 0, 0, true)
	if err == nil {
		t.Fatal("expected error for closed DB, got nil")
	}
}

// --- countThoughtActionTokens ---

func TestCountThoughtActionTokens_AllAction(t *testing.T) {
	parts := []genai.Part{
		genai.Text("I will now implement the feature."),
		genai.Text("Writing code to handle authentication."),
	}
	actionTokens, thoughtTokens := countThoughtActionTokens(parts)
	if actionTokens == 0 {
		t.Error("expected non-zero action tokens")
	}
	if thoughtTokens != 0 {
		t.Errorf("expected 0 thought tokens, got %d", thoughtTokens)
	}
}

func TestCountThoughtActionTokens_AllThought(t *testing.T) {
	parts := []genai.Part{
		genai.Text("```thought\nthis is my reasoning step"),
		genai.Text("```thought\nanother thought block"),
	}
	actionTokens, thoughtTokens := countThoughtActionTokens(parts)
	if actionTokens != 0 {
		t.Errorf("expected 0 action tokens, got %d", actionTokens)
	}
	if thoughtTokens == 0 {
		t.Error("expected non-zero thought tokens")
	}
}

func TestCountThoughtActionTokens_Mixed(t *testing.T) {
	parts := []genai.Part{
		genai.Text("Let me implement this now."),
		genai.Text("```thought\nreasoning about the approach"),
	}
	actionTokens, thoughtTokens := countThoughtActionTokens(parts)
	if actionTokens == 0 {
		t.Error("expected non-zero action tokens for mixed parts")
	}
	if thoughtTokens == 0 {
		t.Error("expected non-zero thought tokens for mixed parts")
	}
}

func TestCountThoughtActionTokens_CodeBlockThought(t *testing.T) {
	parts := []genai.Part{
		genai.Text("```thought\nreasoning here\n```"),
	}
	actionTokens, thoughtTokens := countThoughtActionTokens(parts)
	if actionTokens != 0 {
		t.Errorf("expected 0 action tokens for code-block thought, got %d", actionTokens)
	}
	if thoughtTokens == 0 {
		t.Error("expected non-zero thought tokens for code-block thought")
	}
}

func TestCountThoughtActionTokens_Empty(t *testing.T) {
	actionTokens, thoughtTokens := countThoughtActionTokens(nil)
	if actionTokens != 0 || thoughtTokens != 0 {
		t.Errorf("expected (0, 0) for nil parts, got (%d, %d)", actionTokens, thoughtTokens)
	}
}

// --- checkGateA ---

func TestCheckGateA_Intervene(t *testing.T) {
	intervene, msg := checkGateA(3.5, 2.0)
	if !intervene {
		t.Error("expected intervene=true when lambda > effectiveMaxLambda")
	}
	if msg == "" {
		t.Error("expected non-empty intervention message")
	}
}

func TestCheckGateA_NoIntervene(t *testing.T) {
	intervene, msg := checkGateA(1.5, 3.0)
	if intervene {
		t.Error("expected intervene=false when lambda < effectiveMaxLambda")
	}
	if msg != "" {
		t.Errorf("expected empty message when no intervention, got %q", msg)
	}
}

func TestCheckGateA_ExactlyAtThreshold(t *testing.T) {
	intervene, _ := checkGateA(2.0, 2.0)
	if intervene {
		t.Error("expected intervene=false when lambda == effectiveMaxLambda (not strictly greater)")
	}
}

// --- checkGateA5 ---

func TestCheckGateA5_NoPreviousLambda(t *testing.T) {
	newCount, intervene, msg := checkGateA5(2.0, 0, 0)
	if intervene {
		t.Error("expected no intervention when previousLambda=0")
	}
	if newCount != 0 {
		t.Errorf("expected newCount=0, got %d", newCount)
	}
	if msg != "" {
		t.Errorf("expected empty message, got %q", msg)
	}
}

func TestCheckGateA5_FirstDivergence(t *testing.T) {
	newCount, intervene, _ := checkGateA5(10.0, 1.0, 0)
	if intervene {
		t.Error("expected no intervention on first divergence (count=1 < 2)")
	}
	if newCount != 1 {
		t.Errorf("expected newCount=1, got %d", newCount)
	}
}

func TestCheckGateA5_SecondDivergence_Intervene(t *testing.T) {
	newCount, intervene, msg := checkGateA5(10.0, 1.0, 1)
	if !intervene {
		t.Error("expected intervention on second consecutive divergence")
	}
	if newCount != 2 {
		t.Errorf("expected newCount=2, got %d", newCount)
	}
	if msg == "" {
		t.Error("expected non-empty intervention message")
	}
}

func TestCheckGateA5_StableStep_ResetsCount(t *testing.T) {
	newCount, intervene, _ := checkGateA5(1.05, 1.0, 3)
	if intervene {
		t.Error("expected no intervention on stable step")
	}
	if newCount != 0 {
		t.Errorf("expected newCount=0 (reset on stable), got %d", newCount)
	}
}

func TestCheckGateA5_NegativePreviousLambda(t *testing.T) {
	newCount, intervene, _ := checkGateA5(2.0, -0.5, 0)
	if intervene {
		t.Error("expected no intervention for negative previousLambda")
	}
	if newCount != 0 {
		t.Errorf("expected newCount=0, got %d", newCount)
	}
}

// --- persistMetrics ---

func TestPersistMetrics(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "metrics-test-task"
	_, err := db.Conn.Exec(
		"INSERT INTO tasks (id, description, status, tier) VALUES (?, ?, ?, ?)",
		taskID, "metrics test", "IN_PROGRESS", "T2",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	err = persistMetrics(db, taskID, 500, 0.05, 1200.0, 0.8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var latencyMs float64
	var tokensUsed int
	var apiCost, mathDelta float64
	err = db.Conn.QueryRow(
		"SELECT latency_ms, tokens_used, api_cost, math_delta FROM tasks WHERE id = ?", taskID,
	).Scan(&latencyMs, &tokensUsed, &apiCost, &mathDelta)
	if err != nil {
		t.Fatalf("failed to query metrics: %v", err)
	}
	if latencyMs != 1200.0 {
		t.Errorf("expected latency_ms=1200, got %.2f", latencyMs)
	}
	if tokensUsed != 500 {
		t.Errorf("expected tokens_used=500, got %d", tokensUsed)
	}
}

func TestPersistMetrics_ClosedDB(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	db.Close()

	err := persistMetrics(db, "any-task", 100, 0.01, 500.0, 0.5)
	if err == nil {
		t.Fatal("expected error for closed DB, got nil")
	}
}

// --- containsSovereignAudit ---

func TestContainsSovereignAudit(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		if !containsSovereignAudit([]string{"Some text", "Sovereign Audit Report"}) {
			t.Error("expected true when Sovereign Audit Report is present")
		}
	})
	t.Run("absent", func(t *testing.T) {
		if containsSovereignAudit([]string{"Just some text"}) {
			t.Error("expected false when Sovereign Audit Report is absent")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if containsSovereignAudit(nil) {
			t.Error("expected false for nil input")
		}
	})
}

// --- ValidateArguments coverage for tools ---

func TestReadFileTool_ValidateArguments(t *testing.T) {
	tool := &ReadFileTool{}
	v := &reflect.Validator{}

	t.Run("missing path", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{})
		if err == nil {
			t.Fatal("expected error for missing path")
		}
	})

	t.Run("valid path", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{"path": "main.go"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestWriteFileTool_ValidateArguments(t *testing.T) {
	tool := &WriteFileTool{}
	v := &reflect.Validator{}

	t.Run("missing path", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{})
		if err == nil {
			t.Fatal("expected error for missing path")
		}
	})

	t.Run("valid path", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{"path": "output.go"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestReplaceTool_ValidateArguments(t *testing.T) {
	tool := &ReplaceTool{}
	v := &reflect.Validator{}

	t.Run("missing path", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{})
		if err == nil {
			t.Fatal("expected error for missing path")
		}
	})

	t.Run("valid path", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{"path": "main.go"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunTool_ValidateArguments(t *testing.T) {
	tool := &RunTool{}
	v := &reflect.Validator{}

	t.Run("missing command", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{})
		if err == nil {
			t.Fatal("expected error for missing command")
		}
	})

	t.Run("valid command", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{"command": "go test ./..."})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGrepSearchTool_ValidateArguments(t *testing.T) {
	tool := &GrepSearchTool{}
	v := &reflect.Validator{}

	t.Run("no dir_path returns nil", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{"pattern": "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with dir_path validates it", func(t *testing.T) {
		err := tool.ValidateArguments(v, map[string]interface{}{"pattern": "test", "dir_path": "src"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAuditTool_ValidateArguments(t *testing.T) {
	tool := &AuditTool{}
	err := tool.ValidateArguments(nil, nil)
	if err != nil {
		t.Fatalf("expected nil (no args validation), got: %v", err)
	}
}

func TestScanTool_ValidateArguments(t *testing.T) {
	tool := &ScanTool{}
	err := tool.ValidateArguments(nil, nil)
	if err != nil {
		t.Fatalf("expected nil (no args validation), got: %v", err)
	}
}

func TestADRTool_ValidateArguments(t *testing.T) {
	tool := &ADRTool{}

	t.Run("missing field", func(t *testing.T) {
		err := tool.ValidateArguments(nil, map[string]interface{}{
			"title": "Test", "context": "ctx", "decision": "dec",
		})
		if err == nil {
			t.Fatal("expected error for missing fields")
		}
	})

	t.Run("all fields present", func(t *testing.T) {
		err := tool.ValidateArguments(nil, map[string]interface{}{
			"title": "Test", "context": "ctx", "decision": "dec",
			"consequences": "cons", "verification_command": "go test",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty field", func(t *testing.T) {
		err := tool.ValidateArguments(nil, map[string]interface{}{
			"title": "", "context": "ctx", "decision": "dec",
			"consequences": "cons", "verification_command": "go test",
		})
		if err == nil {
			t.Fatal("expected error for empty title")
		}
	})
}

func TestDecomposeTool_ValidateArguments_EdgeCases(t *testing.T) {
	tool := &DecomposeTool{}

	t.Run("missing subtasks key", func(t *testing.T) {
		err := tool.ValidateArguments(nil, map[string]interface{}{})
		if err == nil {
			t.Fatal("expected error for missing subtasks")
		}
	})

	t.Run("subtasks not an array", func(t *testing.T) {
		err := tool.ValidateArguments(nil, map[string]interface{}{
			"subtasks": "not-an-array",
		})
		if err == nil {
			t.Fatal("expected error for non-array subtasks")
		}
	})

	t.Run("invalid subtask object", func(t *testing.T) {
		err := tool.ValidateArguments(nil, map[string]interface{}{
			"subtasks": []interface{}{"not-a-map"},
		})
		if err == nil {
			t.Fatal("expected error for invalid subtask object")
		}
	})

	t.Run("non-string capability", func(t *testing.T) {
		err := tool.ValidateArguments(nil, map[string]interface{}{
			"subtasks": []interface{}{
				map[string]interface{}{
					"description":   "task",
					"branch_name":   "b1",
					"capabilities":  []interface{}{123},
				},
			},
		})
		if err == nil {
			t.Fatal("expected error for non-string capability")
		}
	})
}

// --- ReadFileTool.Execute ---

func TestReadFileTool_Execute(t *testing.T) {
	tool := &ReadFileTool{}

	t.Run("nonexistent file returns error", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "nonexistent_file_12345.go",
		})
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

// --- WriteFileTool.Execute: AST validation ---

func TestWriteFileTool_Execute_InvalidGo(t *testing.T) {
	tool := &WriteFileTool{}
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    "test_bad.go",
		"content": "this is not valid go !!!",
	})
	if err == nil {
		t.Fatal("expected error for invalid Go syntax")
	}
}

// --- ReplaceTool.Execute: old_string not found ---

func TestReplaceTool_Execute_OldStringNotFound(t *testing.T) {
	tool := &ReplaceTool{}
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":       "nonexistent_replace.go",
		"old_string": "does not exist",
		"new_string": "replacement",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file in replace")
	}
}
