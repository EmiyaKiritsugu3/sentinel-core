package agents

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/shlex"
	"github.com/google/uuid"
)

// --- ReadFileTool ---

type ReadFileTool struct {
	db *sqlite.DB
}

func (t *ReadFileTool) Name() string { return "read_file" }
func (t *ReadFileTool) Description() string {
	return "Reads a file from the project directory. Supports line range."
}

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

func (t *ReadFileTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	path, ok := args["path"].(string)
	if !ok {
		return fmt.Errorf("missing 'path' argument")
	}
	return v.ValidatePath(path)
}

func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)

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

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Writes content to a file. Overwrites existing files."
}

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

func (t *WriteFileTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	path, ok := args["path"].(string)
	if !ok {
		return fmt.Errorf("missing 'path' argument")
	}
	return v.ValidatePath(path)
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

func (t *ReplaceTool) Name() string { return "replace" }
func (t *ReplaceTool) Description() string {
	return "Replaces a specific string in a file with new content."
}

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

func (t *ReplaceTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	path, ok := args["path"].(string)
	if !ok {
		return fmt.Errorf("missing 'path' argument")
	}
	return v.ValidatePath(path)
}

func (t *ReplaceTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	oldStr, _ := args["old_string"].(string)
	newStr, _ := args["new_string"].(string)

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("replace: failed to open %s: %w", path, err)
	}
	defer file.Close()

	var sb strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sb.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("replace: error scanning %s: %w", path, err)
	}
	content := sb.String()

	if !strings.Contains(content, oldStr) {
		return "", fmt.Errorf("replace: old_string not found in file %s", path)
	}

	newContent := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("replace: failed to write file %s: %w", path, err)
	}

	return fmt.Sprintf("Successfully replaced content in %s.", path), nil
}

// --- GrepSearchTool ---

type GrepSearchTool struct {
	db *sqlite.DB
}

func (t *GrepSearchTool) Name() string { return "grep_search" }
func (t *GrepSearchTool) Description() string {
	return "Searches for a regular expression pattern within file contents across the project."
}

func (t *GrepSearchTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"pattern": {
					Type:        genai.TypeString,
					Description: "The regular expression pattern to search for.",
				},
				"dir_path": {
					Type:        genai.TypeString,
					Description: "Directory to search in (relative to project root). Defaults to '.'.",
				},
			},
			Required: []string{"pattern"},
		},
	}
}

func (t *GrepSearchTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	if dir, ok := args["dir_path"].(string); ok {
		return v.ValidatePath(dir)
	}
	return nil
}

func (t *GrepSearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	pattern, _ := args["pattern"].(string)
	dir, ok := args["dir_path"].(string)
	if !ok {
		dir = "."
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("grep_search: invalid regex: %w", err)
	}

	var matches []string
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only scan text files (simple heuristic)
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".md" && ext != ".json" && ext != ".yaml" && ext != ".yml" && ext != ".sql" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil // Skip files we can't open
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 1
		for scanner.Scan() {
			if re.MatchString(scanner.Text()) {
				matches = append(matches, fmt.Sprintf("%s:%d: %s", path, lineNum, scanner.Text()))
			}
			if len(matches) > 100 {
				return fmt.Errorf("too many matches found")
			}
			lineNum++
		}
		return nil
	})

	if err != nil && err.Error() != "too many matches found" {
		return "", fmt.Errorf("grep_search: walk error: %w", err)
	}

	if len(matches) == 0 {
		return "No matches found.", nil
	}

	result := strings.Join(matches, "\n")
	if len(matches) > 100 {
		result += "\n... [TRUNCATED] Too many results."
	}
	return result, nil
}

// --- AuditTool ---

type AuditTool struct {
	db *sqlite.DB
}

func (t *AuditTool) Name() string { return "sentinel:audit" }
func (t *AuditTool) Description() string {
	return "Runs the Sovereign Validator across the project to detect Standard violations."
}

func (t *AuditTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
		},
	}
}

func (t *AuditTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	return nil
}

