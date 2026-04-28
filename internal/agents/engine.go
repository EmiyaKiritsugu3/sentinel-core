package agents

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/sync/errgroup"
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
	Registry *Registry
}

// NewEngine initializes a new agent engine.
func NewEngine(r *Registry) *Engine {
	return &Engine{Registry: r}
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
		// TODO: Implement LLM integration (Gemini Pro)
		log.Printf("[PHASE: THINKING] Step %d/%d", ctx.Budget.StepsTaken, ctx.Budget.MaxSteps)

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
