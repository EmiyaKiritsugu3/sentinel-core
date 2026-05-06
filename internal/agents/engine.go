package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/generative-ai-go/genai"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

// Tool defines the interface for agent capabilities.
type Tool interface {
	Name() string
	Description() string
	Definition() *genai.FunctionDeclaration
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// Registry manages available agents and tools.
type Registry struct {
	Agents map[string]*AgentDefinition
	Tools  map[string]Tool
}

// NewRegistry initializes an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		Agents: make(map[string]*AgentDefinition),
		Tools:  make(map[string]Tool),
	}
}

// Engine orchestrates the 6-Phase ReAct loop for subagents.
type Engine struct {
	Registry      *Registry
	genaiClient   *genai.Client
	authProvider  AuthProvider
	promptFactory *bridge.Factory
	validator     *reflect.Validator
	Dispatcher    *Dispatcher // Added for Phase 5.8
	DB            *sqlite.DB
}

// NewEngine initializes a new agent engine.
func NewEngine(r *Registry, auth AuthProvider, v *reflect.Validator, db *sqlite.DB) (*Engine, error) {
	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("engine: failed to get API key: %w", err)
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("engine: failed to create genai client: %w", err)
	}

	geminiClassifier := bridge.NewGeminiClassifier(client)
	classifier := bridge.NewIntentClassifier(geminiClassifier, 0.60)
	factory := bridge.NewFactory(db, classifier)

	return &Engine{
		Registry:      r,
		genaiClient:   client,
		authProvider:  auth,
		promptFactory: factory,
		validator:     v,
		DB:            db,
	}, nil
}

// Close releases engine resources.
func (e *Engine) Close() error {
	if e.genaiClient != nil {
		return e.genaiClient.Close()
	}
	return nil
}

func (e *Engine) getGenaiTools() []*genai.Tool {
	var decls []*genai.FunctionDeclaration
	for _, t := range e.Registry.Tools {
		decls = append(decls, t.Definition())
	}
	if len(decls) == 0 {
		return nil
	}
	return []*genai.Tool{{FunctionDeclarations: decls}}
}

