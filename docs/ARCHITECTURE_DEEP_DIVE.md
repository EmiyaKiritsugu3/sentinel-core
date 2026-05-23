# Sentinel Core - Architecture Deep Dive
*Date: 2026-05-23*

## Executive Summary
Sentinel Core is a Go-based Deterministic Architectural Governance & Context Compiler Engine. It acts as a "Warden" for AI-native software development. By reading the abstract syntax tree (AST) of a codebase and storing it in a SQLite database, it enforces "Hard Gates" to govern AI modifications to code, preventing hallucinations or architectural regressions.

## Codebase Overview
The project is built mainly in Go (99 source files, 43 test files) with a small, React-based web frontend for the "LiveView". The backend acts as a single static binary.

### 1. The Core Graph Engine (`internal/graph`)
**Status: Highly Developed (Core Functionality)**
- **AST Parsing**: Scans Go code (with support for extension to Tree-sitter) to identify nodes (functions, structs, interfaces) and edges (calls, implements, imports).
- **SQLite Ledger**: All state is tracked in a local `.sentinel/graph.db`. This includes the AST mapping, tasks, agent trust scores, and cognitive patterns.
- **Incremental Scanning**: Uses a `Skip-if-Hash-Match` pattern based on file hashing to avoid unnecessary reprocessing.
- **Thread-Safety**: Uses `sync.RWMutex` to manage parallel orchestration effectively.

### 2. Agent Orchestration (`internal/agents` & `internal/intake`)
**Status: Developed / Actively Maturing**
- **Sovereign Triad**:
  - *Warden*: Sentinel itself.
  - *Chief Engineer (Dispatcher)*: Handles complex routing (`internal/agents/dispatcher.go`).
  - *Operators (Sub-tasks)*: Ephemeral task execution.
- **Scout Protocol (Data-Driven Intent)**: Found in `internal/intake/disambiguator.go`. It analyzes the `graph.db` to identify "God Objects" or code hotspots to refine vague tasks into precise architectural plans.
- **Agent Definitions**: Configurations for specialists (e.g., `Sovereign Architect` via `gemini-1.5-pro`) exist under `internal/agents/definitions/`.

### 3. Verification and Gates (`internal/audit` & `internal/reflect`)
**Status: Robust & Functional**
- **Executable Contracts (ADRs)**: Every architectural change must be verified by a command (like `go test`). The `Audit Runner` (`internal/audit/runner.go`) securely executes these commands using shell-splitting and timeouts.
- **Sovereign Gates**:
  - *Technical Gate*: Build and test checks.
  - *Cognitive Entropy*: Validation ratios tracked in the database to prevent looping/hallucination.
  - *Nil Guard Hardening*: Specific Go patterns ensuring error/nil validations systematically exist.

### 4. CLI Interface (`cmd/sentinel`)
**Status: Mature**
Built with Cobra (`spf13/cobra`), it manages the entire workflow deterministically:
- `plan`: Translates goals to actionable tasks using disambiguation.
- `scan`: Invokes the graph engine.
- `visualize`: Builds Master Graph and C4 diagrams.
- `start`: Initiates the cognitive loop for a task.
- `audit`: Evaluates the output through the verification gates.
- `live`: Boots up the WebSocket Server and UI for real-time visualization.

### 5. LiveView Frontend (`web/` & `internal/liveview`)
**Status: Prototype / Early Development**
- **Backend API**: Found in `internal/liveview/server.go`, it exposes WebSockets and endpoints to track system state dynamically.
- **Frontend App**: Built with Vite + React + TypeScript. It is currently a basic template structure with `main.tsx` and `App.tsx` containing the UI foundations for observing Sentinel's live events.

## Conclusion & Next Steps
The backend governance engine is sophisticated, highly structured, and actively utilizes advanced architectural patterns for autonomous AI development (e.g., SQLite Ledgers, AST graph-based context scoping, Event Reconciliation). The system's rules are strictly defined via documents like `GEMINI.md`.

**Area of Growth**:
- The `web/` UI represents an emerging feature that requires alignment with the robust backend WebSocket output to provide the promised real-time observability.
