# internal/knowledge

Session event capture and debrief generation for architectural governance.

## Overview

This package records runtime events during a sentinel session and produces structured Markdown debriefs. Events flow through a thread-safe ring buffer to a debrief service that renders templates and persists results to both filesystem and SQLite.

## Key Types

### `SessionEvent`
Captures a single session event with timestamp, type, domain, summary, detail, file path, and tags. Supported event types: `decision`, `error`, `pattern`, `file_change`, `command`, `metric`.

### `EventBuffer`
Thread-safe ring buffer (`sync.RWMutex`) with configurable capacity (default: 1000). Methods:
- `Record(SessionEvent)` — append event, oldest overwritten when full
- `Snapshot()` — all events in chronological order
- `ByDomain(string)` / `ByType(EventType)` — filtered queries
- `Patterns()`, `Decisions()`, `Errors()` — convenience accessors
- `Len()` — current count

### `GlobalBuffer`
Process-wide singleton (`var GlobalBuffer = NewEventBuffer(1000)`) used by all subsystems. Agents record decisions, errors, and patterns here during execution.

### `DebriefService`
Generates and persists session debriefs. Holds a reference to the buffer, optional `*sqlite.DB`, and a base directory.
- `NewDebriefService(buffer, db, baseDir)` — constructor, panics on nil buffer
- `Generate()` — renders Markdown from current buffer state
- `Save(ctx)` / `SaveContent(ctx, content)` — persists to `baseDir/sessions/` and `knowledge_sessions` table

## Dependencies

- `pkg/sqlite` — DB validation and persistence
- `github.com/google/uuid` — session ID generation
- `text/template` — debrief Markdown rendering

## Usage

```go
import "github.com/EmiyaKiritsugu3/sentinel-core/internal/knowledge"

// Record events throughout a session
knowledge.GlobalBuffer.Record(knowledge.SessionEvent{
    Type:    knowledge.EventDecision,
    Domain:  "engine",
    Summary: "Agent terminated: audit complete",
    Tags:    []string{"termination"},
})

// Generate and save debrief
svc := knowledge.NewDebriefService(knowledge.GlobalBuffer, db, "~/knowledge")
sessionID, path, err := svc.Save(context.Background())
```