func (t *AuditTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	v := reflect.NewValidator(t.db)
	violations, err := v.ValidateProject(".")
	if err != nil {
		return "", fmt.Errorf("audit tool: %w", err)
	}

	if len(violations) == 0 {
		return "Sovereign Audit: 0 violations found. System is compliant.", nil
	}

	var report strings.Builder
	report.WriteString(fmt.Sprintf("Sovereign Audit Report: %d violation(s) found\n", len(violations)))
	for i, v := range violations {
		if i >= 30 {
			report.WriteString("\n... [TRUNCATED] Please fix the above violations first.")
			break
		}
		report.WriteString(fmt.Sprintf("- [%s] %s:%d: %s\n", v.StandardID, v.FilePath, v.Line, v.Reason))
	}

	return report.String(), nil
}

// --- RunTool ---

type RunTool struct {
	db *sqlite.DB
}

func (t *RunTool) Name() string { return "sentinel:run" }
func (t *RunTool) Description() string {
	return "Runs a safe, approved shell command (e.g., 'go build ./...', 'go test -v ./...'). Does not support pipes or redirection."
}

func (t *RunTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"command": {
					Type:        genai.TypeString,
					Description: "The shell command to execute.",
				},
			},
			Required: []string{"command"},
		},
	}
}

func (t *RunTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	cmd, ok := args["command"].(string)
	if !ok {
		return fmt.Errorf("missing 'command' argument")
	}
	return v.ValidateCommand(cmd)
}

func (t *RunTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	cmdStr, _ := args["command"].(string)

	parts, err := shlex.Split(cmdStr)
	if err != nil {
		return "", fmt.Errorf("run: failed to parse command: %w", err)
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("run: empty command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()

	output := out.String()
	// [PID-SENTINEL] Auditor Constraint: Context Protection
	// Limit output to ~10KB or 200 lines to prevent context exhaustion
	lines := strings.Split(output, "\n")
	if len(lines) > 200 {
		output = strings.Join(lines[:200], "\n") + "\n... [TRUNCATED] Too many lines of output."
	}
	if len(output) > 10000 {
		output = output[:10000] + "\n... [TRUNCATED] Output too large."
	}

	if err != nil {
		return fmt.Sprintf("Command failed with error: %v\n\nOutput:\n%s", err, output), nil
	}

	return output, nil
}

// --- ADRTool ---

type ADRTool struct {
	db *sqlite.DB
}

func (t *ADRTool) Name() string { return "sentinel:adr" }
func (t *ADRTool) Description() string {
	return "Generates a formal Architectural Decision Record (ADR) file for the current task."
}

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
				"verification_command": {
					Type:        genai.TypeString,
					Description: "A shell command (e.g., go test) to verify the implementation.",
				},
			},
			Required: []string{"title", "context", "decision", "consequences", "verification_command"},
		},
	}
}

func (t *ADRTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	for _, field := range []string{"title", "context", "decision", "consequences", "verification_command"} {
		if _, ok := args[field].(string); !ok {
			return fmt.Errorf("adr tool: missing or invalid '%s'", field)
		}
	}
	return nil
}

func (t *ADRTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	manager := state.NewManager(t.db)
	task, err := manager.GetActiveTask()
	if err != nil {
		return "", fmt.Errorf("adr tool: failed to get active task: %w", err)
	}

	gen := graph.NewADRGenerator()
	title, _ := args["title"].(string)
	contextStr, _ := args["context"].(string)
	decision, _ := args["decision"].(string)
	consequences, _ := args["consequences"].(string)
	verification, _ := args["verification_command"].(string)

	path, err := gen.Generate(graph.ADRData{
		TaskID:              task.ID,
		Title:               title,
		Context:             contextStr,
		Decision:            decision,
		Consequences:        consequences,
		VerificationCommand: verification,
		Status:              "PROPOSED",
	})
	if err != nil {
		return "", fmt.Errorf("adr tool: generation failed: %w", err)
	}

	return fmt.Sprintf("ADR successfully generated and linked to task [%s] at: %s", task.ID, path), nil
}

// --- ScanTool ---

type ScanTool struct {
	db *sqlite.DB
}

func (t *ScanTool) Name() string { return "sentinel_scan" }
func (t *ScanTool) Description() string {
	return "Updates the architectural graph by scanning the project's source code."
}

func (t *ScanTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
		},
	}
}

