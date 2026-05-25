# internal/graph

Dependency graph engine with schema migration, multi-language scanning, incremental hash verification, and observer-based event broadcasting.

## Overview

The graph package maintains a SQLite-backed dependency graph of your codebase. It supports multi-language file scanning (Go AST, Tree-sitter), incremental updates via content hashing, and an observer pattern for real-time event propagation (used by the liveview WebSocket server).

## Schema

The `Migrate(ctx, db)` function creates and evolves the following tables:

| Table | Purpose |
|-------|---------|
| `nodes` | Code symbols (file, function, struct, interface) with line ranges and content hash |
| `edges` | Directed relationships: imports, calls, implements |
| `tasks` | Governance tasks with status, tier, verification command, and SME metrics |
| `audit_logs` | Verification gate execution history |
| `standards` | Semantic rule definitions |
| `specialist_registry` | Evolutionary agent personas |
| `sub_tasks` | Decomposed task units with worktree tracking |
| `agent_trust` | Bayesian trust scores per agent |
| `patterns` | Architectural patterns with FTS5 full-text search |
| `patterns_fts` | Virtual FTS5 table with insert/update/delete triggers |
| `knowledge_sessions` / `session_events` | Session debrief persistence |

SME migrations add columns (`latency_ms`, `tokens_used`, `api_cost`, `math_delta`) conditionally using `PRAGMA table_info`.

## Key Types

### `Engine`
Coordinates the scanning pipeline with worker pool, ignore filter, incremental hashing, and observer notifications.
- `NewEngine(db)` — validates DB, creates engine
- `RegisterScanner(scanner)` — registers per-extension file scanners
- `RegisterObserver(observer)` — subscribes to graph lifecycle events
- `ScanProject(ctx, root)` — walks project directory, scans files via 8-worker pool, persists nodes/edges, runs linker

### `FileScanner` (interface)
Each language scanner implements: `Scan(path) ScanResult`, `SupportedExtensions() []string`.

### `Observer` (interface)
Receives `Notify(GraphEvent)` for events: `SCAN_STARTED`, `NODE_UPSERTED`, `EDGE_CREATED`, `SCAN_COMPLETED`.

### `Node`, `Edge`, `ScanResult`
Core data types: `Node` (ID, name, type, file path, line range, hash), `Edge` (from, to, relation), `ScanResult` (nodes, edges, error).

## Built-in Scanners

- **GoScanner** (`scanner_go.go`): Uses `go/ast` and `go/parser`. Extracts functions, structs, interfaces, imports, and method calls from `.go` files.
- **TreeSitterScanner** (`scanner_treesitter.go`): Multi-language scanner using Tree-sitter grammars.

## Dependencies

- `pkg/sqlite` — DB abstraction and validation
- `pkg/utils` — ignore filter and hash calculation
- `modernc.org/sqlite` — CGo-free SQLite driver

## Usage

```go
db, _ := sqlite.Init()
graph.Migrate(context.Background(), db)

engine, _ := graph.NewEngine(db)
engine.RegisterScanner(graph.NewGoScanner())
if err := engine.ScanProject(context.Background(), "."); err != nil { /* handle */ }
```
