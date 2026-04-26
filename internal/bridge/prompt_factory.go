package bridge

import (
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

	// 1. Coleta ADRs relevantes (por enquanto, o principal)
	adrs, _ := f.loadADRs()

	// 2. Coleta Contexto Cirúrgico (Simulado para o MVP, em breve via Grafo)
	nodes, _ := f.loadSurgicalContext(taskID)

	data := PromptData{
		Task:                task,
		ADRs:                adrs,
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
	// Carrega o SENTINEL-SYSTEM-DESIGN.md como ADR base
	path := "docs/architecture/SENTINEL-SYSTEM-DESIGN.md"
	content, _ := os.ReadFile(path)
	return []ADR{{Title: "System Design", Content: string(content)}}, nil
}

func (f *Factory) loadSurgicalContext(taskID string) ([]ContextNode, error) {
	// 1. Busca os nós que o banco diz que pertencem a esta tarefa ou ao projeto
	// Para o MVP, pegamos todos os nós do tipo 'struct' e 'function' 
	// (Futuramente: filtro baseado no grafo de impacto da taskID)
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

		// 2. Extração Real do Código no Disco
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

// extractLines lê um arquivo e retorna apenas o range de linhas solicitado
func extractLines(path string, start, end int) (string, error) {
	if start <= 0 || end <= 0 {
		return "", fmt.Errorf("invalid line range")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if start > len(lines) {
		return "", fmt.Errorf("start line out of bounds")
	}
	if end > len(lines) {
		end = len(lines)
	}

	return strings.Join(lines[start-1:end], "\n"), nil
}
