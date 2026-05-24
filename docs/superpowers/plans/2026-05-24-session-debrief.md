# Session Debrief Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `sentinel debrief` — a CLI command that captures session decisions, errors, patterns, and file changes into `~/knowledge/sessions/<timestamp>.md` and mirrors them in `graph.db` for future querying.

**Architecture:** In-memory `EventBuffer` (ring buffer, thread-safe) collects events during session. `sentinel debrief` reads buffer, renders markdown via template, saves to filesystem + graph.db. Markdown is source of truth; graph.db enables machine querying. Follows existing DI pattern (`*sqlite.DB` injection, `registry.Register()`).

**Tech Stack:** Go 1.26+, Cobra CLI, SQLite (modernc.org/sqlite), existing `internal/graph` schema migration pattern.

---

## File Structure

| File | Responsibility |
|------|---------------|
| `internal/knowledge/buffer.go` | EventBuffer: ring buffer, Record/Snapshot/ByDomain/ByType |
| `internal/knowledge/buffer_test.go` | EventBuffer unit tests |
| `internal/knowledge/debrief.go` | DebriefService: template rendering, markdown write, graph persistence |
| `internal/knowledge/debrief_test.go` | DebriefService unit tests |
| `cmd/sentinel/commands/debrief.go` | Cobra command: `sentinel debrief` with --auto, --editor, --dry-run |
| `cmd/sentinel/commands/debrief_test.go` | CLI integration tests |
| `internal/graph/schema.go` | +2 tables: `knowledge_sessions`, `session_events` |
| `~/knowledge/meta/template.md` | Debrief markdown template |
| `~/knowledge/meta/index.md` | Knowledge base index |

---

### Task 1: Add `knowledge_sessions` and `session_events` to graph schema

**Files:**
- Modify: `internal/graph/schema.go` (append to `const schema` before closing backtick)
- Modify: `internal/graph/schema.go` (add entries to `pragmaTableInfo` map)

- [ ] **Step 1: Append new table DDL to schema constant**

In `internal/graph/schema.go`, add before the closing backtick of `const schema = ` on line ~167:

```sql
CREATE TABLE IF NOT EXISTS knowledge_sessions (
    id TEXT PRIMARY KEY,
    markdown_path TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    event_count INTEGER DEFAULT 0,
    decision_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    pattern_count INTEGER DEFAULT 0,
    domains TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS session_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    domain TEXT NOT NULL,
    summary TEXT NOT NULL,
    detail TEXT,
    file_path TEXT,
    tags TEXT NOT NULL DEFAULT '',
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES knowledge_sessions(id) ON DELETE CASCADE
);
```

- [ ] **Step 2: Add new tables to `pragmaTableInfo` map**

In `internal/graph/schema.go`, add entries to the `pragmaTableInfo` map (~line 251):

```go
"knowledge_sessions": "PRAGMA table_info(knowledge_sessions)",
"session_events":    "PRAGMA table_info(session_events)",
```

- [ ] **Step 3: Verify existing tests still pass**

Run: `go test ./internal/graph/... -run TestMigrate -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/graph/schema.go
git commit -m "feat(graph): add knowledge_sessions and session_events tables"
```

---

### Task 2: Create EventBuffer (`internal/knowledge/buffer.go`)

**Files:**
- Create: `internal/knowledge/buffer.go`
- Create: `internal/knowledge/buffer_test.go`

- [ ] **Step 1: Create directory and write buffer.go**

```bash
mkdir -p internal/knowledge
```

Write `internal/knowledge/buffer.go`:

