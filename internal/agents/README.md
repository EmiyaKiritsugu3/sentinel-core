# internal/agents

Agent engine implementing a 6-phase ReAct loop with PAC deliberation, tool execution, and sub-task dispatch.

## Overview

The agents package orchestrates AI-powered subagents through a deterministic execution loop. It manages agent definitions, token budgets, tool registration, and a tripartite deliberation system (PAC: Proceed, Simplify, Pivot, Escalate) that continuously evaluates agent strategy.

## Key Types

### `Engine`
Core orchestrator holding a registry, GenaiClient, AuthProvider, prompt factory, path/command validator, and dispatcher.
- `NewEngine(registry, auth, validator, db)` — bootstraps Gemini client and prompt factory
- `Execute(ctx *AgentContext)` — runs the full 6-phase ReAct loop
- `SetDispatcher(d *Dispatcher)` — wires sub-task orchestration
- `Close()` — releases SDK resources

### `Dispatcher`
Sub-task assignment and event reconciliation (write serializer). Manages specialist selection, worktree creation, and ledger persistence.
- `Dispatch(ctx, subTask)` — selects specialist, creates worktree, persists to DB
- `ReconcileEvents(ctx)` — reads atomic event files from `.sentinel/events/` and updates sub-task status

### `PAC Deliberation`
Tripartite decision system (`runPACDeliberation`):
- **Angle A (Minimalist)**: YAGNI check — high thought/action ratio or budget exhaustion
- **Angle B (Structuralist)**: Pivot check — consecutive divergence or repeated failures
- **Angle C (Auditor)**: Escalation check — pro model still failing or critical budget exhaustion

Worst-case wins: Escalate > Pivot > Simplify > Proceed.

### Registry, Tool, AgentContext
- `Registry` — thread-safe map of agents and tools (`sync.RWMutex`)
- `Tool` interface — `Name()`, `Description()`, `Definition()`, `Execute()`
- `AgentContext` — runtime state including budget, memory, Lyapunov divergence tracking, and metrics

## Tools (registered via `RegisterCoreTools`)

| Tool | Name | Purpose |
|------|------|---------|
| `ReadFileTool` | `read_file` | Read file with line range |
| `WriteFileTool` | `write_file` | Write file with AST validation (Gate B) |
| `ReplaceTool` | `replace` | String replacement with structural validation |
| `GrepSearchTool` | `grep_search` | Regex search across project files |
| `AuditTool` | `sentinel:audit` | Run Sovereign Validator |
| `RunTool` | `sentinel:run` | Execute safe shell commands |
| `ADRTool` | `sentinel:adr` | Generate Architectural Decision Records |
| `ScanTool` | `sentinel_scan` | Trigger graph engine scan |
| `DecomposeTool` | `sentinel:decompose` | Decompose task into sub-tasks (max 5) |

## Dependencies

- `internal/bridge` — AI client abstraction, prompt factory, intent classifier
- `internal/math` — Lyapunov divergence and trust score formulas
- `internal/knowledge` — session event recording via GlobalBuffer
- `internal/reflect` — path and command validation
- `pkg/sqlite` — ledger persistence
- `github.com/google/generative-ai-go/genai` — Gemini SDK
- `github.com/google/shlex` — shell command parsing

## Usage

```go
reg := agents.NewRegistry()
agents.RegisterCoreTools(reg, db)

auth := agents.NewFileAuthProvider(".sentinel/auth.json")
v, _ := reflect.NewValidator(db)

engine, err := agents.NewEngine(reg, auth, v, db)
if err != nil { /* handle */ }
defer engine.Close()

ctx := agents.NewAgentContext(context.Background(), "task-123", &agents.AgentDefinition{
    Name:        "architect",
    ModelID:     agents.ModelPro,
    MaxSteps:    10,
    SystemPrompt: "...",
})
if err := engine.Execute(ctx); err != nil { /* handle */ }
```
