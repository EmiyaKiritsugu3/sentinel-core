package agents

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
	"github.com/google/generative-ai-go/genai"
)

func setupTestCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if strings.HasSuffix(cwd, "internal/agents") {
		if err := os.Chdir("../../"); err != nil {
			t.Fatalf("failed to chdir to root: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(cwd)
		})
	}
}

func TestEngine_Execute_HappyPath(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Insert task
	taskID := "happy-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test happy path", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Configure mock response for successful termination
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Analysis complete. Sovereign Audit Report: All tests green."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{
					TotalTokenCount: 100,
				},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:         "happy-agent",
		ModelID:      ModelFlash,
		MaxSteps:     5,
		Temperature:  0.2,
		SystemPrompt: "Be a helpful agent.",
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify metrics were persisted in tasks table
	var tokensUsed int
	var apiCost, mathDelta float64
	err = db.Conn.QueryRow("SELECT tokens_used, api_cost, math_delta FROM tasks WHERE id = ?", taskID).
		Scan(&tokensUsed, &apiCost, &mathDelta)
	if err != nil {
		t.Fatalf("failed to query task metrics: %v", err)
	}

	if tokensUsed != 100 {
		t.Errorf("expected 100 tokens_used, got %d", tokensUsed)
	}
	if apiCost <= 0 {
		t.Errorf("expected positive api_cost, got %f", apiCost)
	}

	// Verify trust was persisted in agent_trust
	var successes, total int
	err = db.Conn.QueryRow("SELECT successes, total FROM agent_trust WHERE agent_name = ?", "happy-agent").
		Scan(&successes, &total)
	if err != nil {
		t.Fatalf("failed to query agent_trust: %v", err)
	}

	if successes != 1 || total != 1 {
		t.Errorf("expected 1 success out of 1, got %d/%d", successes, total)
	}
}

func TestEngine_Execute_EntropyGate(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "entropy-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test entropy gate", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// First response has huge action-to-thought ratio (no think tag, long text)
	// We want stepAction = 100, stepThought = 0 => lambda = 100.
	// 400 chars ≈ 100 action tokens.
	actionText := strings.Repeat("Action code to build a compiler in Go without tests. ", 8)

	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text(actionText),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{
					TotalTokenCount: 150,
				},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>" + strings.Repeat("Reasoning extensively to calibrate the entropy ratio and make sure the action to thought ratio remains stable. ", 10) + "</think>"),
								genai.Text("Sovereign Audit Report: I have corrected my hallucination loop."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{
					TotalTokenCount: 500,
				},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	maxLambdaVal := 1.0
	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:         "entropy-agent",
		ModelID:      ModelFlash,
		MaxSteps:     5,
		MaxLambda:    &maxLambdaVal,
		SystemPrompt: "Prompt",
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify that the agent executed successfully and didn't crash, but trust was tracked
	var successes, total int
	err = db.Conn.QueryRow("SELECT successes, total FROM agent_trust WHERE agent_name = ?", "entropy-agent").
		Scan(&successes, &total)
	if err != nil {
		t.Fatalf("failed to query trust: %v", err)
	}

	if total != 1 || successes != 1 {
		t.Errorf("expected 1/1 trust, got %d/%d", successes, total)
	}
}

func TestEngine_Execute_LyapunovInterruption(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "lyapunov-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test lyapunov gate", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Step 1: low action / high thought.
	// Step 2: high action / low thought (first divergence).
	// Step 3: extremely high action / low thought (consecutive divergence -> Gate A.5 fires).
	// Step 4: terminate.
	step1Thought := "<think>Reasoning step 1: I will plan thoroughly.</think>"
	step1Action := "Action: ok."

	step2Thought := "<think>Reasoning step 2.</think>"
	step2Action := "Action: let's build something large. " + strings.Repeat("more actions ", 10)

	step3Thought := "<think>Reasoning step 3.</think>"
	step3Action := "Action: write infinite loops. " + strings.Repeat("even more actions ", 30)

	step4Thought := "<think>Reasoning step 4: Stabilizing model trajectory.</think>"
	step4Action := "Sovereign Audit Report: Stabilized reasoning and completed task."

	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text(step1Thought),
								genai.Text(step1Action),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 50},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text(step2Thought),
								genai.Text(step2Action),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 100},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text(step3Thought),
								genai.Text(step3Action),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 150},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text(step4Thought),
								genai.Text(step4Action),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 50},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:         "lyapunov-agent",
		ModelID:      ModelFlash,
		MaxSteps:     10,
		SystemPrompt: "Prompt",
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify that the Lyapunov divergence triggered an intervention, executing 5 total steps
	if ctx.Budget.StepsTaken != 5 {
		t.Errorf("expected 5 steps (including 1 Lyapunov intervention step), got %d", ctx.Budget.StepsTaken)
	}
}

