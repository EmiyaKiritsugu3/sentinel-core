package agents

import (
	"context"
	"sync"
)

// TokenBudget handles deterministic execution limits to prevent infinite loops (Standard #06).
type TokenBudget struct {
	mu         sync.RWMutex
	MaxTokens  int
	UsedTokens int
	MaxSteps   int
	StepsTaken int
}

// AddTokens increments the used token count thread-safely.
func (b *TokenBudget) AddTokens(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.UsedTokens += n
}

// IncSteps increments the step counter and returns true if the budget is exceeded.
func (b *TokenBudget) IncSteps() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.StepsTaken++
	return b.StepsTaken > b.MaxSteps || (b.MaxTokens > 0 && b.UsedTokens > b.MaxTokens)
}

// AgentDefinition represents the "Smart Agent Artifact" (.md + YAML).
type AgentDefinition struct {
	Name          string   `yaml:"name" validate:"required"`
	Version       string   `yaml:"version"`
	ModelID       string   `yaml:"model_id" validate:"required"`
	Temperature   float64  `yaml:"temperature"`
	MaxSteps      int      `yaml:"max_steps" validate:"required,min=1"`
	Capabilities  []string `yaml:"capabilities"`
	TierAccess    []string `yaml:"tier_access"`
	SystemPrompt  string   `yaml:"-"` // From Markdown body
}

// Message represents a single interaction in the agent's memory.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentContext encapsulates the runtime state of a subagent.
type AgentContext struct {
	StateID      string
	Definition   *AgentDefinition
	Budget       *TokenBudget
	Memory       []Message
	Context      context.Context
	Cancel       context.CancelFunc
	FailureCount int    // Track consecutive failures
	ActiveModel  string // Current model being used
	Strategy     string // Current technical strategy (Sovereign Pivot)
}

// NewAgentContext initializes a context with cancellation.
func NewAgentContext(ctx context.Context, stateID string, def *AgentDefinition) *AgentContext {
        c, cancel := context.WithCancel(ctx)
        return &AgentContext{
                StateID:    stateID,
                Definition: def,
                Budget: &TokenBudget{
                        MaxSteps: def.MaxSteps,
                },
                Context:     c,
                Cancel:      cancel,
                ActiveModel: def.ModelID,
        }
}

// Specialist represents a persistent autonomous agent in the registry.
type Specialist struct {
	ID                 string
	Name               string
	BasePersona        string
	CurrentPersonaPath string
	ReliabilityScore   float64
	Capabilities       []string
}

// Validator defines the security interface for path and command validation (Standard #10).
type Validator interface {
	ValidatePath(path string) error
	ValidateCommand(cmd string) error
}

