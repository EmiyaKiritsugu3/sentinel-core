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

## 🏁 OUTPUT FORMAT (MANDATORY)
Your response must conclude with a **Sovereign Audit Report** (Standard #08) using exactly these 5 points:
1. ✨ **The Good**: (What is now solid)
2. ⚠️ **The Bad**: (Technical debt introduced)
3. 💥 **The Ugly**: (Riscos e fragilidades detectadas)
4. 💡 **The Lesson**: (What was learned/standardized)
5. 🚀 **The Next**: (Next optimization)

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

type ContextPayload struct {
	SystemInstruction   string
	SurgicalContext     string
	TaskDescription     string
	VerificationCommand string
}

type Factory struct {
	db *sqlite.DB
}

func NewFactory(db *sqlite.DB) *Factory {
	return &Factory{db: db}
}

// GeneratePayload constrói o payload estruturado para a Engine
func (f *Factory) GeneratePayload(taskID string, personaPrompt string) (*ContextPayload, error) {
	mgr := state.NewManager(f.db)
	task, verifyCmd, err := mgr.GetTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("bridge: failed to get task %s: %w", taskID, err)
	}

	adrs, err := f.loadADRs()
	if err != nil {
		return nil, fmt.Errorf("bridge: failed to load ADRs: %w", err)
	}

	standards, err := extractLines("docs/process/ENGINEERING-STANDARDS.md", 1, 100)
	if err != nil {
		return nil, fmt.Errorf("bridge: failed to load standards: %w", err)
	}

	nodes, err := f.loadSurgicalContext(taskID)
	if err != nil {
		return nil, fmt.Errorf("bridge: failed to load surgical context: %w", err)
	}

	// System Instruction: Persona + ADRs + Standards
	systemTmpl := `
	# PERSONA
	{{.Persona}}

	# ARCHITECTURAL CONSTRAINTS (ADRs)
	{{range .ADRs}}
	## {{.Title}}
	{{.Content}}
	{{end}}

	# ENGINEERING STANDARDS
	{{.Standards}}

	# RULES OF ENGAGEMENT
	1. **Scope Integrity**: You are authorized to modify ONLY files listed in the context.
	2. **Error Governance**: Use project-specific standard error classes. No generic Errors.
	3. **Traceability**: All logic must align with the ADRs provided above.
	4. **No Devanios**: Do not refactor unrelated code. Do not add undocumented features.
	{{if or (eq .Tier "T2") (eq .Tier "T3")}}
	5. **[GOVERNANCE]**: For Tier 2 (Structural) or Tier 3 (Architectural) tasks, you MUST use the 'sentinel:adr' tool to document your decision BEFORE modifying any code files. Implementation without a registered ADR is a violation of Standard #14.
	{{end}}

	# OUTPUT FORMAT (MANDATORY)
	Your response must conclude with a **Sovereign Audit Report** (Standard #08) using exactly these 5 points:
	1. ✨ **The Good**: (What is now solid)
	2. ⚠️ **The Bad**: (Technical debt introduced)
	3. 💥 **The Ugly**: (Riscos e fragilidades detectadas)
	4. 💡 **The Lesson**: (What was learned/standardized)
	5. 🚀 **The Next**: (Next optimization)
	`

	type SystemData struct {
		Persona   string
		ADRs      []ADR
		Standards string
		Tier      string
	}

	tmpl, err := template.New("system").Parse(systemTmpl)
	if err != nil {
		return nil, fmt.Errorf("bridge: system template parse error: %w", err)
	}

	var systemOut strings.Builder
	if err := tmpl.Execute(&systemOut, SystemData{
		Persona:   personaPrompt,
		ADRs:      adrs,
		Standards: standards,
		Tier:      task.Tier,
	}); err != nil {
		return nil, fmt.Errorf("bridge: system template execution error: %w", err)
	}
	// Surgical Context: Just the code nodes
	var contextOut strings.Builder
	for _, n := range nodes {
		contextOut.WriteString(fmt.Sprintf("\n---\n**Symbol**: %s (%s)\n**Location**: %s [Lines %d-%d]\n%s\n", n.Name, n.Type, n.FilePath, n.StartLine, n.EndLine, n.CodeSnippet))
	}

	return &ContextPayload{
		SystemInstruction:   systemOut.String(),
		SurgicalContext:     contextOut.String(),
		TaskDescription:     task.Description,
		VerificationCommand: verifyCmd,
	}, nil
}

func (f *Factory) loadADRs() ([]ADR, error) {
	path := "docs/architecture/SENTINEL-SYSTEM-DESIGN.md"
	content, err := extractLines(path, 1, 500)
	if err != nil {
		return nil, fmt.Errorf("bridge: failed to read ADR file: %w", err)
	}
	return []ADR{{Title: "System Design", Content: content}}, nil
}

func (f *Factory) loadSurgicalContext(taskID string) ([]ContextNode, error) {
	mgr := state.NewManager(f.db)
	_, _, err := mgr.GetTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("bridge: failed to find task context: %w", err)
	}

	rows, err := f.db.Conn.Query("SELECT name, type, file_path, start_line, end_line FROM nodes WHERE type IN ('struct', 'function') ORDER BY last_indexed DESC LIMIT 10")
	if err != nil {
		return nil, fmt.Errorf("bridge: db query error: %w", err)
	}
	defer rows.Close()

	var nodes []ContextNode
	for rows.Next() {
		var n ContextNode
		if err := rows.Scan(&n.Name, &n.Type, &n.FilePath, &n.StartLine, &n.EndLine); err != nil {
			return nil, fmt.Errorf("bridge: row scan error: %w", err)
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
