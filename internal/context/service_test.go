package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractDocuments_FromGraphifyOutput(t *testing.T) {
	raw := `NODE JWT Middleware [src=internal/auth/middleware.go]
EDGE JWT Middleware --calls--> TokenValidator [src=internal/auth/validator.go]
NODE Authentication Flow [src=docs/architecture/ROADMAP.md]`
	docs := extractDocuments(raw)
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d: %v", len(docs), docs)
	}
}

func TestExtractConcepts_FromGraphifyOutput(t *testing.T) {
	raw := `NODE JWT Middleware [src=internal/auth/middleware.go]
NODE TokenValidator [src=internal/auth/validator.go]
EDGE JWT Middleware --calls--> TokenValidator`
	concepts := extractConcepts(raw)
	if len(concepts) != 2 {
		t.Fatalf("expected 2 concepts, got %d: %v", len(concepts), concepts)
	}
}

func TestFormat_GeneratesMarkdown(t *testing.T) {
	result := &QueryResult{
		Documents: []string{"docs/architecture/ROADMAP.md", "internal/auth/middleware.go"},
		Concepts:  []string{"JWT Middleware", "Authentication Flow"},
	}
	content := Format(result, "authentication", 5)
	required := []string{"## Sentinel Context", "sentinel context \"authentication\"", "ROADMAP.md", "JWT Middleware"}
	for _, r := range required {
		if !strings.Contains(content, r) {
			t.Errorf("missing %q in output:\n%s", r, content)
		}
	}
}

func TestFormat_RespectsLimit(t *testing.T) {
	result := &QueryResult{
		Documents: []string{"a.md", "b.md", "c.md", "d.md", "e.md", "f.md"},
		Concepts:  []string{},
	}
	content := Format(result, "test", 3)
	if strings.Count(content, ".md") > 4 {
		t.Errorf("expected at most 3 documents, got:\n%s", content)
	}
}

func TestInject_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "AGENTS.md")
	content := Format(&QueryResult{
		Documents: []string{"test.md"},
		Concepts:  []string{"Test Concept"},
	}, "test", 5)
	if err := Inject(filePath, content); err != nil {
		t.Fatalf("Inject failed: %v", err)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !strings.Contains(string(data), "## Sentinel Context") {
		t.Error("missing injected section")
	}
}

func TestInject_ReplaceExisting(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "AGENTS.md")
	initial := "Some existing rules.\n\n## Sentinel Context\nold content\n<!-- end Sentinel Context -->\n\nMore rules."
	os.WriteFile(filePath, []byte(initial), 0644)
	newContent := Format(&QueryResult{
		Documents: []string{"new.md"},
		Concepts:  []string{"New Concept"},
	}, "test", 5)
	if err := Inject(filePath, newContent); err != nil {
		t.Fatalf("Inject failed: %v", err)
	}
	data, _ := os.ReadFile(filePath)
	text := string(data)
	if strings.Contains(text, "old content") {
		t.Error("old context not replaced")
	}
	if !strings.Contains(text, "new.md") {
		t.Error("new context not injected")
	}
	if !strings.Contains(text, "Some existing rules") {
		t.Error("existing content lost")
	}
	if !strings.Contains(text, "More rules") {
		t.Error("trailing content lost")
	}
}

func TestExtractDocuments_Malformed(t *testing.T) {
	raw := `NODE Some Concept [src=file1.go
NODE Another Concept [src=file2.go]`
	docs := extractDocuments(raw)
	if len(docs) != 1 {
		t.Fatalf("expected 1 document (the properly formed one), got %d: %v", len(docs), docs)
	}
}

func TestExtractConcepts_Malformed(t *testing.T) {
	raw := `NODE Concept without brackets
NODE ` // trailing NODE without concept
	concepts := extractConcepts(raw)
	// We expect "Concept without brackets" and nothing else
	if len(concepts) != 1 {
		t.Fatalf("expected 1 concept, got %d: %v", len(concepts), concepts)
	}
	if concepts[0] != "Concept without brackets" {
		t.Fatalf("expected 'Concept without brackets', got %v", concepts[0])
	}
}

func TestNewContextService(t *testing.T) {
	svc := NewContextService()
	if svc.graphifyBinary != "graphify" {
		t.Errorf("expected graphifyBinary to be 'graphify', got %s", svc.graphifyBinary)
	}
	if svc.graphFile != "graphify-out/graph.json" {
		t.Errorf("expected graphFile to be 'graphify-out/graph.json', got %s", svc.graphFile)
	}
}

func TestQuery_GraphNotFound(t *testing.T) {
	svc := NewContextService()
	svc.graphFile = "nonexistent.json"
	_, err := svc.Query("test", 100)
	if err == nil {
		t.Error("expected error for missing graph file, got nil")
	}
}

func TestInject_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	// Deliberately create a directory to cause read to fail with IsNotExist = false on some systems,
	// or try to read a non-existent path that is deeply nested to test error handling if IsNotExist is true
	filePath := filepath.Join(tmpDir, "does-not-exist", "AGENTS.md")
	err := Inject(filePath, "test content")
	if err == nil {
		t.Errorf("expected error injecting to deeply nested non-existent path, got nil")
	}
}

func TestQuery_Success(t *testing.T) {
	// Since we mock or skip the exec call for integration, we'll just test that
	// it checks the file exists.
	tmpDir := t.TempDir()
	svc := NewContextService()
	svc.graphFile = filepath.Join(tmpDir, "graph.json")
	os.WriteFile(svc.graphFile, []byte("{}"), 0644)

	// Assuming `graphify` is not in PATH, this will fail at exec
	_, err := svc.Query("test", 100)
	if err == nil {
		t.Errorf("expected error during exec if graphify is missing")
	}
}
