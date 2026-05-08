package agents

import (
	"context"
	"os"
	"path/filepath"
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

func TestDecomposeTool_ValidateArguments_InvalidSubtaskType(t *testing.T) {
	tool := &DecomposeTool{}
	args := map[string]interface{}{
		"subtasks": []interface{}{
			"not-a-map",
		},
	}
	err := tool.ValidateArguments(nil, args)
	if err == nil {
		t.Fatal("expected error for invalid subtask object type")
	}
}

func TestDecomposeTool_ValidateArguments_NonStringCapability(t *testing.T) {
	tool := &DecomposeTool{}
	args := map[string]interface{}{
		"subtasks": []interface{}{
			map[string]interface{}{"description": "do stuff", "capabilities": []interface{}{42}, "branch_name": "b1"},
		},
	}
	err := tool.ValidateArguments(nil, args)
	if err == nil {
		t.Fatal("expected error for non-string capability")
	}
}

func TestDecomposeTool_ValidateArguments_MissingSubtasksKey(t *testing.T) {
	tool := &DecomposeTool{}
	err := tool.ValidateArguments(nil, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing subtasks key")
	}
}

func TestDecomposeTool_ValidateArguments_SubtasksNotArray(t *testing.T) {
	tool := &DecomposeTool{}
	args := map[string]interface{}{
		"subtasks": "not-an-array",
	}
	err := tool.ValidateArguments(nil, args)
	if err == nil {
		t.Fatal("expected error for subtasks not being an array")
	}
}

func TestReadFileTool_Execute_Success(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "hello.txt")
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tool := &ReadFileTool{}
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": tmpFile,
	})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if result != "line1\nline2\nline3" {
		t.Errorf("Execute() = %q, want %q", result, "line1\nline2\nline3")
	}
}

func TestReadFileTool_Execute_LineRange(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "ranged.txt")
	if err := os.WriteFile(tmpFile, []byte("a\nb\nc\nd\ne\n"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tool := &ReadFileTool{}
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":       tmpFile,
		"start_line": float64(2),
		"end_line":   float64(4),
	})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if result != "b\nc\nd" {
		t.Errorf("Execute() = %q, want %q", result, "b\nc\nd")
	}
}

func TestRunTool_Execute(t *testing.T) {
	tool := &RunTool{}
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo hello",
	})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if result != "hello\n" {
		t.Errorf("Execute() = %q, want %q", result, "hello\n")
	}
}

func TestRunTool_Execute_EmptyCommand(t *testing.T) {
	tool := &RunTool{}
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "",
	})
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestGrepSearchTool_ValidateArguments_DefaultDir(t *testing.T) {
	tool := &GrepSearchTool{}
	err := tool.ValidateArguments(nil, map[string]interface{}{})
	if err != nil {
		t.Errorf("ValidateArguments with no dir_path should succeed, got: %v", err)
	}
}

func TestADRTool_ValidateArguments_MissingField(t *testing.T) {
	tool := &ADRTool{}
	err := tool.ValidateArguments(nil, map[string]interface{}{
		"title":   "test",
		"context": "ctx",
	})
	if err == nil {
		t.Fatal("expected error for missing ADR fields")
	}
}

func TestWriteFileTool_Execute_InvalidGoCode(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad.go")

	tool := &WriteFileTool{}
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    tmpFile,
		"content": "package main\nfunc broken({",
	})
	if err == nil {
		t.Fatal("expected Gate B rejection for invalid Go syntax")
	}
}

func TestReplaceTool_Execute_InvalidGoCode(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "replace.go")
	initialContent := "package main\n\nfunc hello() {}\n"
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tool := &ReplaceTool{}
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":       tmpFile,
		"old_string": "func hello() {}",
		"new_string": "func broken({",
	})
	if err == nil {
		t.Fatal("expected Gate B rejection for invalid Go syntax after replace")
	}
}
