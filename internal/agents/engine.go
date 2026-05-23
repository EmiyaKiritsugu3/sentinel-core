package agents

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/math"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/generative-ai-go/genai"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

// DefaultTokenPrice is the cost per token used for API cost calculation.
const DefaultTokenPrice = 0.00001

// Model constants for Gemini model selection.
const (
	ModelFlash = "gemini-1.5-flash"
	ModelPro   = "gemini-1.5-pro"
)

// pacAngleLogMsg is the log message key used for PAC angle analysis.
const pacAngleLogMsg = "PAC angle analysis"

// Tool defines the interface for agent capabilities.
type Tool interface {
	Name() string
	Description() string
	Definition() *genai.FunctionDeclaration
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// Registry manages available agents and tools.
type Registry struct {
	mu     sync.RWMutex
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

// GetTool returns a tool by name with read lock.
func (r *Registry) GetTool(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.Tools[name]
	return t, ok
}

// SetTool sets a tool by name with write lock.
func (r *Registry) SetTool(name string, t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Tools[name] = t
}

// ToolsSnapshot returns a copy of the tools map for safe concurrent iteration.
func (r *Registry) ToolsSnapshot() map[string]Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snap := make(map[string]Tool, len(r.Tools))
	for k, v := range r.Tools {
		snap[k] = v
	}
	return snap
}

// Engine orchestrates the 6-Phase ReAct loop for subagents.
type Engine struct {
	registry      *Registry
	genaiClient   bridge.GenaiClient
	authProvider  AuthProvider
	promptFactory *bridge.Factory
	validator     *reflect.Validator
	dispatcher    *Dispatcher
	db            *sqlite.DB
}

// newSDKClientFunc wraps bridge.NewSDKClient and can be overridden in tests
// to inject failures that are otherwise unreachable through the real SDK.
var newSDKClientFunc = bridge.NewSDKClient

// newFactoryFunc wraps bridge.NewFactory and can be overridden in tests
// to make newEngineFromComponents fail from within NewEngine without
// corrupting the real DB state.
var newFactoryFunc = bridge.NewFactory

// NewEngine initializes a new agent engine.
func NewEngine(r *Registry, auth AuthProvider, v *reflect.Validator, db *sqlite.DB) (*Engine, error) {
	if r == nil {
		return nil, fmt.Errorf("engine: nil registry")
	}
	if auth == nil {
		return nil, fmt.Errorf("engine: nil auth provider")
	}
	if v == nil {
		return nil, fmt.Errorf("engine: nil validator")
	}
	if err := sqlite.ValidateDB(db, "engine"); err != nil {
		return nil, err
	}

	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("engine: failed to get API key: %w", err)
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("engine: failed to create genai client: %w", err)
	}

	sdkClt, err := newSDKClientFunc(client)
	if err != nil {
		if client != nil {
			_ = client.Close()
		}
		return nil, fmt.Errorf("engine: failed to wrap sdk client: %w", err)
	}

	e, err := newEngineFromComponents(r, sdkClt, auth, v, db)
	if err != nil {
		_ = sdkClt.Close()
		return nil, err
	}
	return e, nil
}

// newEngineFromComponents wires a pre-built GenaiClient into the engine.
// Separated from NewEngine to allow unit testing of the classifier/factory
// initialization paths without a live genai.Client.
func newEngineFromComponents(r *Registry, clt bridge.GenaiClient, auth AuthProvider, v *reflect.Validator, db *sqlite.DB) (*Engine, error) {
	closeOnErr := true
	defer func() {
		if closeOnErr && clt != nil {
			_ = clt.Close()
		}
	}()

	geminiClassifier, err := bridge.NewGeminiClassifier(clt)
	if err != nil {
		return nil, fmt.Errorf("engine: failed to create gemini classifier: %w", err)
	}

	classifier := bridge.NewIntentClassifier(geminiClassifier, 0.60)
	factory, err := newFactoryFunc(db, classifier)
	if err != nil {
		return nil, fmt.Errorf("engine: failed to create prompt factory: %w", err)
	}

	closeOnErr = false
	return &Engine{
		registry:      r,
		genaiClient:   clt,
		authProvider:  auth,
		promptFactory: factory,
		validator:     v,
		db:            db,
	}, nil
}

// Close releases engine resources.
func (e *Engine) Close() error {
	if e.genaiClient != nil {
		return e.genaiClient.Close()
	}
	return nil
}

