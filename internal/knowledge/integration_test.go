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

    buf.Record(SessionEvent{Type: EventDecision, Domain: "systems", Summary: "chose sqlite cgo-free", Tags: []string{"database"}})
    buf.Record(SessionEvent{Type: EventDecision, Domain: "methodology", Summary: "adopted DI pattern", Tags: []string{"architecture"}})
    buf.Record(SessionEvent{Type: EventError, Domain: "methodology", Summary: "nil pointer in handler", Tags: []string{"bug"}})
    buf.Record(SessionEvent{Type: EventPattern, Domain: "methodology", Summary: "always clamp slice bounds", Tags: []string{"safety"}})
    buf.Record(SessionEvent{Type: EventPattern, Domain: "tools", Summary: "use filepath.Clean before I/O", Tags: []string{"security"}})
    buf.Record(SessionEvent{Type: EventFileChange, File: "api.go", Summary: "added code handler", Tags: []string{"feature"}})

    tmpDir := t.TempDir()
    svc := NewDebriefService(buf, nil, tmpDir)

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

    if !strings.Contains(path, filepath.Join("sessions", "")) {
        t.Errorf("path not in sessions dir: %s", path)
    }
}
