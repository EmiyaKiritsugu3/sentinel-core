package knowledge

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestDebriefService_Generate_EmptyBuffer(t *testing.T) {
	buf := NewEventBuffer(100)
	tmpDir := t.TempDir()
	svc := NewDebriefService(buf, nil, tmpDir)

	result := svc.Generate()

	requiredSections := []string{
		"# Session Debrief",
		"## Decisions Made",
		"## Patterns Observed",
		"### Anti-Patterns",
		"### Success Patterns",
		"## Files Changed",
		"## Domain Tags",
		"## Follow-ups",
	}
	for _, section := range requiredSections {
		if !strings.Contains(result, section) {
			t.Errorf("expected section %q in generated output", section)
		}
	}
}

func TestDebriefService_Generate_WithEvents(t *testing.T) {
	buf := NewEventBuffer(100)
	buf.Record(SessionEvent{Type: EventDecision, Summary: "use gRPC for service mesh"})
	buf.Record(SessionEvent{Type: EventError, Summary: "nil pointer dereference in handler"})
	buf.Record(SessionEvent{Type: EventPattern, Summary: "table-driven tests consistently pass"})
	buf.Record(SessionEvent{
		Type:    EventFileChange,
		File:    "internal/knowledge/buffer.go",
		Summary: "added ring buffer implementation",
	})
	buf.Record(SessionEvent{
		Type:    EventFileChange,
		File:    "pkg/sqlite/validation.go",
		Summary: "added nil guard",
	})

	tmpDir := t.TempDir()
	svc := NewDebriefService(buf, nil, tmpDir)

	result := svc.Generate()

	checks := []string{
		"use gRPC for service mesh",
		"nil pointer dereference in handler",
		"table-driven tests consistently pass",
		"internal/knowledge/buffer.go",
		"pkg/sqlite/validation.go",
		"added ring buffer implementation",
		"added nil guard",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected %q in generated output", check)
		}
	}
}

func TestDebriefService_Save_WritesMarkdown(t *testing.T) {
	buf := NewEventBuffer(100)
	buf.Record(SessionEvent{Type: EventDecision, Summary: "use embedded SQLite"})
	buf.Record(SessionEvent{Type: EventError, Summary: "race condition in cache"})
	buf.Record(SessionEvent{Type: EventPattern, Summary: "batch writes improve throughput"})

	tmpDir := t.TempDir()
	svc := NewDebriefService(buf, nil, tmpDir)

	sessionID, path, err := svc.Save(t.Context())
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if sessionID == "" {
		t.Error("expected non-empty session ID")
	}
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("expected .md extension, got %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read markdown file: %v", err)
	}
	text := string(content)

	expectedChecks := []string{
		"use embedded SQLite",
		"race condition in cache",
		"batch writes improve throughput",
	}
	for _, check := range expectedChecks {
		if !strings.Contains(text, check) {
			t.Errorf("expected %q in saved markdown", check)
		}
	}
}

func TestDebriefService_Save_CreatesDirectory(t *testing.T) {
	buf := NewEventBuffer(100)
	buf.Record(SessionEvent{Type: EventDecision, Summary: "directory creation test"})

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "deeply", "nested", "knowledge")
	svc := NewDebriefService(buf, nil, nestedDir)

	_, path, err := svc.Save(t.Context())
	if err != nil {
		t.Fatalf("Save failed with nested dir: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("markdown file not found at %s: %v", path, err)
	}

	sessionsDir := filepath.Join(nestedDir, "sessions")
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		t.Errorf("sessions directory not created at %s", sessionsDir)
	}
}

func TestDebriefService_Generate_DomainsDeterministic(t *testing.T) {
	buf := NewEventBuffer(100)
	buf.Record(SessionEvent{Type: EventDecision, Domain: "systems", Summary: "d1"})
	buf.Record(SessionEvent{Type: EventError, Domain: "methodology", Summary: "e1"})
	buf.Record(SessionEvent{Type: EventPattern, Domain: "systems", Summary: "p1"})

	tmpDir := t.TempDir()
	svc := NewDebriefService(buf, nil, tmpDir)

	run1 := svc.Generate()
	run2 := svc.Generate()

	if run1 != run2 {
		t.Error("Generate() is not deterministic across calls with same buffer")
	}

	methodologyIdx := strings.Index(run1, "methodology")
	systemsIdx := strings.Index(run1, "systems")
	if methodologyIdx == -1 || systemsIdx == -1 {
		t.Fatal("domain tags not found in output")
	}
	if methodologyIdx >= systemsIdx {
		t.Error("domains not in alphabetical order (expected methodology before systems)")
	}
}

func TestDebriefService_SaveContent_UniqueFilenames(t *testing.T) {
	buf := NewEventBuffer(10)
	tmpDir := t.TempDir()
	svc := NewDebriefService(buf, nil, tmpDir)
	_, path1, _ := svc.SaveContent(context.Background(), "c1")
	_, path2, _ := svc.SaveContent(context.Background(), "c2")
	if path1 == path2 {
		t.Errorf("filenames must be unique: both got %s", path1)
	}
}

func TestDebriefService_SaveContent_UsesProvidedContent(t *testing.T) {
	buf := NewEventBuffer(10)
	tmpDir := t.TempDir()
	svc := NewDebriefService(buf, nil, tmpDir)

	customContent := "# Custom Debrief\n\nCustom content here.\n"
	_, path, err := svc.SaveContent(context.Background(), customContent)
	if err != nil {
		t.Fatalf("SaveContent failed: %v", err)
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(saved) != customContent {
		t.Errorf("SaveContent did not save provided content:\nwant: %q\ngot:  %q", customContent, string(saved))
	}
}

func TestNewDebriefService_NilBufferPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil buffer")
		}
	}()
	_ = NewDebriefService(nil, nil, t.TempDir())
}

func TestDebriefService_SaveToGraph_WithRealDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sqlite.InitAtPath(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	buf := NewEventBuffer(10)
	buf.Record(SessionEvent{Type: EventDecision, Domain: "systems", Summary: "test decision", Tags: []string{"db"}})
	buf.Record(SessionEvent{Type: EventError, Domain: "methodology", Summary: "test error", Tags: []string{"db"}})

	svc := NewDebriefService(buf, db, tmpDir)
	id, path, err := svc.Save(context.Background())
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty session ID")
	}
	if path == "" {
		t.Error("expected non-empty path")
	}

	var count int
	row := db.Conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM knowledge_sessions WHERE id = ?", id)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("query sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 session record, got %d", count)
	}

	row = db.Conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM session_events WHERE session_id = ?", id)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("query events: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 event records, got %d", count)
	}
}
