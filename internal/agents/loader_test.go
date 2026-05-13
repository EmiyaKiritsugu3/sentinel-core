package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_MaxLambdaValidation(t *testing.T) {
	t.Parallel()
	loader := NewLoader()

	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
	}{
		{
			name: "Valid max_lambda",
			yamlContent: `---
name: "Test Agent"
model_id: "test-model"
max_steps: 10
max_lambda: 0.5
---
System prompt body`,
			wantErr: false,
		},
		{
			name: "Omitted max_lambda is valid (omitempty)",
			yamlContent: `---
name: "Test Agent"
model_id: "test-model"
max_steps: 10
---
System prompt body`,
			wantErr: false,
		},
		{
			name: "Explicit zero max_lambda is invalid (min=0.1)",
			yamlContent: `---
name: "Test Agent"
model_id: "test-model"
max_steps: 10
max_lambda: 0.0
---
System prompt body`,
			wantErr: true, // Should fail min=0.1 validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary file for the test

			tempFile := filepath.Join(t.TempDir(), "agent.md")
			err := os.WriteFile(tempFile, []byte(tt.yamlContent), 0644) //nolint:gosec // test fixture
			if err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}

			// Load the agent
			def, err := loader.LoadAgent(tempFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected an error but got nil, definition: %+v", def)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