func TestEngine_Execute_BudgetExceededPath(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "budget-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test budget exceeded", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Return responses that never terminate (no "sovereign audit")
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text("Thinking and keeping to work...")},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text("Still working...")},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "budget-agent",
		ModelID:  ModelFlash,
		MaxSteps: 2, // only 2 steps budget
	})

	err = engine.Execute(ctx)
	if err == nil {
		t.Fatal("expected error for budget exceeded, got nil")
	}
	if got, want := err.Error(), "agent budget exceeded (MaxSteps: 2)"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEngine_Execute_EmptyResponseError(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "empty-resp-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test empty response error", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Model returns an empty response (no candidates)
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "empty-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
	if got, want := err.Error(), "gemini: empty response from model"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEngine_Execute_SubTaskOrchestration(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Insert parent task as IN_PROGRESS so active task works
	parentID := "parent-task-123"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		parentID, "test decomposition and subtask dispatching", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	// Create a temp Git repository
	tmpDir := t.TempDir()

	// Helper to run git commands
	runCmd := func(args ...string) {
		c := exec.CommandContext(context.Background(), "git", args...)
		c.Dir = tmpDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v (output: %s)", args, err, out)
		}
	}
	runCmd("init")

	// Configure git author to avoid errors in environments without git config
	runCmd("config", "user.email", "test@example.com")
	runCmd("config", "user.name", "Test User")

	_ = os.WriteFile(filepath.Join(tmpDir, "dummy"), []byte("data"), 0644)
	runCmd("add", ".")
	runCmd("commit", "-m", "initial")

	// Create subtask-branch-1 in the git repo
	runCmd("branch", "subtask-branch-1")

	registry := NewRegistry()
	RegisterCoreTools(registry, db)

	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Wire the dispatcher!
	regMgr, err := NewRegistryManager(db)
	if err != nil {
		t.Fatalf("failed to create registry manager: %v", err)
	}
	shield := NewGitShield(tmpDir, validator)
	dispatcher, err := NewDispatcher(regMgr, shield, db)
	if err != nil {
		t.Fatalf("failed to create dispatcher: %v", err)
	}
	engine.SetDispatcher(dispatcher)

	// Verify SetDispatcher, DB(), and Registry() methods
	if engine.DB() != db {
		t.Errorf("DB() returned unexpected db handle")
	}
	if engine.Registry() != registry {
		t.Errorf("Registry() returned unexpected registry")
	}

	// Mock model responses:
	// Turn 1: Calls sentinel:decompose
	// Turn 2: Provides Sovereign Audit Report to terminate
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Let's decompose the task.</think>"),
								genai.FunctionCall{
									Name: "sentinel:decompose",
									Args: map[string]interface{}{
										"subtasks": []interface{}{
											map[string]interface{}{
												"description": "Implement feature X",
												"branch_name": "subtask-branch-1",
												"capabilities": []interface{}{"go"},
											},
										},
									},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 50},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("All subtasks dispatched. Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 50},
			},
		},
	}

	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), parentID, &AgentDefinition{
		Name:         "dispatch-agent",
		ModelID:      ModelFlash,
		MaxSteps:     5,
		Temperature:  0.2,
		SystemPrompt: "Be a coordinating agent.",
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify that the sub-task was inserted and dispatched in the database
	var count int
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM sub_tasks WHERE parent_task_id = ? AND status = 'DISPATCHED'", parentID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query sub_tasks count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 DISPATCHED subtask, got %d", count)
	}
}