// SetDispatcher wires the sub-task dispatcher for orchestration (Phase 5.8+).
func (e *Engine) SetDispatcher(d *Dispatcher) {
	e.dispatcher = d
}

// DB returns the engine's database handle for read-only access.
func (e *Engine) DB() *sqlite.DB {
	return e.db
}

// Registry returns the engine's agent/tool registry for read-only access.
func (e *Engine) Registry() *Registry {
	return e.registry
}

func (e *Engine) getGenaiTools() []*genai.Tool {
	var decls []*genai.FunctionDeclaration
	for _, t := range e.registry.ToolsSnapshot() {
		decls = append(decls, t.Definition())
	}
	if len(decls) == 0 {
		return nil
	}
	return []*genai.Tool{{FunctionDeclarations: decls}}
}

// Execute starts the execution of a subagent for a given task.
func (e *Engine) Execute(ctx *AgentContext) (retErr error) {
	if err := sqlite.ValidateDB(e.db, "engine"); err != nil {
		return err
	}

	// Check budget before any I/O — MaxSteps == 0 means zero budget.
	if ctx.Definition.MaxSteps <= 0 {
		return fmt.Errorf("agent budget exceeded (MaxSteps: %d)", ctx.Definition.MaxSteps)
	}

	if e.promptFactory == nil {
		return fmt.Errorf("engine: prompt factory is nil")
	}
	if e.genaiClient == nil {
		return fmt.Errorf("engine: genai client is nil")
	}

	defer ctx.Cancel()

	ctx.StartTime = time.Now()

	slog.Info("starting agent", "agent", ctx.Definition.Name, "task", ctx.StateID)

	// Phase 1: Initialization
	session, currentParts, priorTrust, err := e.initSession(ctx)
	if err != nil {
		return err
	}

	priorSuccesses, priorTotal, _, _ := readPriorTrust(ctx.Context, e.db, ctx.Definition.Name)

	defer func() {
		_ = persistTrust(ctx.Context, e.db, ctx.Definition.Name, priorSuccesses, priorTotal, retErr == nil)
	}()

	// Phase 2-6: ReAct Loop
	for {
		// Phase 2: Pre-check (Budget & Context)
		if ctx.Budget.IncSteps() {
			return fmt.Errorf("agent budget exceeded (MaxSteps: %d)", ctx.Definition.MaxSteps)
		}

		select {
		case <-ctx.Context.Done():
			return ctx.Context.Err()
		default:
		}

		// Phase 3: Generation (Thinking & Action Decision)
		resp, err := session.SendMessage(ctx.Context, currentParts...)
		if err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		currentParts, err = e.processResponse(ctx, resp, priorTrust)
		if err != nil {
			return err
		}
		if currentParts != nil {
			continue
		}

		// Phase 4: Parse response parts
		toolCalls, textResponses := parseResponseParts(resp)

		if len(textResponses) > 0 {
			slog.Info("agent response", "text", strings.Join(textResponses, "\n"))
		}

		// Phase 5: Check for Termination (Final Sovereign Audit)
		if shouldTerminate(toolCalls, textResponses) {
			slog.Info("termination detected", "reason", "audit_report")
			break
		}

		// Phase 6: Tool Execution & Orchestration
		currentParts, err = e.executePhase(ctx, toolCalls)
		if err != nil {
			return err
		}
	}

	slog.Info("agent completed", "agent", ctx.Definition.Name)

	return e.finalize(ctx, priorTrust)
}

// initSession sets up the generative model session and initial prompt.
func (e *Engine) initSession(ctx *AgentContext) (bridge.MessageSender, []genai.Part, float64, error) {
	payload, err := e.promptFactory.GeneratePayload(ctx.Context, ctx.StateID, ctx.Definition.SystemPrompt)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("engine: failed to generate prompt payload: %w", err)
	}

	model := e.genaiClient.GenerativeModel(ctx.ActiveModel)
	model.SetTemperature(float32(ctx.Definition.Temperature))
	model.SetSystemInstructionContent(genai.NewUserContent(genai.Text(payload.SystemInstruction)))
	model.SetTools(e.getGenaiTools())

	session := model.StartChat()
	initialPrompt := fmt.Sprintf("TASK OBJECTIVE: %s\n\nSURGICAL CONTEXT:%s", payload.TaskDescription, payload.SurgicalContext)
	currentParts := []genai.Part{genai.Text(initialPrompt)}

	_, _, priorTrust, _ := readPriorTrust(ctx.Context, e.db, ctx.Definition.Name)

	return session, currentParts, priorTrust, nil
}

