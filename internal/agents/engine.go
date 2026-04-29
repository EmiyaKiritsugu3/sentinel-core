package agents

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
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
}

// NewEngine initializes a new agent engine.
func NewEngine(r *Registry, auth AuthProvider, factory *bridge.Factory, v *reflect.Validator) (*Engine, error) {
	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("engine: failed to get API key: %w", err)
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("engine: failed to create genai client: %w", err)
	}

	return &Engine{
		Registry:      r,
		genaiClient:   client,
		authProvider:  auth,
		promptFactory: factory,
		validator:     v,
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

	payload, err := e.promptFactory.GeneratePayload(ctx.StateID, ctx.Definition.SystemPrompt)
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
		return nil, err
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
