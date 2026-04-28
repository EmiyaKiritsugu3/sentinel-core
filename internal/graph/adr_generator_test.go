package graph

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestADRGenerator_Generate(t *testing.T) {
	// Setup
	tempDir := "temp_adr_test"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	gen := &ADRGenerator{basePath: tempDir}
	taskID := "test-task"
	// Simulating the fullIntent that ADRTool sends
	fullIntent := "Title\n\nContext: Context details\nDecision: The decision made\nConsequences: Resulting trade-offs"

	path, err := gen.Generate(taskID, fullIntent)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open generated ADR: %v", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Failed to read generated ADR: %v", err)
	}

	contentStr := string(content)

	// Check if the detailed parts are present
	if !strings.Contains(contentStr, "Context details") {
		t.Errorf("Generated ADR missing context details. Content: %s", contentStr)
	}
	if !strings.Contains(contentStr, "The decision made") {
		t.Errorf("Generated ADR missing decision. Content: %s", contentStr)
	}
	if !strings.Contains(contentStr, "Resulting trade-offs") {
		t.Errorf("Generated ADR missing consequences. Content: %s", contentStr)
	}
}
