package bridge

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

const promptTemplate = `
# SENTINEL SOVEREIGN INSTRUCTION [PID-SENTINEL]
**Task ID**: {{.Task.ID}}
**Protocol Status**: ACTIVE (Strict Mode)

## 🎯 OBJECTIVE
{{.Task.Description}}

## 📜 ARCHITECTURAL CONSTRAINTS (ADRs)
{{range .ADRs}}
### {{.Title}}
{{.Content}}
{{end}}

## 💎 ENGINEERING STANDARDS (MANDATORY)
{{.Standards}}

## 🕸️ SURGICAL CONTEXT (AST NODES)
{{range .ContextNodes}}
---
**Symbol**: {{.Name}} ({{.Type}})
**Location**: {{.FilePath}} [Lines {{.StartLine}}-{{.EndLine}}]
{{.CodeSnippet}}
{{end}}

## 🛡️ RULES OF ENGAGEMENT
1. **Scope Integrity**: You are authorized to modify ONLY files listed in the context.
2. **Error Governance**: Use project-specific standard error classes. No generic Errors.
3. **Traceability**: All logic must align with the ADRs provided above.
4. **No Devanios**: Do not refactor unrelated code. Do not add undocumented features.

## 🏁 VERIFICATION GATE
Upon completion, I will execute: 
` + "`{{.VerificationCommand}}`" + `
Your changes will be REJECTED if this command returns a non-zero exit code.

**Proceed with implementation.**
`

type ADR struct {
	Title   string
	Content string
}

type ContextNode struct {
	Name        string
	Type        string
	FilePath    string
	StartLine   int
	EndLine     int
	CodeSnippet string
}

type PromptData struct {
	Task                *state.Task
	ADRs                []ADR
	Standards           string
	ContextNodes        []ContextNode
	VerificationCommand string
}

type Factory struct {
	db *sqlite.DB
}

func NewFactory(db *sqlite.DB) *Factory {
	return &Factory{db: db}
}

// GenerateInstruction constrói o prompt final de elite
func (f *Factory) GenerateInstruction(taskID string) (string, error) {
	mgr := state.NewManager(f.db)
	task, verifyCmd, err := mgr.GetTaskByID(taskID)
	if err != nil {
		return "", fmt.Errorf("bridge: failed to get task %s: %w", taskID, err)
	}

	// 1. Coleta ADRs relevantes
	adrs, err := f.loadADRs()
	if err != nil {
		return "", fmt.Errorf("bridge: failed to load ADRs: %w", err)
	}

	// 2. Coleta Standards de Elite (O Motor de Aprendizado)
	// Standard #01: Buffered Read via extractLines
	standards, err := extractLines("docs/process/ENGINEERING-STANDARDS.md", 1, 100)
	if err != nil {
		return "", fmt.Errorf("bridge: failed to load standards: %w", err)
	}

	// 3. Coleta Contexto Cirúrgico
	nodes, err := f.loadSurgicalContext(taskID)
	if err != nil {
		return "", fmt.Errorf("bridge: failed to load surgical context: %w", err)
	}

	data := PromptData{
		Task:                task,
		ADRs:                adrs,
		Standards:           standards,
		ContextNodes:        nodes,
		VerificationCommand: verifyCmd,
	}

	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("bridge: template parse error: %w", err)
	}

	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("bridge: template execution error: %w", err)
	}
	return out.String(), nil
}

func (f *Factory) loadADRs() ([]ADR, error) {
	path := "docs/architecture/SENTINEL-SYSTEM-DESIGN.md"
	// Standard #01: Buffered Read via extractLines
	content, err := extractLines(path, 1, 500)
	if err != nil {
		return nil, fmt.Errorf("bridge: failed to read ADR file: %w", err)
	}
	return []ADR{{Title: "System Design", Content: content}}, nil
}

func (f *Factory) loadSurgicalContext(taskID string) ([]ContextNode, error) {
	rows, err := f.db.Conn.Query("SELECT name, type, file_path, start_line, end_line FROM nodes WHERE type IN ('struct', 'function') LIMIT 10")
	if err != nil {
		return nil, fmt.Errorf("bridge: db query error: %w", err)
	}
	defer rows.Close()

	var nodes []ContextNode
	for rows.Next() {
		var n ContextNode
		if err := rows.Scan(&n.Name, &n.Type, &n.FilePath, &n.StartLine, &n.EndLine); err != nil {
			continue
		}

		snippet, err := extractLines(n.FilePath, n.StartLine, n.EndLine)
		if err == nil {
			n.CodeSnippet = snippet
		} else {
			n.CodeSnippet = fmt.Sprintf("// Error extracting code: %v", err)
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func extractLines(path string, start, end int) (string, error) {
	if start <= 0 || end <= 0 {
		return "", fmt.Errorf("extract: invalid line range %d-%d", start, end)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("extract: failed to open %s: %w", path, err)
	}
	defer file.Close()

	var result []string
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

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
		return "", fmt.Errorf("extract: error scanning %s: %w", path, err)
	}

	return strings.Join(result, "\n"), nil
}