// Execute starts the execution of a subagent for a given task.
func (e *Engine) Execute(ctx *AgentContext) error {
	defer ctx.Cancel()

	log.Printf("[SENTINEL] Starting agent '%s' for task '%s'", ctx.Definition.Name, ctx.StateID)

	payload, err := e.promptFactory.GeneratePayload(ctx.Context, ctx.StateID, ctx.Definition.SystemPrompt)
	if err != nil {
		return fmt.Errorf("engine: failed to generate prompt payload: %w", err)
	}

	model := e.genaiClient.GenerativeModel(ctx.ActiveModel)
	model.SetTemperature(float32(ctx.Definition.Temperature))
	model.SystemInstruction = genai.NewUserContent(genai.Text(payload.SystemInstruction))
	model.Tools = e.getGenaiTools()

	session := model.StartChat()
	// Initial objective message
	initialPrompt := fmt.Sprintf("TASK OBJECTIVE: %s\n\nSURGICAL CONTEXT:%s", payload.TaskDescription, payload.SurgicalContext)

	currentParts := []genai.Part{genai.Text(initialPrompt)}

	for {
		// 1. Pre-check (Budget & Context)
		if ctx.Budget.IncSteps() {
			return fmt.Errorf("agent budget exceeded (MaxSteps: %d)", ctx.Definition.MaxSteps)
		}

		select {
		case <-ctx.Context.Done():
			return ctx.Context.Err()
		default:
		}

		// 2. Generation (Thinking & Action Decision)
		log.Printf("[PHASE: GENERATION] Step %d/%d", ctx.Budget.StepsTaken, ctx.Budget.MaxSteps)
		resp, err := session.SendMessage(ctx.Context, currentParts...)
		if err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			return fmt.Errorf("gemini: empty response from model")
		}

		content := resp.Candidates[0].Content
		var toolCalls []map[string]interface{}
		var textResponses []string

		for _, part := range content.Parts {
			if text, ok := part.(genai.Text); ok {
				textResponses = append(textResponses, string(text))
			}
			if call, ok := part.(genai.FunctionCall); ok {
				toolCalls = append(toolCalls, map[string]interface{}{
					"name": call.Name,
					"args": call.Args,
				})
			}
		}

		if len(textResponses) > 0 {
			log.Printf("[SENTINEL] Agent Response: %s", strings.Join(textResponses, "\n"))
		}

		// 3. Check for Termination (Final Sovereign Audit)
		if len(toolCalls) == 0 && strings.Contains(strings.Join(textResponses, ""), "Sovereign Audit Report") {
			log.Printf("[SENTINEL] Termination detected via Audit Report.")
			break
		}

		// 4. Tool Execution (Phase 5)
		if len(toolCalls) > 0 {
			log.Printf("[PHASE: EXECUTION] Running %d tool(s) in parallel...", len(toolCalls))
			results, err := e.executeToolsWithResults(ctx, toolCalls)
			if err != nil {
				log.Printf("[SENTINEL] Tool execution error: %v", err)
				ctx.FailureCount++

				if e.shouldEscalate(ctx) {
					e.escalate(ctx)
					// Re-configure model with escalated identity
					model = e.genaiClient.GenerativeModel(ctx.ActiveModel)
					// PAC pivot could be injected here
					continue
				}
				// Feed the error back to the model as a system failure
				currentParts = []genai.Part{genai.Text(fmt.Sprintf("ERROR: Tool execution failed: %v. Please adjust your strategy.", err))}
				continue
			}

			// KISS: Check if sentinel:decompose was called to trigger sub-task processing
			decomposed := false
			for _, call := range toolCalls {
				if call["name"].(string) == "sentinel:decompose" {
					decomposed = true
					break
				}
			}

			if decomposed && e.Dispatcher != nil {
				log.Printf("[PHASE: ORCHESTRATION] Processing sub-tasks for goal %s", ctx.StateID)
				if err := e.processSubTasks(ctx); err != nil {
					log.Printf("[SENTINEL] Orchestration failed: %v", err)
					currentParts = []genai.Part{genai.Text(fmt.Sprintf("ERROR: Orchestration failed: %v. Please adjust your decomposition strategy.", err))}
					continue
				}
			}

			// Format results as FunctionResponses for the next turn
			var responseParts []genai.Part
			for name, result := range results {
				responseParts = append(responseParts, genai.FunctionResponse{
					Name:     name,
					Response: map[string]interface{}{"result": result},
				})
			}
			// Important: Use FunctionResponse parts for the next SendMessage
			currentParts = responseParts
		} else {
			// No tool calls and no termination? Ask for next step or provide more context.
			currentParts = []genai.Part{genai.Text("Strategy confirmed. If complete, provide the Sovereign Audit Report. Otherwise, execute the next tool.")}
		}
	}

	log.Printf("[SENTINEL] Agent '%s' completed successfully.", ctx.Definition.Name)
	return nil
}

