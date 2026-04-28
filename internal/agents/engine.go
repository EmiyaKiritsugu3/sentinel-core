package agents

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

// Tool defines the interface for agent capabilities.
type Tool interface {
	Name() string
	Description() string
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
	Registry     *Registry
	genaiClient  *genai.Client
	authProvider AuthProvider
}

// NewEngine initializes a new agent engine.
func NewEngine(r *Registry, auth AuthProvider) (*Engine, error) {
	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("engine: failed to get API key: %w", err)
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("engine: failed to create genai client: %w", err)
	}

	return &Engine{
		Registry:     r,
		genaiClient:  client,
		authProvider: auth,
	}, nil
}

// Close releases engine resources.
func (e *Engine) Close() error {
	if e.genaiClient != nil {
		return e.genaiClient.Close()
	}
	return nil
}

// callLLM executes a real call to the Gemini API.
func (e *Engine) callLLM(ctx *AgentContext, prompt string) (string, error) {
	model := e.genaiClient.GenerativeModel(ctx.ActiveModel)
	
	// Set reasonable defaults for generation
	model.SetTemperature(float32(ctx.Definition.Temperature))

	resp, err := model.GenerateContent(ctx.Context, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini: failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", fmt.Errorf("gemini: empty response from model")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			sb.WriteString(string(text))
		}
	}

	return sb.String(), nil
}

// Execute starts the execution of a subagent for a given task.
func (e *Engine) Execute(ctx *AgentContext) error {
	defer ctx.Cancel()

	log.Printf("[SENTINEL] Starting agent '%s' for state '%s'", ctx.Definition.Name, ctx.StateID)

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

		// 2. Thinking (Phase 2)
		log.Printf("[PHASE: THINKING] Step %d/%d", ctx.Budget.StepsTaken, ctx.Budget.MaxSteps)
		
		// TODO: Construct full prompt with history and system prompt
		prompt := "Focus on achieving the task goal. Respond with a technical action plan."
		
		response, err := e.callLLM(ctx, prompt)
		if err != nil {
			return fmt.Errorf("thinking phase failed: %w", err)
		}
		log.Printf("[SENTINEL] LLM Response: %s", response)

		// 3. Critique (Phase 3)
		// TODO: Implement local verification (Gemini Flash)
		
		// 4. Action (Phase 4)
		// For now, we simulate a successful completion until LLM is wired up.
		log.Printf("[PHASE: ACTION] Decision: Finalizing task.")
		
		// 5. Execution (Phase 5 - Concurrent Tool Execution)
		if err := e.executeTools(ctx, nil); err != nil {
			log.Printf("[SENTINEL] Tool execution error: %v", err)
			ctx.FailureCount++

			if e.shouldEscalate(ctx) {
				e.escalate(ctx)
				strategy, err := e.runPACDeliberation(ctx)
				if err != nil {
					return fmt.Errorf("PAC deliberation failed: %w", err)
				}
				ctx.Strategy = strategy
				log.Printf("[PAC] New Sovereign Strategy: %s", ctx.Strategy)
				// Continue the loop with the new strategy
				continue
			}

			return fmt.Errorf("tool execution failed after %d attempts: %w", ctx.FailureCount, err)
		}

		// 6. Post-Processing (Phase 6)
		// Simulating loop termination
		log.Printf("[PHASE: POST-PROCESS] Task state updated.")
		break
	}

	log.Printf("[SENTINEL] Agent '%s' completed successfully.", ctx.Definition.Name)
	return nil
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

// executeTools runs multiple tools in parallel using errgroup (Standard #06).
func (e *Engine) executeTools(ctx *AgentContext, toolCalls []map[string]interface{}) error {
	if len(toolCalls) == 0 {
		return nil
	}

	g, gCtx := errgroup.WithContext(ctx.Context)

	for _, call := range toolCalls {
		call := call // capture range variable
		g.Go(func() error {
			name, ok := call["name"].(string)
			if !ok {
				return fmt.Errorf("missing tool name in call")
			}

			tool, exists := e.Registry.Tools[name]
			if !exists {
				return fmt.Errorf("tool not found: %s", name)
			}

			args, _ := call["args"].(map[string]interface{})
			result, err := tool.Execute(gCtx, args)
			if err != nil {
				return fmt.Errorf("tool '%s' execution error: %w", name, err)
			}

			log.Printf("[TOOL: %s] Result: %s", name, result)
			return nil
		})
	}

	return g.Wait()
}