```go
package knowledge

import (
    "sort"
    "sync"
    "time"
)

// EventType classifies a session event for debrief categorization.
type EventType string

const (
    EventDecision   EventType = "decision"
    EventError      EventType = "error"
    EventPattern    EventType = "pattern"
    EventFileChange EventType = "file_change"
    EventCommand    EventType = "command"
    EventMetric     EventType = "metric"
)

// SessionEvent represents a single captured event during a sentinel session.
type SessionEvent struct {
    Timestamp time.Time
    Type      EventType
    Domain    string   // hardware, methodology, tools, systems
    Summary   string
    Detail    string
    File      string
    Tags      []string
}

// EventBuffer is a thread-safe ring buffer that collects session events
// for later debrief generation. It lives for the duration of a program run.
type EventBuffer struct {
    mu     sync.RWMutex
    events []SessionEvent
    max    int
    head   int
    size   int
}

// NewEventBuffer creates a ring buffer with the given maximum capacity.
func NewEventBuffer(maxSize int) *EventBuffer {
    if maxSize < 1 {
        maxSize = 1000
    }
    return &EventBuffer{
        events: make([]SessionEvent, maxSize),
        max:    maxSize,
    }
}

// Record appends an event to the buffer. Thread-safe. Oldest events are
// overwritten when the buffer is full.
func (b *EventBuffer) Record(event SessionEvent) {
    b.mu.Lock()
    defer b.mu.Unlock()
    if event.Timestamp.IsZero() {
        event.Timestamp = time.Now()
    }
    b.events[b.head] = event
    b.head = (b.head + 1) % b.max
    if b.size < b.max {
        b.size++
    }
}

// Snapshot returns all events in chronological order (oldest first).
func (b *EventBuffer) Snapshot() []SessionEvent {
    b.mu.RLock()
    defer b.mu.RUnlock()
    result := make([]SessionEvent, b.size)
    if b.size == 0 {
        return result
    }
    start := (b.head - b.size + b.max) % b.max
    for i := 0; i < b.size; i++ {
        idx := (start + i) % b.max
        result[i] = b.events[idx]
    }
    sort.Slice(result, func(i, j int) bool {
        return result[i].Timestamp.Before(result[j].Timestamp)
    })
    return result
}

// ByDomain returns all events matching the given domain.
func (b *EventBuffer) ByDomain(domain string) []SessionEvent {
    return b.filter(func(e SessionEvent) bool {
        return e.Domain == domain
    })
}

// ByType returns all events matching the given event type.
func (b *EventBuffer) ByType(typ EventType) []SessionEvent {
    return b.filter(func(e SessionEvent) bool {
        return e.Type == typ
    })
}

// Patterns returns all events of type EventPattern.
func (b *EventBuffer) Patterns() []SessionEvent {
    return b.ByType(EventPattern)
}

// Decisions returns all events of type EventDecision.
func (b *EventBuffer) Decisions() []SessionEvent {
    return b.ByType(EventDecision)
}

// Errors returns all events of type EventError.
func (b *EventBuffer) Errors() []SessionEvent {
    return b.ByType(EventError)
}

// Len returns the current number of events in the buffer.
func (b *EventBuffer) Len() int {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.size
}

// GlobalBuffer is the process-wide singleton event buffer. All sentinel
// subsystems (engine, commands, etc.) record events here during a session.
var GlobalBuffer = NewEventBuffer(1000)

func (b *EventBuffer) filter(pred func(SessionEvent) bool) []SessionEvent {
    b.mu.RLock()
    defer b.mu.RUnlock()
    var result []SessionEvent
    if b.size == 0 {
        return result
    }
    start := (b.head - b.size + b.max) % b.max
    for i := 0; i < b.size; i++ {
        idx := (start + i) % b.max
        if pred(b.events[idx]) {
            result = append(result, b.events[idx])
        }
    }
    return result
}
```

- [ ] **Step 2: Write tests in `internal/knowledge/buffer_test.go`**