func (t *ScanTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	return nil
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

// --- DecomposeTool ---

type DecomposeTool struct {
	db *sqlite.DB
}

func (t *DecomposeTool) Name() string { return "sentinel:decompose" }
func (t *DecomposeTool) Description() string {
	return "Decomposes a complex task into multiple atomic sub-tasks for parallel execution."
}

func (t *DecomposeTool) Definition() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"subtasks": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"description": {
								Type:        genai.TypeString,
								Description: "Detailed intent of the sub-task.",
							},
							"capabilities": {
								Type: genai.TypeArray,
								Items: &genai.Schema{
									Type: genai.TypeString,
								},
								Description: "List of required capabilities (e.g., 'go', 'git').",
							},
							"branch_name": {
								Type:        genai.TypeString,
								Description: "The ephemeral branch where the sub-task will operate.",
							},
						},
						Required: []string{"description", "capabilities", "branch_name"},
					},
				},
			},
			Required: []string{"subtasks"},
		},
	}
}

func (t *DecomposeTool) ValidateArguments(v *reflect.Validator, args map[string]interface{}) error {
	subtasksRaw, ok := args["subtasks"]
	if !ok {
		return fmt.Errorf("decompose: missing 'subtasks' array")
	}

	subtasks, ok := subtasksRaw.([]interface{})
	if !ok {
		return fmt.Errorf("decompose: 'subtasks' must be an array")
	}

	// KISS Hard Gate: Max 5 subtasks
	if len(subtasks) > 5 {
		return fmt.Errorf("decompose: security violation - maximum of 5 subtasks allowed per goal")
	}

	for _, st := range subtasks {
		task, ok := st.(map[string]interface{})
		if !ok {
			return fmt.Errorf("decompose: invalid subtask object")
		}
		if desc, ok := task["description"].(string); !ok || desc == "" {
			return fmt.Errorf("decompose: subtask description is required")
		}
		if branch, ok := task["branch_name"].(string); !ok || branch == "" {
			return fmt.Errorf("decompose: subtask branch_name is required")
		}

		// Issue 5: Validate capabilities
		capsRaw, ok := task["capabilities"]
		if !ok {
			return fmt.Errorf("decompose: subtask capabilities array is required")
		}
		caps, ok := capsRaw.([]interface{})
		if !ok {
			return fmt.Errorf("decompose: capabilities must be an array")
		}
		for _, c := range caps {
			if _, ok := c.(string); !ok {
				return fmt.Errorf("decompose: capability must be a string")
			}
		}
	}
	return nil
}

func (t *DecomposeTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Issue 2: Defensive validation before processing
	if err := t.ValidateArguments(nil, args); err != nil {
		return "", err
	}

	manager := state.NewManager(t.db)
	parentTask, err := manager.GetActiveTask()
	if err != nil {
		return "", fmt.Errorf("decompose: no active task: %w", err)
	}

	subtasks := args["subtasks"].([]interface{})
	var results []string

	// Issue 6: Atomic sub-task insertion via transaction
	tx, err := t.db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("decompose: failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	for _, stRaw := range subtasks {
		st := stRaw.(map[string]interface{})
		description := st["description"].(string)
		branch := st["branch_name"].(string)
		capabilities := st["capabilities"].([]interface{})

		capsJSON, _ := json.Marshal(capabilities)
		id := uuid.New().String()[:8]

		query := `INSERT INTO sub_tasks (id, parent_task_id, description, status, branch_name, required_capabilities) VALUES (?, ?, ?, ?, ?, ?)`
		_, err = tx.ExecContext(ctx, query, id, parentTask.ID, description, "PENDING", branch, string(capsJSON))
		if err != nil {
			return "", fmt.Errorf("decompose: failed to insert sub-task %s: %w", id, err)
		}
		results = append(results, id)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("decompose: failed to commit transaction: %w", err)
	}

	return fmt.Sprintf("Successfully decomposed task into %d sub-tasks: %s", len(results), strings.Join(results, ", ")), nil
}

// RegisterCoreTools adiciona as ferramentas fundamentais ao registro.
func RegisterCoreTools(r *Registry, db *sqlite.DB) {
	r.Tools["read_file"] = &ReadFileTool{db: db}
	r.Tools["write_file"] = &WriteFileTool{db: db}
	r.Tools["replace"] = &ReplaceTool{db: db}
	r.Tools["sentinel_scan"] = &ScanTool{db: db}
	r.Tools["grep_search"] = &GrepSearchTool{db: db}
	r.Tools["sentinel:audit"] = &AuditTool{db: db}
	r.Tools["sentinel:run"] = &RunTool{db: db}
	r.Tools["sentinel:adr"] = &ADRTool{db: db}
	r.Tools["sentinel:decompose"] = &DecomposeTool{db: db}
}
