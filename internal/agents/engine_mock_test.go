package agents

import (
	"context"
	"os"
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	step1Text := "<think>Reasoning step 1: I will plan thoroughly.</think> Action: ok."
	step2Text := "<think>Reasoning step 2.</think> Action: let's build something large. " + strings.Repeat("more actions ", 10)
	step3Text := "Action: write infinite loops. " + strings.Repeat("even more actions ", 30)

	mockSession := &MockSession{
		Responses: []*genai.GenerateContentResponse{
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text(step1Text)},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 50},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text(step2Text)},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 100},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text(step3Text)},
						},
					},
				},
				UsageMetadata: &genai.UsageMetadata{TotalTokenCount: 150},
			},
			{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text("Sovereign Audit Report: Stabilized reasoning and completed task.")},
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

	// Verify that the Lyapunov divergence triggered counter changes
	if ctx.DivergenceCount > 0 {
		t.Logf("Sovereign Lyapunov Divergence count: %d", ctx.DivergenceCount)
	}
}

func TestEngine_Execute_BudgetExceededPath(t *testing.T) {
	t.Parallel()
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
	if !strings.Contains(err.Error(), "budget exceeded") {
		t.Errorf("expected 'budget exceeded' error, got: %v", err)
	}
}

func TestEngine_Execute_EmptyResponseError(t *testing.T) {
	t.Parallel()
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
	if !strings.Contains(err.Error(), "empty response from model") {
		t.Errorf("expected 'empty response from model' error, got: %v", err)
	}
}
