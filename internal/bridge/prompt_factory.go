package bridge

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

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
	db         *sqlite.DB
	classifier *IntentClassifier
}

func NewFactory(db *sqlite.DB, classifier *IntentClassifier) *Factory {
	return &Factory{db: db, classifier: classifier}
}

// GeneratePayload constrói o payload estruturado para a Engine
func (f *Factory) GeneratePayload(ctx context.Context, taskID string, personaPrompt string) (*ContextPayload, error) {
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

	intent := IntentUnknown
	if f.classifier != nil {
		intent = f.classifier.Classify(ctx, taskID, task.Description)
	}
	strategy := StrategyFor(intent)
	nodes, err := f.loadContextByStrategy(taskID, strategy)
	if err != nil {
		return nil, fmt.Errorf("bridge: failed to load context: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("bridge: row iteration error: %w", err)
	}

	return nodes, nil
}

func (f *Factory) loadContextByStrategy(taskID string, strategy ContextStrategy) ([]ContextNode, error) {
	// Zero-value strategy → use existing default behavior
	if strategy.NodeLimit == 0 {
		return f.loadSurgicalContext(taskID)
	}

	limit := strategy.NodeLimit
	var orderClause string
	if strategy.HighCoupling {
		// Nodes with most incoming edges (highest fan-in)
		orderClause = `ORDER BY (
			SELECT COUNT(*) FROM edges WHERE to_node_id = nodes.id
		) DESC`
	} else {
		orderClause = "ORDER BY last_indexed DESC"
	}

	typeFilter := `type IN ('struct', 'function')`
	if strategy.IncludeTests {
		typeFilter = `type IN ('struct', 'function') OR file_path LIKE '%_test.go'`
	}

	query := fmt.Sprintf(
		"SELECT name, type, file_path, start_line, end_line FROM nodes WHERE %s %s LIMIT %d",
		typeFilter, orderClause, limit,
	)
	rows, err := f.db.Conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("bridge: context query error: %w", err)
	}
	defer rows.Close()

	var nodes []ContextNode
	for rows.Next() {
		var n ContextNode
		if err := rows.Scan(&n.Name, &n.Type, &n.FilePath, &n.StartLine, &n.EndLine); err != nil {
			return nil, fmt.Errorf("bridge: row scan error: %w", err)
		}
		if snippet, err := extractLines(n.FilePath, n.StartLine, n.EndLine); err == nil {
			n.CodeSnippet = snippet
		}
		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("bridge: row iteration error: %w", err)
	}

	// File-based context appended as synthetic nodes
	if strategy.IncludeADRs {
		adrNodes, _ := f.loadADRNodes()
		nodes = append(nodes, adrNodes...)
	}
	if strategy.IncludeDebtMarkers {
		debtPath := "docs/process/TECHNICAL-DEBT.md"
		debtContent, _ := extractLines(debtPath, 1, 100)
		if debtContent != "" {
			nodes = append(nodes, ContextNode{
				Name:        "TECHNICAL-DEBT",
				Type:        "doc",
				FilePath:    debtPath,
				CodeSnippet: debtContent,
			})
		}
	}
	return nodes, nil
}

func (f *Factory) loadADRNodes() ([]ContextNode, error) {
	entries, err := os.ReadDir("docs/architecture/adr")
	if err != nil {
		return nil, fmt.Errorf("bridge: read adr dir: %w", err)
	}
	var nodes []ContextNode
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := "docs/architecture/adr/" + e.Name()
		content, err := extractLines(path, 1, 80)
		if err != nil {
			continue
		}
		nodes = append(nodes, ContextNode{
			Name:        e.Name(),
			Type:        "adr",
			FilePath:    path,
			CodeSnippet: content,
		})
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
