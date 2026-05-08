package agents

import (
	"testing"

	"github.com/google/generative-ai-go/genai"
)

func TestReadFileTool_Metadata(t *testing.T) {
	tool := &ReadFileTool{}

	if tool.Name() != "read_file" {
		t.Errorf("ReadFileTool.Name() = %q, want %q", tool.Name(), "read_file")
	}
	if tool.Description() == "" {
		t.Error("ReadFileTool.Description() should not be empty")
	}
	def := tool.Definition()
	if def.Name != "read_file" {
		t.Errorf("ReadFileTool.Definition().Name = %q, want %q", def.Name, "read_file")
	}
	if def.Parameters.Type != genai.TypeObject {
		t.Errorf("ReadFileTool.Definition().Parameters.Type = %v, want TypeObject", def.Parameters.Type)
	}
	if _, ok := def.Parameters.Properties["path"]; !ok {
		t.Error("ReadFileTool.Definition() missing 'path' property")
	}
	if len(def.Parameters.Required) != 1 || def.Parameters.Required[0] != "path" {
		t.Errorf("ReadFileTool.Definition().Required = %v, want [path]", def.Parameters.Required)
	}
}

func TestWriteFileTool_Metadata(t *testing.T) {
	tool := &WriteFileTool{}

	if tool.Name() != "write_file" {
		t.Errorf("WriteFileTool.Name() = %q, want %q", tool.Name(), "write_file")
	}
	if tool.Description() == "" {
		t.Error("WriteFileTool.Description() should not be empty")
	}
	def := tool.Definition()
	if def.Name != "write_file" {
		t.Errorf("WriteFileTool.Definition().Name = %q, want %q", def.Name, "write_file")
	}
	if _, ok := def.Parameters.Properties["path"]; !ok {
		t.Error("WriteFileTool.Definition() missing 'path' property")
	}
	if _, ok := def.Parameters.Properties["content"]; !ok {
		t.Error("WriteFileTool.Definition() missing 'content' property")
	}
	if len(def.Parameters.Required) != 2 {
		t.Errorf("WriteFileTool.Definition().Required = %v, want 2 required fields", def.Parameters.Required)
	}
}

func TestReplaceTool_Metadata(t *testing.T) {
	tool := &ReplaceTool{}

	if tool.Name() != "replace" {
		t.Errorf("ReplaceTool.Name() = %q, want %q", tool.Name(), "replace")
	}
	if tool.Description() == "" {
		t.Error("ReplaceTool.Description() should not be empty")
	}
	def := tool.Definition()
	if def.Name != "replace" {
		t.Errorf("ReplaceTool.Definition().Name = %q, want %q", def.Name, "replace")
	}
}

func TestGrepSearchTool_Metadata(t *testing.T) {
	tool := &GrepSearchTool{}

	if tool.Name() != "grep_search" {
		t.Errorf("GrepSearchTool.Name() = %q, want %q", tool.Name(), "grep_search")
	}
	def := tool.Definition()
	if _, ok := def.Parameters.Properties["pattern"]; !ok {
		t.Error("GrepSearchTool.Definition() missing 'pattern' property")
	}
}

func TestAuditTool_Metadata(t *testing.T) {
	tool := &AuditTool{}

	if tool.Name() != "sentinel:audit" {
		t.Errorf("AuditTool.Name() = %q, want %q", tool.Name(), "sentinel:audit")
	}
	def := tool.Definition()
	if def.Name != "sentinel:audit" {
		t.Errorf("AuditTool.Definition().Name = %q, want %q", def.Name, "sentinel:audit")
	}
}

func TestRunTool_Metadata(t *testing.T) {
	tool := &RunTool{}

	if tool.Name() != "sentinel:run" {
		t.Errorf("RunTool.Name() = %q, want %q", tool.Name(), "sentinel:run")
	}
	def := tool.Definition()
	if _, ok := def.Parameters.Properties["command"]; !ok {
		t.Error("RunTool.Definition() missing 'command' property")
	}
}