```go
package knowledge

import (
    "sync"
    "testing"
    "time"
)

func TestNewEventBuffer_DefaultsTo1000(t *testing.T) {
    buf := NewEventBuffer(0)
    if buf.max != 1000 {
        t.Fatalf("expected max=1000 for zero arg, got %d", buf.max)
    }
    if buf.Len() != 0 {
        t.Fatal("expected empty buffer")
    }
}

func TestRecordAndSnapshot_Basic(t *testing.T) {
    buf := NewEventBuffer(10)
    buf.Record(SessionEvent{
        Type:    EventDecision,
        Domain:  "systems",
        Summary: "chose sqlite",
    })
    buf.Record(SessionEvent{
        Type:    EventError,
        Domain:  "methodology",
        Summary: "nil pointer",
    })

    if buf.Len() != 2 {
        t.Fatalf("expected len=2, got %d", buf.Len())
    }

    snap := buf.Snapshot()
    if len(snap) != 2 {
        t.Fatalf("expected snapshot len=2, got %d", len(snap))
    }
    if snap[0].Summary != "chose sqlite" {
        t.Errorf("first event: %q", snap[0].Summary)
    }
    if snap[1].Summary != "nil pointer" {
        t.Errorf("second event: %q", snap[1].Summary)
    }
}

func TestSnapshot_ChronologicalOrder(t *testing.T) {
    buf := NewEventBuffer(5)
    t1 := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC)
    t2 := t1.Add(time.Hour)
    t3 := t1.Add(2 * time.Hour)

    buf.Record(SessionEvent{Timestamp: t2, Type: EventDecision, Summary: "second"})
    buf.Record(SessionEvent{Timestamp: t1, Type: EventDecision, Summary: "first"})
    buf.Record(SessionEvent{Timestamp: t3, Type: EventDecision, Summary: "third"})

    snap := buf.Snapshot()
    if snap[0].Summary != "first" || snap[1].Summary != "second" || snap[2].Summary != "third" {
        t.Errorf("wrong order: %v", []string{snap[0].Summary, snap[1].Summary, snap[2].Summary})
    }
}

func TestRingBuffer_Wraparound(t *testing.T) {
    buf := NewEventBuffer(3)
    for i := 0; i < 5; i++ {
        buf.Record(SessionEvent{Type: EventMetric, Summary: "event"})
    }
    if buf.Len() != 3 {
        t.Fatalf("expected len=3 after wraparound, got %d", buf.Len())
    }
}

func TestByDomain(t *testing.T) {
    buf := NewEventBuffer(10)
    buf.Record(SessionEvent{Type: EventDecision, Domain: "systems", Summary: "a"})
    buf.Record(SessionEvent{Type: EventError, Domain: "methodology", Summary: "b"})
    buf.Record(SessionEvent{Type: EventPattern, Domain: "systems", Summary: "c"})

    sys := buf.ByDomain("systems")
    if len(sys) != 2 {
        t.Fatalf("expected 2 systems events, got %d", len(sys))
    }
}

func TestByType(t *testing.T) {
    buf := NewEventBuffer(10)
    buf.Record(SessionEvent{Type: EventDecision, Summary: "a"})
    buf.Record(SessionEvent{Type: EventError, Summary: "b"})
    buf.Record(SessionEvent{Type: EventError, Summary: "c"})

    errs := buf.ByType(EventError)
    if len(errs) != 2 {
        t.Fatalf("expected 2 errors, got %d", len(errs))
    }
}

func TestPatternsDecisionsErrors_Shortcuts(t *testing.T) {
    buf := NewEventBuffer(10)
    buf.Record(SessionEvent{Type: EventDecision, Summary: "d"})
    buf.Record(SessionEvent{Type: EventError, Summary: "e"})
    buf.Record(SessionEvent{Type: EventPattern, Summary: "p"})

    if len(buf.Patterns()) != 1 {
        t.Error("Patterns() count wrong")
    }
    if len(buf.Decisions()) != 1 {
        t.Error("Decisions() count wrong")
    }
    if len(buf.Errors()) != 1 {
        t.Error("Errors() count wrong")
    }
}

func TestEmptyBuffer(t *testing.T) {
    buf := NewEventBuffer(10)
    if buf.Len() != 0 {
        t.Fatal("expected len=0")
    }
    snap := buf.Snapshot()
    if len(snap) != 0 {
        t.Fatal("expected empty snapshot")
    }
    if len(buf.ByDomain("systems")) != 0 {
        t.Fatal("expected empty ByDomain")
    }
    if len(buf.Patterns()) != 0 {
        t.Fatal("expected empty Patterns")
    }
}

func TestConcurrentRecordAndRead(t *testing.T) {
    buf := NewEventBuffer(100)
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            for j := 0; j < 20; j++ {
                buf.Record(SessionEvent{Type: EventMetric, Summary: "event"})
            }
        }(i)
    }
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 50; j++ {
                _ = buf.Snapshot()
                _ = buf.Len()
                _ = buf.Patterns()
            }
        }(i)
    }
    wg.Wait()
    // No race detector failures = pass
}
```

- [ ] **Step 3: Run tests with race detector**

Run: `go test ./internal/knowledge/... -race -v`
Expected: all tests PASS, no race conditions

- [ ] **Step 4: Commit**

```bash
git add internal/knowledge/buffer.go internal/knowledge/buffer_test.go
git commit -m "feat(knowledge): add EventBuffer ring buffer for session event capture"
```

---

### Task 3: Create DebriefService (`internal/knowledge/debrief.go`)

**Files:**
- Create: `internal/knowledge/debrief.go`
- Create: `internal/knowledge/debrief_test.go`

- [ ] **Step 1: Write debrief.go**

Write `internal/knowledge/debrief.go`:

```go
package knowledge

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
    "github.com/google/uuid"
)

const debriefTemplate = `# Session Debrief — {{.Date}} {{.Time}}

