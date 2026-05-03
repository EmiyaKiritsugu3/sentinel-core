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
	data := ADRData{
		TaskID:              "test-task",
		Title:               "Test Title",
		Context:             "Context details",
		Decision:            "The decision made",
		Consequences:        "Resulting trade-offs",
		VerificationCommand: "go test ./...",
		Status:              "PROPOSED",
	}

	path, err := gen.Generate(data)
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
		t.Errorf("Generated ADR missing context details")
	}
	if !strings.Contains(contentStr, "The decision made") {
		t.Errorf("Generated ADR missing decision")
	}
	if !strings.Contains(contentStr, "Protocolo de Verificação") {
		t.Errorf("Generated ADR missing Verification Protocol section")
	}
	if !strings.Contains(contentStr, "go test ./...") {
		t.Errorf("Generated ADR missing verification command")
	}
}
