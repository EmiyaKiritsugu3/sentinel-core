package bridge

import (
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
	// Aqui o motor de grafos entraria em ação. 
	// Para o MVP, retornamos um nó placeholder.
	return []ContextNode{
		{
			Name:        "AuditRunner",
			Type:        "struct",
			FilePath:    "internal/audit/runner.go",
			StartLine:   1,
			EndLine:     10,
			CodeSnippet: "// [Code extraction logic coming soon]",
		},
	}, nil
}