// processSubTasks handles the KISS sequential execution of pending sub-tasks.
func (e *Engine) processSubTasks(ctx *AgentContext) error {
	query := "SELECT id, parent_task_id, description, status, branch_name, required_capabilities FROM sub_tasks WHERE parent_task_id = ? AND status = 'PENDING'"
	rows, err := e.DB.Conn.QueryContext(ctx.Context, query, ctx.StateID)
	if err != nil {
		return fmt.Errorf("engine: failed to query sub-tasks: %w", err)
	}
	defer rows.Close()

	var pending []SubTask
	for rows.Next() {
		var st SubTask
		var capsJSON string
		if err := rows.Scan(&st.ID, &st.ParentTaskID, &st.Description, &st.Status, &st.BranchName, &capsJSON); err != nil {
			return fmt.Errorf("engine: failed to scan sub-task: %w", err)
		}
		if err := json.Unmarshal([]byte(capsJSON), &st.RequiredCapabilities); err != nil {
			return fmt.Errorf("engine: failed to unmarshal capabilities for sub-task %s: %w", st.ID, err)
		}
		pending = append(pending, st)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("engine: row iteration error: %w", err)
	}

	for _, st := range pending {
		log.Printf("[ORCHESTRATOR] Dispatching sub-task %s: %s", st.ID, st.Description)
		if err := e.Dispatcher.Dispatch(ctx.Context, &st); err != nil {
			return fmt.Errorf("engine: failed to dispatch sub-task %s: %w", st.ID, err)
		}

		// KISS: For now, we simulate success or wait for manual confirmation?
		// In a real autonomous loop, we would start another Engine instance here.
		// For Phase 5.8, we just mark as DISPATCHED and let the user know.
		log.Printf("[ORCHESTRATOR] Sub-task %s dispatched to worktree %s", st.ID, st.WorktreePath)
	}

	return nil
}

// executeToolsWithResults runs tools and returns their outputs indexed by tool name.
func (e *Engine) executeToolsWithResults(ctx *AgentContext, toolCalls []map[string]interface{}) (map[string]string, error) {
	results := make(map[string]string)
	var mu sync.Mutex
	g, gCtx := errgroup.WithContext(ctx.Context)

	for _, call := range toolCalls {
		call := call
		g.Go(func() error {
			name := call["name"].(string)
			tool, exists := e.Registry.Tools[name]
			if !exists {
				return fmt.Errorf("tool not found: %s", name)
			}

			args := call["args"].(map[string]interface{})

			// Hard Gate: Dynamic Argument Validation (Standard #10)
			for key, val := range args {
				if strVal, ok := val.(string); ok {
					switch key {
					case "path", "file", "filepath":
						if err := e.validator.ValidatePath(strVal); err != nil {
							return fmt.Errorf("hard gate: %w", err)
						}
					case "command", "cmd":
						if err := e.validator.ValidateCommand(strVal); err != nil {
							return fmt.Errorf("hard gate: %w", err)
						}
					}
				}
			}

			result, err := tool.Execute(gCtx, args)
			if err != nil {
				return err
			}

			mu.Lock()
			results[name] = result
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("engine: parallel execution failed: %w", err)
	}
	return results, nil
}

// runPACDeliberation executes the tripartite deliberation (Minimalist, Structuralist, Auditor).
func (e *Engine) runPACDeliberation(ctx *AgentContext) (string, error) {
	log.Printf("[PAC] Starting Tripartite Deliberation (3 Angles)")

	// Phase 1: Angle A (Minimalist) - YAGNI check
	log.Printf("[PAC: ANGLE A] Analyzing minimalist approach: Can we achieve the goal by deleting code or simplifying the requirement?")

	// Phase 2: Angle B (Structuralist) - Plan pivot check
	log.Printf("[PAC: ANGLE B] Analyzing structural plan pivot: Is the current architectural approach fundamentally flawed for this task?")

	// Phase 3: Angle C (Auditor) - Security & Environment check
	log.Printf("[PAC: ANGLE C] Analyzing system locks and compliance: Are there environment constraints or security blockers?")

	// In future phases, these will be real LLM calls.
	return "Sovereign Pivot Generated: Switching technical approach based on tripartite analysis.", nil
}

func (e *Engine) shouldEscalate(ctx *AgentContext) bool {
	return ctx.FailureCount >= 3 && ctx.ActiveModel == "gemini-1.5-flash"
}

func (e *Engine) escalate(ctx *AgentContext) {
	log.Printf("[PAC] Escalating to gemini-1.5-pro for deep deliberation.")
	ctx.ActiveModel = "gemini-1.5-pro"
	// Reset failure count after escalation for the new model session
	ctx.FailureCount = 0
}