func TestEngine_Execute_NoToolCalls(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "no-tool-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test no tool calls path", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Mock model responses:
	// Turn 1: Returns text but no tool calls and no "sovereign audit"
	// Turn 2: Terminates
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking...</think> Just text, no tools."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "no-tool-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestEngine_Execute_ToolNotFound(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "notfound-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test tool not found", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Model returns a tool call for a tool that doesn't exist
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking...</think>"),
								genai.FunctionCall{
									Name: "non_existent_tool",
									Args: map[string]interface{}{},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "notfound-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestEngine_Execute_HardGateValidationFailure(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "validation-fail-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test validation failure path", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	RegisterCoreTools(registry, db)

	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Model returns a tool call with an invalid path (e.g. absolute path outside repo)
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking...</think>"),
								genai.FunctionCall{
									Name: "read_file",
									Args: map[string]interface{}{
										"path": "/etc/passwd",
									},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "validation-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestEngine_Execute_PACEscalation(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "escalate-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test pac escalation", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	RegisterCoreTools(registry, db)

	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Mock responses that fail tool execution 3 times, causing PAC escalation
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking 1...</think>"),
								genai.FunctionCall{
									Name: "read_file",
									Args: map[string]interface{}{
										"path": "/etc/passwd",
									},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking 2...</think>"),
								genai.FunctionCall{
									Name: "read_file",
									Args: map[string]interface{}{
										"path": "/etc/passwd",
									},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking 3...</think>"),
								genai.FunctionCall{
									Name: "read_file",
									Args: map[string]interface{}{
										"path": "/etc/passwd",
									},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Escalated! Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "escalate-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if ctx.ActiveModel != ModelPro {
		t.Errorf("expected model to escalate to %s, got %s", ModelPro, ctx.ActiveModel)
	}
}

func TestEngine_Execute_CancelledContext(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "cancelled-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test cancelled context", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	mockClt := &MockClient{
		Model: &MockModel{
			Session: &MockSession{
				Responses: []*genai.GenerateContentResponse{
					{
						Candidates: []*genai.Candidate{
							{
								Content: &genai.Content{
									Parts: []genai.Part{
										genai.Text("Continuing..."),
									},
								},
							},
						},
						UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
					},
				},
			},
		},
	}
	engine.genaiClient = mockClt

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	ctx := NewAgentContext(cancelCtx, taskID, &AgentDefinition{
		Name:     "cancelled-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestEngine_Execute_ContextCancelledInLoop covers the select-case
// <-ctx.Context.Done() branch inside the ReAct loop (line 232 in engine.go).
// The context is cancelled mid-loop by the mock session on its 2nd SendMessage
// call, so the select fires on the 3rd iteration before SendMessage is reached.
func TestEngine_Execute_ContextCancelledInLoop(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "ctx-cancel-loop-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test cancel in loop", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// A response that won't trigger shouldTerminate (no "sovereign audit").
	nonTerminatingResp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{genai.Text("Executing next step.")},
				},
			},
		},
		UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
	}

	cancelCtx, cancel := context.WithCancel(context.Background())

	// The cancelOnSecondCallSession fires cancel() on the 2nd SendMessage call.
	// Iter 1: SendMessage (call 1) → response → executePhase → loop.
	// Iter 2: select (not cancelled) → SendMessage (call 2, fires cancel) → response → executePhase → loop.
	// Iter 3: select (ctx IS cancelled) → returns context.Canceled ✅
	mockSession := &cancelOnSecondCallSession{
		Response: nonTerminatingResp,
		cancel:   cancel,
	}
	engine.genaiClient = &MockClient{
		Model: &MockModel{ChatSession: mockSession},
	}

	ctx := NewAgentContext(cancelCtx, taskID, &AgentDefinition{
		Name:     "cancel-loop-agent",
		ModelID:  ModelFlash,
		MaxSteps: 10,
	})

	err = engine.Execute(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestEngine_Execute_SendMessageError(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "send-err-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test send message error", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	expectedErr := errors.New("simulated send message error")
	mockClt := &MockClient{
		Model: &MockModel{
			Session: &MockSession{
				Err: expectedErr,
			},
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "send-err-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "generation failed:") {
		t.Errorf("expected 'generation failed:' error, got %v", err)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped simulated error, got %v", err)
	}
}

func TestEngine_Execute_SubTaskOrchestrationFailure(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	parentID := "parent-task-fail"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		parentID, "test decomposition failure", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	tmpDir := t.TempDir()
	runCmd := func(args ...string) {
		c := exec.CommandContext(context.Background(), "git", args...)
		c.Dir = tmpDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v (output: %s)", args, err, out)
		}
	}
	runCmd("init")
	runCmd("config", "user.email", "test@example.com")
	runCmd("config", "user.name", "Test User")

	_ = os.WriteFile(filepath.Join(tmpDir, "dummy"), []byte("data"), 0644)
	runCmd("add", ".")
	runCmd("commit", "-m", "initial")

	registry := NewRegistry()
	RegisterCoreTools(registry, db)

	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	regMgr, err := NewRegistryManager(db)
	if err != nil {
		t.Fatalf("failed to create registry manager: %v", err)
	}
	shield := NewGitShield(tmpDir, validator)
	dispatcher, err := NewDispatcher(regMgr, shield, db)
	if err != nil {
		t.Fatalf("failed to create dispatcher: %v", err)
	}
	engine.SetDispatcher(dispatcher)

	// Model returns a decomposition subtask requesting a capability no specialist has ("unknown-capability")
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking...</think>"),
								genai.FunctionCall{
									Name: "sentinel:decompose",
									Args: map[string]interface{}{
										"subtasks": []interface{}{
											map[string]interface{}{
												"description": "Fail feature",
												"branch_name": "subtask-branch-fail",
												"capabilities": []interface{}{"unknown-capability"},
											},
										},
									},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 50},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Failed as expected. Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 50},
			},
		},
	}

	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), parentID, &AgentDefinition{
		Name:     "dispatch-fail-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestEngine_Execute_HardGateValidationFailureCommand(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "validation-fail-cmd-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test validation cmd failure path", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	RegisterCoreTools(registry, db)

	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Model returns a tool call with forbidden command character in arguments
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking...</think>"),
								genai.FunctionCall{
									Name: "read_file",
									Args: map[string]interface{}{
										"path": "dummy.txt",
										"command": "rm -rf /; echo",
									},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "validation-fail-cmd-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

type dummyTool struct{}

func (t *dummyTool) Name() string { return "dummy_tool" }
func (t *dummyTool) Description() string { return "dummy description" }
func (t *dummyTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{Name: t.Name(), Description: t.Description()}
}
func (t *dummyTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	return nil
}
func (t *dummyTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	return "success", nil
}

func TestEngine_Execute_SuccessfulToolCallNotDecompose(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "dummy-tool-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test dummy tool path", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	registry.Tools["dummy_tool"] = &dummyTool{}

	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Let's call dummy tool.</think>"),
								genai.FunctionCall{
									Name: "dummy_tool",
									Args: map[string]interface{}{},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "dummy-tool-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestEngine_Execute_ToolExecutionFailure(t *testing.T) {
	setupTestCwd(t)
	db := testutil.SetupTestDB(t)
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	taskID := "tool-exec-fail-task"
	_, err := db.Conn.ExecContext(context.Background(),
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		taskID, "test tool execution error", "IN_PROGRESS", "T2", "",
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	registry := NewRegistry()
	RegisterCoreTools(registry, db)

	auth := &mockAuthProvider{key: "test-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	engine, err := NewEngine(registry, auth, validator, db)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Model returns a valid path format but file does not exist, causing read_file to fail during execution
	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("<think>Thinking...</think>"),
								genai.FunctionCall{
									Name: "read_file",
									Args: map[string]interface{}{
										"path": "nonexistent_file_xyz.txt",
									},
								},
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Sovereign Audit Report: Done."),
							},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 10},
			},
		},
	}
	mockClt := &MockClient{
		Model: &MockModel{
			Session: mockSession,
		},
	}
	engine.genaiClient = mockClt

	ctx := NewAgentContext(context.Background(), taskID, &AgentDefinition{
		Name:     "tool-exec-fail-agent",
		ModelID:  ModelFlash,
		MaxSteps: 5,
	})

	err = engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}



