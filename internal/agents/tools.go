package agents

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/generative-ai-go/genai"
)

// --- ReadFileTool ---

type ReadFileTool struct {
	db *sqlite.DB
}

func (t *ReadFileTool) Name() string        { return "read_file" }
func (t *ReadFileTool) Description() string { return "Reads a file from the project directory. Supports line range." }

func (t *ReadFileTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"path": {
					Type:        genai.TypeString,
					Description: "Path to the file relative to project root.",
				},
				"start_line": {
					Type:        genai.TypeInteger,
					Description: "1-based line number to start reading from.",
				},
				"end_line": {
					Type:        genai.TypeInteger,
					Description: "1-based line number to end reading at.",
				},
			},
			Required: []string{"path"},
		},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("missing path argument")
	}

	start := 1
	if s, ok := args["start_line"].(float64); ok {
		start = int(s)
	}
	end := 1000
	if e, ok := args["end_line"].(float64); ok {
		end = int(e)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("read_file: failed to open %s: %w", path, err)
	}
	defer file.Close()

	var result []string
	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine >= start && currentLine <= end {
			result = append(result, scanner.Text())
		}
		if currentLine > end {
			break
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read_file: error scanning %s: %w", path, err)
	}

	return strings.Join(result, "\n"), nil
}

// --- WriteFileTool ---

type WriteFileTool struct {
	db *sqlite.DB
}

func (t *WriteFileTool) Name() string        { return "write_file" }
func (t *WriteFileTool) Description() string { return "Writes content to a file. Overwrites existing files." }

func (t *WriteFileTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"path": {
					Type:        genai.TypeString,
					Description: "Path to the file relative to project root.",
				},
				"content": {
					Type:        genai.TypeString,
					Description: "The complete content to write to the file.",
				},
			},
			Required: []string{"path", "content"},
		},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)

	// Garante que o diretório pai exista
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("write_file: failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write_file: failed to write file %s: %w", path, err)
	}

	return fmt.Sprintf("File %s written successfully.", path), nil
}

// --- ReplaceTool ---

type ReplaceTool struct {
	db *sqlite.DB
}

func (t *ReplaceTool) Name() string        { return "replace" }
func (t *ReplaceTool) Description() string { return "Replaces a specific string in a file with new content." }

func (t *ReplaceTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"path": {
					Type:        genai.TypeString,
					Description: "Path to the file relative to project root.",
				},
				"old_string": {
					Type:        genai.TypeString,
					Description: "The exact literal text to search for.",
				},
				"new_string": {
					Type:        genai.TypeString,
					Description: "The literal text to replace it with.",
				},
			},
			Required: []string{"path", "old_string", "new_string"},
		},
	}
}

func (t *ReplaceTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	oldStr, _ := args["old_string"].(string)
	newStr, _ := args["new_string"].(string)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("replace: failed to read file %s: %w", path, err)
	}

	content := string(data)
	if !strings.Contains(content, oldStr) {
		return "", fmt.Errorf("replace: old_string not found in file %s", path)
	}

	newContent := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("replace: failed to write file %s: %w", path, err)
	}

	return fmt.Sprintf("Successfully replaced content in %s.", path), nil
}

// --- ADRTool ---

type ADRTool struct {
	db *sqlite.DB
}

func (t *ADRTool) Name() string        { return "sentinel:adr" }
func (t *ADRTool) Description() string { return "Generates a formal Architectural Decision Record (ADR) file for the current task." }

func (t *ADRTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"title": {
					Type:        genai.TypeString,
					Description: "A concise title for the architectural decision.",
				},
				"context": {
					Type:        genai.TypeString,
					Description: "Detailed technical context and the problem being solved.",
				},
				"decision": {
					Type:        genai.TypeString,
					Description: "The technical approach, tools, and patterns chosen.",
				},
				"consequences": {
					Type:        genai.TypeString,
					Description: "Expected trade-offs (positive and negative).",
				},
			},
			Required: []string{"title", "context", "decision", "consequences"},
		},
	}
}

func (t *ADRTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	for _, field := range []string{"title", "context", "decision", "consequences"} {
		if _, ok := args[field].(string); !ok {
			return fmt.Errorf("adr tool: missing or invalid '%s'", field)
		}
	}
	return nil
}

func (t *ADRTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// 1. Get Active Task from DB
	manager := state.NewManager(t.db)
	task, err := manager.GetActiveTask()
	if err != nil {
		return "", fmt.Errorf("adr tool: failed to get active task: %w", err)
	}

	// 2. Format content for ADRGenerator
	gen := graph.NewADRGenerator()
	title, _ := args["title"].(string)
	contextStr, _ := args["context"].(string)
	decision, _ := args["decision"].(string)
	consequences, _ := args["consequences"].(string)

	fullIntent := fmt.Sprintf("%s\n\nContext: %s\nDecision: %s\nConsequences: %s",
		title, contextStr, decision, consequences)

	path, err := gen.Generate(task.ID, fullIntent)
	if err != nil {
		return "", fmt.Errorf("adr tool: generation failed: %w", err)
	}

	return fmt.Sprintf("ADR successfully generated and linked to task [%s] at: %s", task.ID, path), nil
}

// --- ScanTool ---

type ScanTool struct {
	db *sqlite.DB
}

func (t *ScanTool) Name() string        { return "sentinel_scan" }
func (t *ScanTool) Description() string { return "Updates the architectural graph by scanning the project's source code." }

func (t *ScanTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
		},
	}
}

func (t *ScanTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	engine := graph.NewEngine(t.db)
	engine.RegisterScanner(graph.NewGoScanner())
	engine.RegisterScanner(graph.NewTreeSitterScanner())
	
	if err := engine.ScanProject("."); err != nil {
		return "", fmt.Errorf("scan: failed: %w", err)
	}
	
	return "Scan complete. Graph database updated successfully.", nil
}

// RegisterCoreTools adiciona as ferramentas fundamentais ao registro.
func RegisterCoreTools(r *Registry, db *sqlite.DB) {
	r.Tools["read_file"] = &ReadFileTool{db: db}
	r.Tools["write_file"] = &WriteFileTool{db: db}
	r.Tools["replace"] = &ReplaceTool{db: db}
	r.Tools["sentinel_scan"] = &ScanTool{db: db}
	r.Tools["sentinel:adr"] = &ADRTool{db: db}
}