func TestADRTool_Metadata(t *testing.T) {
	tool := &ADRTool{}

	if tool.Name() != "sentinel:adr" {
		t.Errorf("ADRTool.Name() = %q, want %q", tool.Name(), "sentinel:adr")
	}
	def := tool.Definition()
	if len(def.Parameters.Required) != 5 {
		t.Errorf("ADRTool.Definition().Required has %d fields, want 5", len(def.Parameters.Required))
	}
}

func TestScanTool_Metadata(t *testing.T) {
	tool := &ScanTool{}

	if tool.Name() != "sentinel_scan" {
		t.Errorf("ScanTool.Name() = %q, want %q", tool.Name(), "sentinel_scan")
	}
	def := tool.Definition()
	if def.Name != "sentinel_scan" {
		t.Errorf("ScanTool.Definition().Name = %q, want %q", def.Name, "sentinel_scan")
	}
}

func TestDecomposeTool_Metadata(t *testing.T) {
	tool := &DecomposeTool{}

	if tool.Name() != "sentinel:decompose" {
		t.Errorf("DecomposeTool.Name() = %q, want %q", tool.Name(), "sentinel:decompose")
	}
	def := tool.Definition()
	if _, ok := def.Parameters.Properties["subtasks"]; !ok {
		t.Error("DecomposeTool.Definition() missing 'subtasks' property")
	}
}

func TestDecomposeTool_ValidateArguments_MaxSubtasks(t *testing.T) {
	tool := &DecomposeTool{}
	args := map[string]interface{}{
		"subtasks": []interface{}{
			map[string]interface{}{"description": "a", "capabilities": []interface{}{"go"}, "branch_name": "b1"},
			map[string]interface{}{"description": "b", "capabilities": []interface{}{"go"}, "branch_name": "b2"},
			map[string]interface{}{"description": "c", "capabilities": []interface{}{"go"}, "branch_name": "b3"},
			map[string]interface{}{"description": "d", "capabilities": []interface{}{"go"}, "branch_name": "b4"},
			map[string]interface{}{"description": "e", "capabilities": []interface{}{"go"}, "branch_name": "b5"},
			map[string]interface{}{"description": "f", "capabilities": []interface{}{"go"}, "branch_name": "b6"},
		},
	}
	err := tool.ValidateArguments(nil, args)
	if err == nil {
		t.Fatal("expected error for >5 subtasks")
	}
}

func TestDecomposeTool_ValidateArguments_MissingDescription(t *testing.T) {
	tool := &DecomposeTool{}
	args := map[string]interface{}{
		"subtasks": []interface{}{
			map[string]interface{}{"capabilities": []interface{}{"go"}, "branch_name": "b1"},
		},
	}
	err := tool.ValidateArguments(nil, args)
	if err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestDecomposeTool_ValidateArguments_MissingBranch(t *testing.T) {
	tool := &DecomposeTool{}
	args := map[string]interface{}{
		"subtasks": []interface{}{
			map[string]interface{}{"description": "do stuff", "capabilities": []interface{}{"go"}},
		},
	}
	err := tool.ValidateArguments(nil, args)
	if err == nil {
		t.Fatal("expected error for missing branch_name")
	}
}

func TestDecomposeTool_ValidateArguments_MissingCapabilities(t *testing.T) {
	tool := &DecomposeTool{}
	args := map[string]interface{}{
		"subtasks": []interface{}{
			map[string]interface{}{"description": "do stuff", "branch_name": "b1"},
		},
	}
	err := tool.ValidateArguments(nil, args)
	if err == nil {
		t.Fatal("expected error for missing capabilities")
	}
}

func TestDecomposeTool_ValidateArguments_Valid(t *testing.T) {
	tool := &DecomposeTool{}
	args := map[string]interface{}{
		"subtasks": []interface{}{
			map[string]interface{}{"description": "do stuff", "capabilities": []interface{}{"go", "git"}, "branch_name": "b1"},
		},
	}
	if err := tool.ValidateArguments(nil, args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
