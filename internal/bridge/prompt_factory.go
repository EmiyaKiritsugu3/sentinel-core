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
		return "", err
	}

	// 1. Coleta ADRs relevantes
	adrs, _ := f.loadADRs()

	// 2. Coleta Standards de Elite (O Motor de Aprendizado)
	standards, _ := os.ReadFile("docs/process/ENGINEERING-STANDARDS.md")

	// 3. Coleta Contexto Cirúrgico
	nodes, _ := f.loadSurgicalContext(taskID)

	data := PromptData{
		Task:                task,
		ADRs:                adrs,
		Standards:           string(standards),
		ContextNodes:        nodes,
		VerificationCommand: verifyCmd,
	}

	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	err = tmpl.Execute(&out, data)
	return out.String(), err
}

func (f *Factory) loadADRs() ([]ADR, error) {
	path := "docs/architecture/SENTINEL-SYSTEM-DESIGN.md"
	content, _ := os.ReadFile(path)
	return []ADR{{Title: "System Design", Content: string(content)}}, nil
}

func (f *Factory) loadSurgicalContext(taskID string) ([]ContextNode, error) {
	rows, err := f.db.Conn.Query("SELECT name, type, file_path, start_line, end_line FROM nodes WHERE type IN ('struct', 'function') LIMIT 10")
	if err != nil {
		return nil, err
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
		return "", fmt.Errorf("invalid line range")
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
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
		return "", fmt.Errorf("error reading lines: %w", err)
	}

	return strings.Join(result, "\n"), nil
}