// processResponse handles token accounting and entropy gate checks.
// Returns parts to send if a gate intervention occurred, nil if normal flow should continue,
// or an error if the response is empty.
func (e *Engine) processResponse(ctx *AgentContext, resp *genai.GenerateContentResponse, priorTrust float64) ([]genai.Part, error) {
	if resp.UsageMetadata != nil {
		count := int(resp.UsageMetadata.TotalTokenCount)
		ctx.TokensUsed += count
		ctx.APICost += float64(count) * DefaultTokenPrice
		ctx.Budget.AddTokens(count)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		content := resp.Candidates[0].Content

		stepActionTokens, stepThoughtTokens := countThoughtActionTokens(content.Parts)
		ctx.ActionTokens += stepActionTokens
		ctx.ThoughtTokens += stepThoughtTokens

		stepLambda := math.CalculateLambda(stepActionTokens, stepThoughtTokens)
		lambda := math.CalculateLambda(ctx.ActionTokens, ctx.ThoughtTokens)

		if ctx.Definition.MaxLambda != nil {
			effectiveMaxLambda := *ctx.Definition.MaxLambda * math.TrustToDynamicLambda(priorTrust)
			if intervene, msg := checkGateA(lambda, effectiveMaxLambda); intervene {
				ctx.Budget.IncSteps()
				ctx.PreviousLambda = stepLambda
				return []genai.Part{genai.Text(msg)}, nil
			}
		}

		newDivCount, intervene, msg := checkGateA5(stepLambda, ctx.PreviousLambda, ctx.DivergenceCount)
		ctx.DivergenceCount = newDivCount
		if intervene {
			ctx.Budget.IncSteps()
			ctx.PreviousLambda = stepLambda
			return []genai.Part{genai.Text(msg)}, nil
		}
		ctx.PreviousLambda = stepLambda
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("gemini: empty response from model")
	}

	return nil, nil
}

// parseResponseParts extracts tool calls and text responses from model output.
func parseResponseParts(resp *genai.GenerateContentResponse) (toolCalls []map[string]interface{}, textResponses []string) {
	content := resp.Candidates[0].Content
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
	return toolCalls, textResponses
}