## Decisions Made
{{range .Decisions}}- {{.Summary}}
{{end}}
## Patterns Observed
### Anti-Patterns (what failed)
{{range .Errors}}- {{.Summary}}
{{end}}
### Success Patterns (what worked)
{{range .Patterns}}- {{.Summary}}
{{end}}
## Files Changed
{{range .FileChanges}}- {{.File}} — {{.Summary}}
{{end}}
## Domain Tags
{{range .Domains}}- {{.}}
{{end}}
## Follow-ups
- [ ] ...
`

// DebriefData holds the structured template variables for debrief rendering.
type DebriefData struct {
    Date        string
    Time        string
    Decisions   []SessionEvent
    Errors      []SessionEvent
    Patterns    []SessionEvent
    FileChanges []SessionEvent
    Domains     []string
}

// DebriefService generates and persists session debriefs.
type DebriefService struct {
    buffer   *EventBuffer
    db       *sqlite.DB
    baseDir  string // ~/knowledge
    tmpl     string
}

// NewDebriefService creates a debrief service. baseDir is the knowledge root
// directory (typically ~/knowledge).
func NewDebriefService(buffer *EventBuffer, db *sqlite.DB, baseDir string) *DebriefService {
    return &DebriefService{
        buffer:  buffer,
        db:      db,
        baseDir: baseDir,
        tmpl:    debriefTemplate,
    }
}

// Generate renders the debrief markdown from the current buffer contents.
func (s *DebriefService) Generate() string {
    now := time.Now()
    data := DebriefData{
        Date:        now.Format("2006-01-02"),
        Time:        now.Format("15:04"),
        Decisions:   s.buffer.Decisions(),
        Errors:      s.buffer.Errors(),
        Patterns:    s.buffer.Patterns(),
        FileChanges: s.buffer.ByType(EventFileChange),
    }

    domainSet := make(map[string]bool)
    for _, e := range s.buffer.Snapshot() {
        if e.Domain != "" {
            domainSet[e.Domain] = true
        }
    }
    for d := range domainSet {
        data.Domains = append(data.Domains, d)
    }

    return s.renderTemplate(data)
}

func (s *DebriefService) renderTemplate(data DebriefData) string {
    result := s.tmpl
    result = strings.Replace(result, "{{.Date}}", data.Date, 1)
    result = strings.Replace(result, "{{.Time}}", data.Time, 1)

    var decisions strings.Builder
    for _, d := range data.Decisions {
        decisions.WriteString(fmt.Sprintf("- %s\n", d.Summary))
    }
    result = strings.Replace(result, "{{range .Decisions}}- {{.Summary}}\n{{end}}", decisions.String(), 1)

    var errors strings.Builder
    for _, e := range data.Errors {
        errors.WriteString(fmt.Sprintf("- %s\n", e.Summary))
    }
    result = strings.Replace(result, "{{range .Errors}}- {{.Summary}}\n{{end}}", errors.String(), 1)

    var patterns strings.Builder
    for _, p := range data.Patterns {
        patterns.WriteString(fmt.Sprintf("- %s\n", p.Summary))
    }
    result = strings.Replace(result, "{{range .Patterns}}- {{.Summary}}\n{{end}}", patterns.String(), 1)

    var files strings.Builder
    for _, f := range data.FileChanges {
        files.WriteString(fmt.Sprintf("- %s — %s\n", f.File, f.Summary))
    }
    result = strings.Replace(result, "{{range .FileChanges}}- {{.File}} — {{.Summary}}\n{{end}}", files.String(), 1)

    var domains strings.Builder
    for _, d := range data.Domains {
        domains.WriteString(fmt.Sprintf("- %s\n", d))
    }
    result = strings.Replace(result, "{{range .Domains}}- {{.}}\n{{end}}", domains.String(), 1)

    return result
}

// Save persists the debrief to the filesystem and graph database.
// Returns the session ID, markdown path, and any error.
func (s *DebriefService) Save(ctx context.Context) (string, string, error) {
    now := time.Now()
    sessionID := uuid.New().String()[:8]
    filename := fmt.Sprintf("%s-%s.md", now.Format("2006-01-02"), now.Format("1504"))
    dir := filepath.Join(s.baseDir, "sessions")

    if err := os.MkdirAll(dir, 0755); err != nil {
        return "", "", fmt.Errorf("debrief: create sessions dir %s: %w", dir, err)
    }

    content := s.Generate()
    path := filepath.Join(dir, filename)
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        return "", "", fmt.Errorf("debrief: write markdown: %w", err)
    }

    if s.db != nil {
        if err := s.saveToGraph(ctx, sessionID, path, now); err != nil {
            // Graceful degradation: markdown saved, graph failed
            fmt.Fprintf(os.Stderr, "warning: debrief graph persistence failed: %v\n", err)
        }
    }

    return sessionID, path, nil
}

func (s *DebriefService) saveToGraph(ctx context.Context, sessionID, path string, now time.Time) error {
    if err := sqlite.ValidateDB(s.db, "debrief-graph"); err != nil {
        return err
    }

    tx, err := s.db.Conn.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("debrief: begin tx: %w", err)
    }
    defer func() { _ = tx.Rollback() }()

    decisions := s.buffer.Decisions()
    errors := s.buffer.Errors()
    patterns := s.buffer.Patterns()
    allEvents := s.buffer.Snapshot()

    domainSet := make(map[string]bool)
    for _, e := range allEvents {
        if e.Domain != "" {
            domainSet[e.Domain] = true
        }
    }
    domains := make([]string, 0, len(domainSet))
    for d := range domainSet {
        domains = append(domains, d)
    }

    _, err = tx.ExecContext(ctx,
        `INSERT INTO knowledge_sessions (id, markdown_path, started_at, ended_at, event_count, decision_count, error_count, pattern_count, domains)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        sessionID, path, now.Add(-time.Hour), now, len(allEvents), len(decisions), len(errors), len(patterns),
        strings.Join(domains, ","),
    )
    if err != nil {
        return fmt.Errorf("debrief: insert session: %w", err)
    }

    for _, e := range allEvents {
        tags := strings.Join(e.Tags, ",")
        detail := e.Detail
        if detail == "" {
            detail = ""
        }
        file := e.File
        if file == "" {
            file = ""
        }
        _, err = tx.ExecContext(ctx,
            `INSERT INTO session_events (session_id, event_type, domain, summary, detail, file_path, tags)
             VALUES (?, ?, ?, ?, ?, ?, ?)`,
            sessionID, string(e.Type), e.Domain, e.Summary, e.Detail, e.File, tags,
        )
        if err != nil {
            return fmt.Errorf("debrief: insert event: %w", err)
        }
    }

    if err = tx.Commit(); err != nil {
        return fmt.Errorf("debrief: commit tx: %w", err)
    }
    return nil
}
```

- [ ] **Step 2: Write tests in `internal/knowledge/debrief_test.go`**

```go
package knowledge

