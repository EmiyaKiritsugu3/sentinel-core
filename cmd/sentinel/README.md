# cmd/sentinel

Main entry point and CLI command tree for the Sentinel governance engine.

## Overview

Sentinel ships as a single static binary with 11 subcommands organized under the root `sentinel` command. Each subcommand registers itself via the `registry` package's `init()` pattern, enabling dynamic command tree construction with database dependency injection.

## Architecture

```
cmd/sentinel/main.go              → sqlite.Init() → commands.Execute(db)
cmd/sentinel/commands/root.go     → NewRootCmd(db) → iterates registry.GetCommands()
cmd/sentinel/commands/*.go        → init() calls registry.Register(factory)
```

The root command validates the DB handle, then aggregates all registered command factories. Each subcommand file uses package-level `init()` to register its factory, keeping the CLI modular and testable.

## Commands

| Command | Description |
|---------|-------------|
| `sentinel plan [goal] [verification]` | Create a new architectural plan and task |
| `sentinel instruct` | Interview mode to capture user intent and generate tasks |
| `sentinel start [task_id]` | Start the cognitive loop for a specific task |
| `sentinel scan` | Scan project code to update the graph database |
| `sentinel audit` | Run the verification gate for the active task |
| `sentinel status` | Check the current governance status |
| `sentinel visualize` | Generate architecture diagrams from the graph database |
| `sentinel live` | Start the Live View WebSocket server |
| `sentinel report` | Show a colorful compliance dashboard and export to Markdown |
| `sentinel debrief` | Generate session debrief from captured events |
| `sentinel pattern [add|list|search|get|backfill]` | Capture and query architectural and cognitive patterns |

## Workflow

```bash
# 1. Plan a new feature
sentinel plan "Add Auth Service" "go test ./internal/auth/..."

# 2. Scan codebase to update graph
sentinel scan

# 3. Start execution loop
sentinel start <task_id>

# 4. Verify implementation
sentinel audit

# 5. Review status
sentinel status

# 6. Visualize architecture (optional)
sentinel visualize
sentinel live  # real-time WebSocket viewer
```

## Dependencies

- `pkg/sqlite` — database initialization
- `internal/registry` — command factory registration
- `github.com/spf13/cobra` — CLI framework