// executePhase runs tool calls, handles escalation, and processes sub-task orchestration.
// Returns the next parts to send to the model, or a default prompt if no tools were called.
func (e *Engine) executePhase(ctx *AgentContext, toolCalls []map[string]interface{}) ([]genai.Part, error) {
	if len(toolCalls) == 0 {
		return []genai.Part{genai.Text("Strategy confirmed. If complete, provide the Sovereign Audit Report. Otherwise, execute the next tool.")}, nil
	}

	slog.Info("executing tools", "count", len(toolCalls))
	results, err := e.executeToolsWithResults(ctx, toolCalls)
	if err != nil {
		slog.Error("tool execution error", "error", err)
		ctx.FailureCount++

		if e.shouldEscalate(ctx) {
			e.escalate(ctx)
			return nil, nil // Re-loop will reconfigure model on next generation
		}
		return []genai.Part{genai.Text(fmt.Sprintf("ERROR: Tool execution failed: %v. Please adjust your strategy.", err))}, nil
	}

	// Check if sentinel:decompose was called to trigger sub-task processing
	if hasDecompose(toolCalls) && e.dispatcher != nil {
		slog.Info("processing sub-tasks", "task", ctx.StateID)
		if err := e.processSubTasks(ctx); err != nil {
			slog.Error("orchestration failed", "error", err)
			return []genai.Part{genai.Text(fmt.Sprintf("ERROR: Orchestration failed: %v. Please adjust your decomposition strategy.", err))}, nil
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
	return responseParts, nil
}

// finalize records execution metrics and timing.
func (e *Engine) finalize(ctx *AgentContext, priorTrust float64) error {
	ctx.EndTime = time.Now()
	latency := float64(ctx.EndTime.Sub(ctx.StartTime).Milliseconds())
	return persistMetrics(ctx.Context, e.db, ctx.StateID, ctx.TokensUsed, ctx.APICost, latency, priorTrust)
}

// hasDecompose checks whether any tool call is sentinel:decompose.
func hasDecompose(toolCalls []map[string]interface{}) bool {
	for _, call := range toolCalls {
		if name, ok := call["name"].(string); ok && name == "sentinel:decompose" {
			return true
		}
	}
	return false
}

// processSubTasks handles the KISS sequential execution of pending sub-tasks.
func (e *Engine) processSubTasks(ctx *AgentContext) error {
	query := "SELECT id, parent_task_id, description, status, branch_name, required_capabilities FROM sub_tasks WHERE parent_task_id = ? AND status = 'PENDING'"
	rows, err := e.db.Conn.QueryContext(ctx.Context, query, ctx.StateID)
	if err != nil {
		return fmt.Errorf("engine: failed to query sub-tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var pending []SubTask
	for rows.Next() {
		var st SubTask
		var capsJSON string
		if err := rows.Scan(&st.ID, &st.ParentTaskID, &st.Description, &st.Status, &st.BranchName, &capsJSON); err != nil {
			return fmt.Errorf("engine: failed to scan sub-task: %w", err)
		}
		subCaps, err := unmarshalCapabilities(capsJSON)
		if err != nil {
			return fmt.Errorf("engine: failed to unmarshal capabilities for sub-task %s: %w", st.ID, err)
		}
		st.RequiredCapabilities = subCaps
		pending = append(pending, st)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("engine: row iteration error: %w", err)
	}

	for _, st := range pending {
		slog.Info("dispatching sub-task", "id", st.ID, "description", st.Description)
		if err := e.dispatcher.Dispatch(ctx.Context, &st); err != nil {
			return fmt.Errorf("engine: failed to dispatch sub-task %s: %w", st.ID, err)
		}

		// KISS: For now, we simulate success or wait for manual confirmation?
		// In a real autonomous loop, we would start another Engine instance here.
		// For Phase 5.8, we just mark as DISPATCHED and let the user know.
		slog.Info("sub-task dispatched", "id", st.ID, "worktree", st.WorktreePath)
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
			tool, exists := e.registry.GetTool(name)
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

// isExplicitThoughtBlock checks if the text starts with a thought block marker.
// ⚡ Bolt Optimization: Avoids strings.TrimSpace to prevent O(N) scanning of trailing
// whitespace on large text blocks. We only care about leading whitespace.
func isExplicitThoughtBlock(text string) bool {
	trimmed := strings.TrimLeftFunc(text, unicode.IsSpace)
	return strings.HasPrefix(trimmed, "<think>") || strings.HasPrefix(trimmed, "```thought")
}

// PACRecommendation represents a single angle's deliberation outcome.
type PACRecommendation int

const (
	// PACProceed means continue with the current approach.
	PACProceed PACRecommendation = iota
	// PACSimplify means reduce scope or simplify the approach (YAGNI).
	PACSimplify
	// PACPivot means the current technical approach is flawed and needs changing.
	PACPivot
	// PACEscalate means the situation requires a more powerful model or human intervention.
	PACEscalate
)

func (r PACRecommendation) String() string {
	switch r {
	case PACProceed:
		return "proceed"
	case PACSimplify:
		return "simplify"
	case PACPivot:
		return "pivot"
	case PACEscalate:
		return "escalate"
	default:
		return "unknown"
	}
}

// PACResult holds the tripartite deliberation outcome.
type PACResult struct {
	AngleA PACRecommendation // Minimalist (YAGNI)
	AngleB PACRecommendation // Structuralist (Plan Pivot)
	AngleC PACRecommendation // Auditor (Compliance)
	Final  PACRecommendation // Weighted consensus (worst-case wins)
	Reason string            // Human-readable explanation
}

// runPACDeliberation executes the tripartite deliberation (Minimalist, Structuralist, Auditor).
// Each angle analyzes the current agent context using existing metrics to produce
// a recommendation. The final decision uses worst-case-wins escalation semantics.
func (e *Engine) runPACDeliberation(ctx *AgentContext) (string, error) {
	slog.Info("PAC deliberation started", "angles", 3)

	result := PACResult{}

	// Phase 1: Angle A (Minimalist) - YAGNI check
	// Analyzes whether the current approach is over-engineered.
	result.AngleA = e.pacAngleMinimalist(ctx)
	slog.Info(pacAngleLogMsg, "angle", "A", "approach", "minimalist", "recommendation", result.AngleA)

	// Phase 2: Angle B (Structuralist) - Plan pivot check
	// Analyzes whether the current technical approach is fundamentally flawed.
	result.AngleB = e.pacAngleStructuralist(ctx)
	slog.Info(pacAngleLogMsg, "angle", "B", "approach", "structural_pivot", "recommendation", result.AngleB)

	// Phase 3: Angle C (Auditor) - Environment & compliance check
	// Analyzes resource constraints and escalation state.
	result.AngleC = e.pacAngleAuditor(ctx)
	slog.Info(pacAngleLogMsg, "angle", "C", "approach", "compliance", "recommendation", result.AngleC)

	// Determine final recommendation: worst-case wins.
	// Escalate > Pivot > Simplify > Proceed
	result.Final = pacWorstCase(result.AngleA, result.AngleB, result.AngleC)

	// Apply the deliberation outcome.
	switch result.Final {
	case PACEscalate:
		e.escalate(ctx)
		result.Reason = "PAC consensus: escalation required — resource constraints or persistent failures on pro model"
	case PACPivot:
		ctx.Strategy = "sovereign-pivot"
		result.Reason = "PAC consensus: structural pivot required — high divergence or repeated failures indicate flawed approach"
	case PACSimplify:
		ctx.Strategy = "simplify"
		result.Reason = "PAC consensus: simplification recommended — over-engineering detected, reduce scope to essential path"
	default:
		result.Reason = "PAC consensus: proceed with current approach — all angles green"
	}

	slog.Info("PAC deliberation complete", "final", result.Final, "reason", result.Reason)
	return result.Reason, nil
}

// pacAngleMinimalist (Angle A) checks whether the approach is over-engineered.
// Triggers Simplify when the agent is thinking too much relative to doing,
// or when the step budget is running low.
func (e *Engine) pacAngleMinimalist(ctx *AgentContext) PACRecommendation {
	if ctx.Budget == nil {
		return PACProceed
	}

	// High thought/action ratio → over-engineering signal
	if ctx.ActionTokens > 0 && float64(ctx.ThoughtTokens)/float64(ctx.ActionTokens) > 2.0 {
		return PACSimplify
	}

	// Step budget more than 70% consumed → need a simpler, faster approach
	if ctx.Budget.MaxSteps > 0 && ctx.Budget.StepsTaken > (ctx.Budget.MaxSteps*70/100) {
		return PACSimplify
	}

	return PACProceed
}

// pacAngleStructuralist (Angle B) checks whether the technical approach needs pivoting.
// Triggers Pivot when reasoning diverges consecutively or failures accumulate
// below the escalation threshold.
func (e *Engine) pacAngleStructuralist(ctx *AgentContext) PACRecommendation {
	// Consecutive divergence → approach is heading in the wrong direction
	if ctx.DivergenceCount >= 2 {
		return PACPivot
	}

	// Repeated failures not yet at escalation threshold → approach may be wrong
	if ctx.FailureCount >= 2 {
		return PACPivot
	}

	return PACProceed
}

// pacAngleAuditor (Angle C) checks resource and compliance constraints.
// Triggers Escalate when already on the pro model and still failing,
// or when the step budget is critically exhausted.
func (e *Engine) pacAngleAuditor(ctx *AgentContext) PACRecommendation {
	// Already escalated to pro but still failing → need external intervention
	if ctx.ActiveModel == ModelPro && ctx.FailureCount > 0 {
		return PACEscalate
	}

	// Step budget critically exhausted (>90%) → need escalation for resolution
	if ctx.Budget != nil && ctx.Budget.MaxSteps > 0 && ctx.Budget.StepsTaken > (ctx.Budget.MaxSteps*90/100) {
		return PACEscalate
	}

	return PACProceed
}

// pacWorstCase returns the most severe recommendation among the three angles.
func pacWorstCase(a, b, c PACRecommendation) PACRecommendation {
	worst := a
	if b > worst {
		worst = b
	}
	if c > worst {
		worst = c
	}
	return worst
}

func (e *Engine) shouldEscalate(ctx *AgentContext) bool {
	return ctx.FailureCount >= 3 && ctx.ActiveModel == ModelFlash
}

func (e *Engine) escalate(ctx *AgentContext) {
	slog.Warn("PAC escalation", "model", ModelPro)
	ctx.ActiveModel = ModelPro
	// Reset failure count after escalation for the new model session
	ctx.FailureCount = 0
}