import (
    "context"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestDebriefService_Generate_EmptyBuffer(t *testing.T) {
    buf := NewEventBuffer(10)
    tmpDir := t.TempDir()
    svc := NewDebriefService(buf, nil, tmpDir)

    result := svc.Generate()
    if !strings.Contains(result, "## Decisions Made") {
        t.Error("missing Decisions Made section")
    }
    if !strings.Contains(result, "## Patterns Observed") {
        t.Error("missing Patterns Observed section")
    }
    if !strings.Contains(result, "## Files Changed") {
        t.Error("missing Files Changed section")
    }
}

func TestDebriefService_Generate_WithEvents(t *testing.T) {
    buf := NewEventBuffer(10)
    buf.Record(SessionEvent{Type: EventDecision, Domain: "systems", Summary: "chose sqlite"})
    buf.Record(SessionEvent{Type: EventError, Domain: "methodology", Summary: "nil pointer"})
    buf.Record(SessionEvent{Type: EventPattern, Domain: "methodology", Summary: "always clean paths"})
    buf.Record(SessionEvent{Type: EventFileChange, File: "api.go", Summary: "added handler"})

    tmpDir := t.TempDir()
    svc := NewDebriefService(buf, nil, tmpDir)

    result := svc.Generate()
    if !strings.Contains(result, "chose sqlite") {
        t.Error("missing decision")
    }
    if !strings.Contains(result, "nil pointer") {
        t.Error("missing error")
    }
    if !strings.Contains(result, "always clean paths") {
        t.Error("missing pattern")
    }
    if !strings.Contains(result, "api.go") {
        t.Error("missing file change")
    }
    if !strings.Contains(result, "methodology") {
        t.Error("missing domain tag")
    }
}

func TestDebriefService_Save_WritesMarkdown(t *testing.T) {
    buf := NewEventBuffer(10)
    buf.Record(SessionEvent{Type: EventDecision, Domain: "systems", Summary: "test decision"})

    tmpDir := t.TempDir()
    svc := NewDebriefService(buf, nil, tmpDir)

    _, path, err := svc.Save(context.Background())
    if err != nil {
        t.Fatalf("Save failed: %v", err)
    }

    content, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("read markdown: %v", err)
    }
    if !strings.Contains(string(content), "test decision") {
        t.Error("markdown missing decision content")
    }
}

func TestDebriefService_Save_CreatesDirectory(t *testing.T) {
    buf := NewEventBuffer(10)
    tmpDir := t.TempDir()
    nestedDir := filepath.Join(tmpDir, "does", "not", "exist")

    svc := NewDebriefService(buf, nil, nestedDir)
    _, _, err := svc.Save(context.Background())
    if err != nil {
        t.Fatalf("Save should create parent dirs: %v", err)
    }
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/knowledge/... -v`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/knowledge/debrief.go internal/knowledge/debrief_test.go
git commit -m "feat(knowledge): add DebriefService for markdown generation and persistence"
```

---

### Task 4: Create `sentinel debrief` CLI command

**Files:**
- Create: `cmd/sentinel/commands/debrief.go`
- Create: `cmd/sentinel/commands/debrief_test.go`

- [ ] **Step 1: Write debrief.go command**

```go
package commands

import (
    "context"
    "fmt"
    "os"
    "os/exec"

    "github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/knowledge"
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
    "github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
    "github.com/spf13/cobra"
)

func init() {
    registry.Register(NewDebriefCmd)
}

// NewDebriefCmd creates the sentinel debrief command, which generates a
// session debrief from captured events and persists it to ~/knowledge/.
func NewDebriefCmd(db *sqlite.DB) *cobra.Command {
    var auto, dryRun bool
    var editor bool
    var outputPath string

    cmd := &cobra.Command{
        Use:   "debrief",
        Short: "Generate session debrief from captured events",
        Long: `Debrief captures decisions, errors, patterns, and file changes
from the current sentinel session and saves them to ~/knowledge/sessions/.

Events are collected automatically via the EventBuffer. Run this command
at the end of your session to persist captured knowledge.`,
    }

    if err := sqlite.ValidateDB(db, "debrief-cmd"); err != nil {
        cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
        return cmd
    }

    cmd.RunE = func(cmd *cobra.Command, args []string) error {
            if err := graph.Migrate(cmd.Context(), db); err != nil {
                return fmt.Errorf("debrief: migration failed: %w", err)
            }

            homeDir, err := os.UserHomeDir()
            if err != nil {
                return fmt.Errorf("debrief: cannot find home directory: %w", err)
            }
            baseDir := fmt.Sprintf("%s/knowledge", homeDir)

            svc := knowledge.NewDebriefService(knowledge.GlobalBuffer, db, baseDir)
            content := svc.Generate()

            if dryRun {
                fmt.Println(content)
                fmt.Printf("\n[DRY RUN] Would save to %s/knowledge/sessions/\n", homeDir)
                return nil
            }

            if auto {
                id, path, err := svc.Save(cmd.Context())
                if err != nil {
                    return err
                }
                fmt.Printf("Saved: %s (session %s, %d events)\n", path, id, knowledge.GlobalBuffer.Len())
                return nil
            }

            if editor {
                return openInEditor(content, svc, cmd.Context())
            }

            return interactiveDebrief(content, svc, cmd.Context())
        },
    }

    cmd.Flags().BoolVar(&auto, "auto", false, "Skip prompts, save all captured events")
    cmd.Flags().BoolVar(&editor, "editor", false, "Open in $EDITOR instead of interactive prompts")
    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be saved, don't persist")
    cmd.Flags().StringVar(&outputPath, "output", "", "Override output path")

    return cmd
}

func interactiveDebrief(content string, svc *knowledge.DebriefService, ctx context.Context) error {
    fmt.Println(content)
    fmt.Print("\nSave this debrief? [Y/n]: ")
    var answer string
    fmt.Scanln(&answer)
    if answer == "" || answer == "y" || answer == "Y" {
        id, path, err := svc.Save(ctx)
        if err != nil {
            return err
        }
        fmt.Printf("Saved: %s (session %s)\n", path, id)
    } else {
        fmt.Println("Debrief discarded.")
    }
    return nil
}

func openInEditor(content string, svc *knowledge.DebriefService, ctx context.Context) error {
    tmpFile, err := os.CreateTemp("", "sentinel-debrief-*.md")
    if err != nil {
        return fmt.Errorf("debrief: create temp file: %w", err)
    }
    defer os.Remove(tmpFile.Name())

    if _, err := tmpFile.WriteString(content); err != nil {
        return fmt.Errorf("debrief: write temp file: %w", err)
    }
    tmpFile.Close()

    editorCmd := os.Getenv("EDITOR")
    if editorCmd == "" {
        editorCmd = "vi"
    }
    c := exec.Command(editorCmd, tmpFile.Name())
    c.Stdin = os.Stdin
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    if err := c.Run(); err != nil {
        return fmt.Errorf("debrief: editor failed: %w", err)
    }

    edited, err := os.ReadFile(tmpFile.Name())
    if err != nil {
        return fmt.Errorf("debrief: read edited file: %w", err)
    }

    fmt.Println(string(edited))
    fmt.Print("\nSave this debrief? [Y/n]: ")
    var answer string
    fmt.Scanln(&answer)
    if answer == "" || answer == "y" || answer == "Y" {
        id, path, err := svc.Save(ctx)
        if err != nil {
            return err
        }
        fmt.Printf("Saved: %s (session %s)\n", path, id)
    } else {
        fmt.Println("Debrief discarded.")
    }
    return nil
}
```

- [ ] **Step 2: Write test in `cmd/sentinel/commands/debrief_test.go`**

```go
package commands

import (
    "errors"
    "testing"

    "github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestNewDebriefCmd_Registers(t *testing.T) {
    cmd := NewDebriefCmd(nil)
    if cmd.Use != "debrief" {
        t.Fatalf("expected Use='debrief', got %q", cmd.Use)
    }
}

func TestNewDebriefCmd_NilDB(t *testing.T) {
    cmd := NewDebriefCmd(nil)
    err := cmd.Execute()
    if !errors.Is(err, sqlite.ErrNilDB) {
        t.Fatalf("expected ErrNilDB, got %v", err)
    }
}

func TestDebriefFlags(t *testing.T) {
    cmd := NewDebriefCmd(nil)
    flags := []string{"auto", "editor", "dry-run", "output"}
    for _, f := range flags {
        if cmd.Flags().Lookup(f) == nil {
            t.Errorf("missing flag: %s", f)
        }
    }
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./cmd/sentinel/`
Expected: exit 0

- [ ] **Step 4: Run tests**

Run: `go test ./cmd/sentinel/commands/ -run Debrief -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/sentinel/commands/debrief.go cmd/sentinel/commands/debrief_test.go
git commit -m "feat(cli): add 'sentinel debrief' command for session knowledge capture"
```

---

### Task 5: Create `~/knowledge/` scaffolding

**Files:**
- Create: `~/knowledge/meta/template.md`
- Create: `~/knowledge/meta/index.md`

No code changes — this task creates the directory structure and markdown templates.

- [ ] **Step 1: Create directories and template.md**

```bash
mkdir -p ~/knowledge/{sessions,domains/{methodology,tools,systems,hardware},meta,patterns}
```

- [ ] **Step 2: Write `~/knowledge/meta/template.md`**

```markdown
# Session Debrief — {{date}} {{time}}

## Decisions Made
<!-- List architectural and technical decisions made this session -->

## Patterns Observed
### Anti-Patterns (what failed)
<!-- Bugs, mistakes, unexpected behaviors -->

### Success Patterns (what worked)
<!-- Techniques, approaches, workflows that succeeded -->

## Errors Encountered
<!-- Errors and how they were resolved -->

## Files Changed
<!-- Key files modified with brief description of changes -->

## Domain Tags
<!-- hardware | methodology | tools | systems -->

## Follow-ups
<!-- Items to address in future sessions -->
```

- [ ] **Step 3: Write `~/knowledge/meta/index.md`**

```markdown
# Knowledge Base Index

## Sessions
<!-- List of debrief session files -->
<!-- Format: YYYY-MM-DD-HHMM.md — one sentence summary -->

## Domains
### Methodology
<!-- Patterns, workflows, process improvements -->
### Tools
<!-- CLI patterns, editor configs, dependency notes -->
### Systems
<!-- Architecture decisions, API design, infra notes -->
### Hardware
<!-- Performance profiles, environment configs -->

## Cross-Cutting Patterns
<!-- Anti-patterns and success patterns observed across sessions -->

---
Last updated: 2026-05-24
```

- [ ] **Step 4: Commit**

```bash
# ~/knowledge/ is outside the git repo — no commit needed.
# This task is manual scaffolding only.
```

---

### Task 6: Integration test — end-to-end debrief flow

**Files:**
- Create: `internal/knowledge/integration_test.go`

- [ ] **Step 1: Write integration test**

```go
package knowledge

import (
    "context"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestIntegration_DebriefFullFlow(t *testing.T) {
    buf := NewEventBuffer(50)

    // Simulate a full session
    buf.Record(SessionEvent{Type: EventDecision, Domain: "systems", Summary: "chose sqlite cgo-free", Tags: []string{"database"}})
    buf.Record(SessionEvent{Type: EventDecision, Domain: "methodology", Summary: "adopted DI pattern", Tags: []string{"architecture"}})
    buf.Record(SessionEvent{Type: EventError, Domain: "methodology", Summary: "nil pointer in handler", Tags: []string{"bug"}})
    buf.Record(SessionEvent{Type: EventPattern, Domain: "methodology", Summary: "always clamp slice bounds", Tags: []string{"safety"}})
    buf.Record(SessionEvent{Type: EventPattern, Domain: "tools", Summary: "use filepath.Clean before I/O", Tags: []string{"security"}})
    buf.Record(SessionEvent{Type: EventFileChange, File: "api.go", Summary: "added code handler", Tags: []string{"feature"}})

    tmpDir := t.TempDir()
    svc := NewDebriefService(buf, nil, tmpDir)

    // Generate
    content := svc.Generate()
    required := []string{
        "chose sqlite cgo-free",
        "adopted DI pattern",
        "nil pointer in handler",
        "always clamp slice bounds",
        "use filepath.Clean before I/O",
        "api.go",
        "## Decisions Made",
        "## Patterns Observed",
        "### Anti-Patterns",
        "### Success Patterns",
        "## Files Changed",
        "## Domain Tags",
        "methodology",
        "tools",
        "systems",
    }
    for _, r := range required {
        if !strings.Contains(content, r) {
            t.Errorf("missing in generated output: %q", r)
        }
    }

    // Save
    _, path, err := svc.Save(context.Background())
    if err != nil {
        t.Fatalf("Save failed: %v", err)
    }

    saved, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("read saved file: %v", err)
    }

    if !strings.Contains(string(saved), "chose sqlite cgo-free") {
        t.Error("saved file missing content")
    }

    // Verify file is in sessions subdirectory
    if !strings.Contains(path, filepath.Join("sessions", "")) {
        t.Errorf("path not in sessions dir: %s", path)
    }
}
```

- [ ] **Step 2: Run integration test**

Run: `go test ./internal/knowledge/... -run Integration -v`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `go test ./internal/knowledge/... -race -v`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/knowledge/integration_test.go
git commit -m "test(knowledge): add end-to-end debrief integration test"
```

---

### Task 7: Final verification — build, lint, full test suite

- [ ] **Step 1: Build**

Run: `go build ./cmd/sentinel/`
Expected: exit 0

- [ ] **Step 2: Lint**

Run: `golangci-lint run ./internal/knowledge/... ./cmd/sentinel/commands/debrief.go`
Expected: 0 issues

- [ ] **Step 3: Full test suite**

Run: `go test ./internal/knowledge/... ./internal/graph/... -race -v`
Expected: all PASS

- [ ] **Step 4: Verify `sentinel debrief --help` works**

Run: `./sentinel debrief --help`
Expected: shows flags: --auto, --editor, --dry-run, --output

- [ ] **Step 5: Verify `sentinel debrief --dry-run` with empty buffer**

Run: `./sentinel debrief --dry-run`
Expected: prints template with empty sections, "Would save to ..."

- [ ] **Step 6: Commit and push**

```bash
git add -A
git commit -m "chore: final verification — build, lint, tests all pass"
git push origin HEAD
```

---

### Self-Review Checklist

**1. Spec coverage:**
- [x] `~/knowledge/` directory structure → Task 5
- [x] EventBuffer → Task 2
- [x] sentinel debrief CLI → Task 4
- [x] Graph integration (knowledge_sessions + session_events) → Task 1
- [x] Markdown template → Task 3 (embedded in DebriefService) + Task 5
- [x] Error handling (empty buffer, missing dir, graph fail) → Task 3 (Save, saveToGraph)
- [x] Interactive + --auto + --editor + --dry-run modes → Task 4
- [x] Graceful degradation (markdown saves even if graph fails) → Task 3 (Save method)
- [x] Testing (unit, concurrent, integration) → Tasks 2, 3, 6

**2. Placeholder scan:** No TBD, TODO, or vague directives. Every code block is complete.

**3. Type consistency:**
- `EventBuffer` API: Record, Snapshot, ByDomain, ByType, Patterns, Decisions, Errors, Len — consistent across all tasks
- `DebriefService` API: Generate, Save — consistent across Tasks 3, 4, 6
- Schema tables: `knowledge_sessions`, `session_events` — column names match across migration and insert SQL
- Command flags: `--auto`, `--editor`, `--dry-run`, `--output` — consistent in Task 4 and Task 7 verification
