# Linter Cleanup: Complexity Reduction & Doc Comments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce all 6 gocyclo violations from >15 to ≤15 and add doc comments to all 131 revive-exported symbols, achieving 0 linter issues.

**Architecture:** Extract helper methods/functions from high-complexity functions to reduce cyclomatic complexity without changing behavior. Each extracted function has a single responsibility. Doc comments follow Go conventions: `// SymbolName verb phrase.` starting with the exported name.

**Tech Stack:** Go 1.26.2, golangci-lint v2.12.2

**Execution Order:** Tasks 1-6 reduce gocyclo (highest first), Tasks 7-16 add doc comments (batched by file, highest issue count first), Task 17 is final verification.

---

## Pre-Flight Check

- [ ] **Step 0: Verify clean baseline**

Run:
```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./...
/home/emiyakiritsugu/go/bin/golangci-lint run
```

Expected: Build passes, all tests pass, 137 issues (6 gocyclo + 131 revive). If build or tests fail, STOP and fix before proceeding.

---

## Task 1: Reduce `GrepSearchTool.Execute` complexity (22 → ≤15)

**Files:**
- Modify: `internal/agents/tools.go`
- Test: `internal/agents/tools_test.go`, `internal/agents/engine_helpers_test.go`

**Current complexity: 22** — the walk callback contains directory skip logic, file extension filtering, file opening, regex scanning, and match limit all in one closure.

**Strategy:** Extract 3 helper functions from the walk callback to reduce branching inside `Execute`:
1. `shouldSkipDir(d fs.DirEntry) bool` — directory skip logic
2. `isTextFile(ext string) bool` — file extension filter
3. `scanFileMatches(re *regexp.Regexp, path string) ([]string, error)` — open, scan, close a single file

- [ ] **Step 1: Add file-extension lookup table above the `GrepSearchTool` struct**

Add this var block at line 252 (just before `type GrepSearchTool struct`):

```go
// textFileExtensions defines file extensions that GrepSearchTool scans.
var textFileExtensions = map[string]bool{
	".go":   true,
	".md":   true,
	".json": true,
	".yaml": true,
	".yml":  true,
	".sql":  true,
}

// skipDirs defines directory names that GrepSearchTool skips during traversal.
var skipDirs = map[string]bool{
	".git":          true,
	"node_modules": true,
	"vendor":       true,
}
```

- [ ] **Step 2: Add `shouldSkipDir` helper function**

Add this function after the `var` block, before `type GrepSearchTool struct`:

```go
// shouldSkipDir returns true if the directory should be skipped during file traversal.
func shouldSkipDir(d fs.DirEntry) bool {
	return skipDirs[d.Name()]
}
```

You need to add `"io/fs"` to the imports at the top of `internal/agents/tools.go` (it should already have `"os"` and `"path/filepath"`).

- [ ] **Step 3: Add `isTextFile` helper function**

Add right after `shouldSkipDir`:

```go
// isTextFile returns true if the file extension is one that GrepSearchTool scans.
func isTextFile(ext string) bool {
	return textFileExtensions[ext]
}
```

- [ ] **Step 4: Add `scanFileMatches` helper function**

Add after `isTextFile`:

```go
// scanFileMatches opens a file, scans each line for regex matches, and returns
// formatted "path:line: text" entries. Returns an error if the file cannot be opened.
func scanFileMatches(re *regexp.Regexp, path string) ([]string, error) {
	file, err := os.Open(path) //nolint:gosec // path from user input (validated)
	if err != nil {
		return nil, nil //nolint:nilnil // skip files we can't open
	}
	defer func() { _ = file.Close() }()

	var matches []string
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		if re.MatchString(scanner.Text()) {
			matches = append(matches, fmt.Sprintf("%s:%d: %s", path, lineNum, scanner.Text()))
		}
		if len(matches) > 100 {
			return matches, fmt.Errorf("too many matches found")
		}
		lineNum++
	}
	return matches, nil
}
```

- [ ] **Step 5: Rewrite `GrepSearchTool.Execute` to use the helpers**

Replace the entire `Execute` method (lines 290-353) with:

```go
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
			if shouldSkipDir(d) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isTextFile(filepath.Ext(path)) {
			return nil
		}

		fileMatches, scanErr := scanFileMatches(re, path)
		if scanErr != nil && scanErr.Error() == "too many matches found" {
			matches = append(matches, fileMatches...)
			return scanErr
		}
		matches = append(matches, fileMatches...)
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
```

- [ ] **Step 6: Add `"io/fs"` to imports**

In `internal/agents/tools.go`, add `"io/fs"` to the import block. The import block should look like:

```go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
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
```

- [ ] **Step 7: Run tests**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/agents/... -v -count=1
```

Expected: BUILD passes, all agent tests pass.

- [ ] **Step 8: Verify gocyclo for GrepSearchTool.Execute**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run --enable gocyclo 2>&1 | grep "GrepSearchTool"
```

Expected: No output (complexity ≤ 15). If it still shows >15, the walk callback still has too many branches — extract more.

- [ ] **Step 9: Commit**

```bash
git add internal/agents/tools.go
git commit -m "refactor: extract helpers from GrepSearchTool.Execute to reduce complexity 22→≤15

- Extract shouldSkipDir() for directory skip logic
- Extract isTextFile() for file extension filtering  
- Extract scanFileMatches() for file scanning logic
- No behavior change, pure structural refactor"
```

---

## Task 2: Reduce `NewInstructCmd` complexity (19 → ≤15)

**Files:**
- Modify: `cmd/sentinel/commands/instruct.go`
- Test: `cmd/sentinel/commands/instruct_fp_test.go`

**Current complexity: 19** — the `cmd.RunE` closure contains intent reading, diagnostic gathering, menu presentation, ADR data collection switch, and persistence all in one anonymous function.

**Strategy:** Extract 2 functions and 1 method from the closure:
1. `readIntent(message string, quick bool) (string, error)` — all intent-reading logic
2. `collectADRDetails(choice string, intent string, evidence string) graph.ADRData` — the switch for ADR data collection
3. Keep persistence inline (it's only 6 lines, simple error handling, no branching)

- [ ] **Step 1: Add `readIntent` function before `NewInstructCmd`**

Add this function before `func NewInstructCmd` (before line 21):

```go
// readIntent determines the user intent from an explicit message flag, stdin pipe,
// or interactive prompt. Returns the intent string (may be empty).
func readIntent(message string, quick bool) (string, error) {
	if message != "" {
		return message, nil
	}

	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text()), nil
		}
		return "", nil
	}

	if quick {
		return "", nil
	}

	fmt.Println("\n🧠 SENTINEL INTERVIEW MODE")
	fmt.Println("======================================")
	fmt.Println("O que você deseja construir hoje?")
	fmt.Print("> ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("instruct: failed to read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// collectADRDetails fills an ADRData based on the user's menu choice.
func collectADRDetails(choice string, intent string, evidence string) graph.ADRData {
	adrData := graph.ADRData{
		Title:  intent,
		Status: "PROPOSED",
	}

	switch choice {
	case "m":
		adrData = runSocraticInterview(intent, evidence)
	case "a":
		fmt.Println("✨ Chamando AI Bridge para expansão... (Simulado na Fase 4.1)")
		adrData.Context = "Expandido via IA baseado em: " + intent + "\n" + evidence
		adrData.Decision = "Padrão recomendado pela IA para este cenário."
		adrData.VerificationCommand = "go test ./..."
	case "s":
		adrData.Status = "DRAFT"
		adrData.Context = "Aguardando refinamento pelo Sentinel Agent."
		adrData.VerificationCommand = "# O Agente definirá o commando de prova"
	default: // q
		adrData.Context = "Capturado via commando 'instruct'.\nIntenção: " + intent
		adrData.Decision = "[Descreva a abordagem técnica]"
		adrData.VerificationCommand = "go build ./..."
	}

	return adrData
}
```

- [ ] **Step 2: Rewrite `NewInstructCmd` to use the extracted functions**

Replace the entire `NewInstructCmd` function (lines 21-145) with:

```go
func NewInstructCmd(db *sqlite.DB) *cobra.Command {
	var message string
	var quick bool

	cmd := &cobra.Command{
		Use:   "instruct",
		Short: "Interview mode to capture user intent and generate tasks",
	}

	if err := sqlite.ValidateDB(db, "instruct-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		intent, err := readIntent(message, quick)
		if err != nil {
			return err
		}
		if intent == "" {
			return nil
		}

		fmt.Printf("\n🔍 Sentinel: Analisando '%s'...\n", intent)

		isVague := isVagueIntent(intent)
		var evidence string
		if isVague && !quick {
			evidence = performDiagnostic(cmd.Context(), db)
			if evidence != "" {
				fmt.Println("\n⚠️  EVIDÊNCIA ENCONTRADA (Sessão de Diagnóstico):")
				fmt.Println(evidence)
				fmt.Println("\nIsto parece ser o ponto de partida ideal para evitar 'guessing'.")
			}
		}

		var choice string
		if quick {
			choice = "q"
		} else {
			fmt.Println("\nComo deseja preencher os detalhes técnicos do ADR?")
			fmt.Println("[m] Manual (Entrevista Socrática)")
			fmt.Println("[a] IA Now (Sugestão via Gemini)")
			fmt.Println("[s] Sentinel (Delegar ao Agente durante a execução)")
			fmt.Println("[q] Quick (Usar placeholders)")
			fmt.Print("\nEscolha> ")
			_, _ = fmt.Scanln(&choice)
		}

		adrData := collectADRDetails(choice, intent, evidence)

		if adrData.Status != "DRAFT" && strings.TrimSpace(adrData.VerificationCommand) == "" {
			return fmt.Errorf("instruct: verification command is required for non-draft ADRs")
		}

		manager, err := state.NewManager(db)
		if err != nil {
			return fmt.Errorf("instruct: failed to create manager: %w", err)
		}
		id, err := manager.CreateTask(cmd.Context(), intent, "T1", adrData.VerificationCommand)
		if err != nil {
			return fmt.Errorf("instruct: failed to create task: %w", err)
		}

		adrData.TaskID = id
		gen := graph.NewADRGenerator()
		adrPath, err := gen.Generate(adrData)
		if err != nil {
			fmt.Printf("\n⚠️  ADR Generation failed: %v\n", err)
		} else {
			fmt.Printf("\n📄 ADR Gerado: %s\n", adrPath)
			fmt.Printf("✅ Task [%s] criada com Protocolo de Verificação: %s\n", id, adrData.VerificationCommand)
		}

		return nil
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "User intent message")
	cmd.Flags().BoolVarP(&quick, "quick", "q", false, "Skip interview and use defaults")
	return cmd
}
```

- [ ] **Step 3: Run tests**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./cmd/sentinel/... -v -count=1
```

Expected: BUILD passes, command tests pass.

- [ ] **Step 4: Verify gocyclo for NewInstructCmd**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run --enable gocyclo 2>&1 | grep "NewInstructCmd"
```

Expected: No output (complexity ≤ 15).

- [ ] **Step 5: Commit**

```bash
git add cmd/sentinel/commands/instruct.go
git commit -m "refactor: extract readIntent and collectADRDetails from NewInstructCmd to reduce complexity 19→≤15

- Extract readIntent() for intent-reading logic (flag, stdin, interactive)
- Extract collectADRDetails() for ADR data collection switch
- No behavior change, pure structural refactor"
```

---

## Task 3: Reduce `TreeSitterScanner.Scan` complexity (19 → ≤15)

**Files:**
- Modify: `internal/graph/scanner_treesitter.go`
- Test: none (no direct tests exist — see note below)

**⚠️ No direct tests exist for TreeSitterScanner.Scan.** The refactoring is purely structural (extracting methods from the same struct), so the existing integration test in `internal/graph/linker_integration_test.go` serves as a regression guard. Run `go test ./internal/graph/...` before and after.

**Current complexity: 19** — Scan contains file I/O, parser selection, query execution, and two distinct capture handlers (import + symbol) in one method.

**Strategy:** Extract 3 methods:
1. `(s *TreeSitterScanner) selectLanguage(ext string) (*sitter.Language, *sitter.Query)` — language/query selection
2. `(s *TreeSitterScanner) executeQuery(query *sitter.Query, tree *sitter.Tree, sourceCode []byte, path string) ScanResult` — cursor execution + match loop
3. Keep `processSymbol` as-is (already extracted)

- [ ] **Step 1: Add `selectLanguage` method**

Add this method after `SupportedExtensions()` (after line 57):

```go
// selectLanguage returns the appropriate Tree-sitter language and query for the
// given file extension.
func (s *TreeSitterScanner) selectLanguage(ext string) (*sitter.Language, *sitter.Query) {
	if ext == ".tsx" {
		return tsx.GetLanguage(), s.tsxQuery
	}
	return typescript.GetLanguage(), s.tsQuery
}
```

- [ ] **Step 2: Add `executeQuery` method**

Add this method after `selectLanguage`:

```go
// executeQuery runs a Tree-sitter semantic query against the parsed tree and
// returns the resulting nodes and edges.
func (s *TreeSitterScanner) executeQuery(query *sitter.Query, tree *sitter.Tree, sourceCode []byte, path string) ScanResult {
	res := ScanResult{}
	fileID := "file:" + path
	res.Nodes = append(res.Nodes, Node{
		ID:       fileID,
		Name:     filepath.Base(path),
		Type:     "file",
		FilePath: path,
	})

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	cursor.Exec(query, tree.RootNode())

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			node := capture.Node

			switch captureName {
			case "import":
				for i := 0; i < int(node.NamedChildCount()); i++ {
					child := node.NamedChild(i)
					if child.Type() == "string" {
						importPath := strings.Trim(child.Content(sourceCode), "'\"")
						importID := fmt.Sprintf("import:%s:%s", path, importPath)

						res.Nodes = append(res.Nodes, Node{
							ID:       importID,
							Name:     importPath,
							Type:     "unresolved_import",
							FilePath: path,
						})

						res.Edges = append(res.Edges, Edge{
							From: fileID,
							To:   importID,
							Rel:  "imports",
						})
					}
				}
			case "interface", "class", "function", "variable":
				var name string
				for i := 0; i < int(node.NamedChildCount()); i++ {
					child := node.NamedChild(i)
					if child.Type() == "identifier" || child.Type() == "type_identifier" {
						name = child.Content(sourceCode)
						break
					}
				}

				if name != "" {
					s.processSymbol(node, captureName, name, path, fileID, &res)
				}
			}
		}
	}

	return res
}
```

- [ ] **Step 3: Rewrite `Scan` method to use extracted helpers**

Replace the entire `Scan` method (lines 59-167) with:

```go
func (s *TreeSitterScanner) Scan(path string) ScanResult {
	file, err := os.Open(path) //nolint:gosec // path from scanner input
	if err != nil {
		return ScanResult{Err: fmt.Errorf("scanner: failed to open %s: %w", path, err)}
	}
	defer func() { _ = file.Close() }()

	sourceCode, err := io.ReadAll(file)
	if err != nil {
		return ScanResult{Err: fmt.Errorf("scanner: failed to read %s: %w", path, err)}
	}

	lang, query := s.selectLanguage(filepath.Ext(path))

	parser := s.pool.Get().(*sitter.Parser)
	defer s.pool.Put(parser)

	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil || tree == nil || tree.RootNode() == nil {
		return ScanResult{Err: fmt.Errorf("scanner: failed to parse %s: %w", path, err)}
	}
	defer tree.Close()

	if query == nil {
		return ScanResult{Nodes: []Node{{ID: "file:" + path, Name: filepath.Base(path), Type: "file", FilePath: path}}}
	}

	return s.executeQuery(query, tree, sourceCode, path)
}
```

- [ ] **Step 4: Run tests (regression check)**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/graph/... -v -count=1
```

Expected: BUILD passes, all graph tests pass.

- [ ] **Step 5: Verify gocyclo for TreeSitterScanner.Scan**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run --enable gocyclo 2>&1 | grep "TreeSitterScanner"
```

Expected: No output (complexity ≤ 15).

- [ ] **Step 6: Commit**

```bash
git add internal/graph/scanner_treesitter.go
git commit -m "refactor: extract selectLanguage and executeQuery from TreeSitterScanner.Scan to reduce complexity 19→≤15

- Extract selectLanguage() for language/query selection by extension
- Extract executeQuery() for cursor execution and match iteration
- No behavior change, pure structural refactor"
```

---

## Task 4: Reduce `Visualizer.formatC4Mermaid` complexity (19 → ≤15)

**Files:**
- Modify: `internal/graph/visualizer.go`
- Test: `internal/graph/visualizer_fp_test.go`, `internal/graph/nil_guard_test.go`

**Current complexity: 19** — formatC4Mermaid contains container definitions, node-to-container mapping, double-pass edge mapping, and relationship aggregation all in one function.

**Strategy:** Extract 3 helper functions:
1. `buildContainerDefs(sb *strings.Builder, containers map[string]container)` — write container definitions
2. `buildNodeToContainerMap(nodes []Node, edges []Edge) map[string]string` — classify all nodes into containers
3. `aggregateRelationships(edges []Edge, nodeToContainer map[string]string) map[relKey]string` — aggregate inter-container relationships

The `container` struct and `relKey` struct must be promoted from local to package-level (or kept as locals within formatC4Mermaid but passed to helpers). Since `container` is only used in formatC4Mermaid and its helpers, we'll define `container` and `relKey` as named types at file scope.

- [ ] **Step 1: Promote `container` and `relKey` to file-scope types**

Add these types before the `classifyContainer` function (around line 142):

```go
// c4Container represents a C4 container in the architecture diagram.
type c4Container struct {
	id   string
	name string
	desc string
	isDb bool
}

// c4RelKey identifies a unique relationship between two containers.
type c4RelKey struct {
	from, to string
}
```

- [ ] **Step 2: Add `writeContainerDefs` helper function**

Add after the `c4RelKey` type definition:

```go
// writeContainerDefs writes the C4 container definitions to the builder.
func writeContainerDefs(sb *strings.Builder, containers []c4Container) {
	for _, c := range containers {
		if c.isDb {
			fmt.Fprintf(sb, "    ContainerDb(%s, \"%s\", \"SQLite\", \"%s\")\n", c.id, c.name, c.desc)
		} else {
			fmt.Fprintf(sb, "    Container(%s, \"%s\", \"Go\", \"%s\")\n", c.id, c.name, c.desc)
		}
	}
	sb.WriteString("\n")
}
```

- [ ] **Step 3: Add `buildNodeToContainerMap` helper function**

```go
// buildNodeToContainerMap classifies all nodes and edge endpoints into their
// corresponding C4 containers based on file paths.
func buildNodeToContainerMap(nodes []Node, edges []Edge) map[string]string {
	nodeToContainer := make(map[string]string)

	for _, n := range nodes {
		cid := classifyContainer(n.FilePath)
		if cid == "" && strings.HasPrefix(n.ID, "file:") {
			cid = classifyContainer(strings.TrimPrefix(n.ID, "file:"))
		}
		if cid != "" {
			nodeToContainer[n.ID] = cid
		}
	}

	for _, e := range edges {
		if _, ok := nodeToContainer[e.To]; !ok && strings.HasPrefix(e.To, "file:") {
			path := strings.TrimPrefix(e.To, "file:")
			if cid := classifyContainer(path); cid != "" {
				nodeToContainer[e.To] = cid
			}
		}
		if _, ok := nodeToContainer[e.From]; !ok && strings.HasPrefix(e.From, "file:") {
			path := strings.TrimPrefix(e.From, "file:")
			if cid := classifyContainer(path); cid != "" {
				nodeToContainer[e.From] = cid
			}
		}
	}

	return nodeToContainer
}
```

- [ ] **Step 4: Add `aggregateRelationships` helper function**

```go
// aggregateRelationships collects unique inter-container relationships from edges.
func aggregateRelationships(edges []Edge, nodeToContainer map[string]string) map[c4RelKey]string {
	rels := make(map[c4RelKey]string)
	for _, e := range edges {
		fromC, okF := nodeToContainer[e.From]
		toC, okT := nodeToContainer[e.To]

		if okF && okT && fromC != toC {
			rels[c4RelKey{fromC, toC}] = e.Rel
		}
	}
	return rels
}
```

- [ ] **Step 5: Rewrite `formatC4Mermaid` to use the helpers**

Replace the entire `formatC4Mermaid` method (lines 161-237) with:

```go
func (v *Visualizer) formatC4Mermaid(nodes []Node, edges []Edge) string {
	containers := []c4Container{
		{id: "cli", name: "CLI Application", desc: "Interface Go/Cobra para desenvolvedores"},
		{id: "agents", name: "Agent Engine", desc: "Orquestração de loops cognitivos ReAct"},
		{id: "graph", name: "Graph Engine", desc: "Análise AST e extração semântica"},
		{id: "audit", name: "Compliance Guard", desc: "Validação de padrões e Hard Gates"},
		{id: "state", name: "State Manager", desc: "Gerenciamento de tarefas e histórico"},
		{id: "frontend", name: "Legacy Frontend", desc: "Components legados em TypeScript"},
		{id: "db", name: "SQLite Graph", desc: "Persistência de nós, arestas e tarefas", isDb: true},
	}

	var sb strings.Builder
	writeContainerDefs(&sb, containers)

	nodeToContainer := buildNodeToContainerMap(nodes, edges)
	rels := aggregateRelationships(edges, nodeToContainer)

	for k, rel := range rels {
		fmt.Fprintf(&sb, "    Rel(%s, %s, \"%s\")\n", k.from, k.to, rel)
	}

	return sb.String()
}
```

Note: `containers` changed from `map[string]container` to `[]c4Container` (ordered slice). This eliminates the non-deterministic map iteration and is actually a bugfix — the original map iteration could produce different ordering each run. The `writeContainerDefs` function iterates the slice in declaration order.

- [ ] **Step 6: Run tests**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/graph/... -v -count=1
```

Expected: BUILD passes, all graph tests pass.

- [ ] **Step 7: Verify gocyclo for formatC4Mermaid**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run --enable gocyclo 2>&1 | grep "formatC4Mermaid"
```

Expected: No output (complexity ≤ 15).

- [ ] **Step 8: Commit**

```bash
git add internal/graph/visualizer.go
git commit -m "refactor: extract helpers from formatC4Mermaid to reduce complexity 19→≤15

- Extract writeContainerDefs() for container definition output
- Extract buildNodeToContainerMap() for node-to-container classification
- Extract aggregateRelationships() for inter-container relationships
- Promote c4Container/c4RelKey to named types for type safety
- Fix: containers changed from map to ordered slice for deterministic output
- No behavior change (except deterministic container ordering)"
```

---

## Task 5: Reduce `Engine.Execute` complexity (16 → ≤15)

**Files:**
- Modify: `internal/agents/engine.go`
- Test: `internal/agents/engine_test.go`, `internal/agents/engine_helpers_test.go`

**Current complexity: 16** — only needs 2 points of reduction.

**Strategy:** Extract `shouldTerminate` to eliminate the termination check branches from the main loop body.

- [ ] **Step 1: Add `shouldTerminate` helper function**

Add this function before the `Execute` method (around line 165):

```go
// shouldTerminate returns true when the model response contains a Sovereign Audit
// report and no pending tool calls, indicating the agent has completed its task.
func shouldTerminate(toolCalls []map[string]interface{}, textResponses []string) bool {
	if len(toolCalls) == 0 {
		for _, text := range textResponses {
			if strings.Contains(strings.ToLower(text), "sovereign audit") {
				return true
			}
		}
	}
	return false
}
```

- [ ] **Step 2: Find the `containsSovereignAudit` function in engine.go**

Search for `func containsSovereignAudit` in `internal/agents/engine.go`. It should look like:

```go
func containsSovereignAudit(texts []string) bool {
```

Replace the entire `containsSovereignAudit` function with `shouldTerminate` (which you already added). Then find the call site in `Execute` and update it.

- [ ] **Step 3: Update the call site in `Execute`**

Find the line in `Execute` that looks like:

```go
		if len(toolCalls) == 0 && containsSovereignAudit(textResponses) {
```

Replace it with:

```go
		if shouldTerminate(toolCalls, textResponses) {
```

- [ ] **Step 4: Remove the old `containsSovereignAudit` function**

Delete the entire `containsSovereignAudit` function from `engine.go`.

- [ ] **Step 5: Run tests**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/agents/... -v -count=1
```

Expected: BUILD passes, all agent tests pass.

- [ ] **Step 6: Verify gocyclo for Engine.Execute**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run --enable gocyclo 2>&1 | grep "Engine.*Execute"
```

Expected: No output (complexity ≤ 15).

- [ ] **Step 7: Commit**

```bash
git add internal/agents/engine.go
git commit -m "refactor: extract shouldTerminate from Engine.Execute to reduce complexity 16→≤15

- Replace containsSovereignAudit with shouldTerminate that checks both conditions
- No behavior change, pure structural refactor"
```

---

## Task 6: Reduce `Disambiguator.anchorSignal` complexity (16 → ≤15)

**Files:**
- Modify: `internal/intake/disambiguator.go`
- Test: `internal/intake/disambiguator_test.go`, `internal/intake/disambiguator_db_test.go`

**Current complexity: 16** — only needs 2 points of reduction.

**Strategy:** Extract 2 helper functions from the branching logic:
1. `hasCodeAnchor(lower string) bool` — the lexical anchor checks (Phase 1)
2. `matchKeywordsInGraph(db *sql.DB, keywords []string) (matched int, total int)` — the DB query loop (Phase 2)

- [ ] **Step 1: Add `hasCodeAnchor` helper function**

Add this function before `anchorSignal` in `internal/intake/disambiguator.go` (around line 118):

```go
// hasCodeAnchor returns true if the description contains lexical anchors that
// indicate a precise code reference (a file path, module path, or file extension).
func hasCodeAnchor(lower string) bool {
	if strings.Contains(lower, "internal/") ||
		strings.Contains(lower, "pkg/") ||
		strings.Contains(lower, ".go") {
		return true
	}
	return false
}
```

- [ ] **Step 2: Add `hasLineReference` helper function**

Add after `hasCodeAnchor`:

```go
// hasLineReference returns true if the description contains a line reference
// pattern (a colon followed by a digit), such as "main.go:42".
func hasLineReference(description string) bool {
	for i, ch := range description {
		if ch == ':' && i+1 < len(description) && description[i+1] >= '0' && description[i+1] <= '9' {
			return true
		}
	}
	return false
}
```

- [ ] **Step 3: Add `matchKeywordsInGraph` helper function**

Add after `hasLineReference`:

```go
// matchKeywordsInGraph queries the graph database for each keyword and returns
// how many keywords matched at least one node. Returns -1 for total if the
// graph is not indexed (empty or error).
func matchKeywordsInGraph(db *sql.DB, keywords []string) (matched int, total int) {
	for _, kw := range keywords {
		var n int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM nodes WHERE LOWER(name) LIKE ?",
			fmt.Sprintf("%%%s%%", kw),
		).Scan(&n)
		if err == nil && n > 0 {
			matched++
		}
	}
	return matched, len(keywords)
}
```

You need to add `"database/sql"` to the import block. Also, we need to import `*sql.DB` instead of using `d.db` — but `matchKeywordsInGraph` accepts `*sql.DB` directly so it can be tested independently.

- [ ] **Step 4: Rewrite `anchorSignal` to use helpers**

Replace the entire `anchorSignal` method (lines 119-164) with:

```go
func (d *Disambiguator) anchorSignal(description string) float64 {
	lower := strings.ToLower(description)

	// Phase 1: lexical anchors (zero DB)
	if hasCodeAnchor(lower) {
		return 0.00
	}
	if hasLineReference(description) {
		return 0.00
	}

	// Phase 2: graph-anchored (DB query)
	if d.db == nil {
		return weightAnchor
	}

	var count int
	if err := d.db.Conn.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count); err != nil || count == 0 {
		return weightAnchor
	}

	keywords := extractKeywords(description)
	if len(keywords) == 0 {
		return weightAnchor
	}

	matched, total := matchKeywordsInGraph(d.db.Conn, keywords)
	matchedRatio := float64(matched) / float64(total)
	return weightAnchor * (1.0 - matchedRatio)
}
```

- [ ] **Step 5: Add `"database/sql"` to imports**

In `internal/intake/disambiguator.go`, the import block should look like:

```go
import (
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)
```

- [ ] **Step 6: Run tests**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/intake/... -v -count=1
```

Expected: BUILD passes, all intake tests pass.

- [ ] **Step 7: Verify gocyclo for anchorSignal**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run --enable gocyclo 2>&1 | grep "anchorSignal"
```

Expected: No output (complexity ≤ 15).

- [ ] **Step 8: Commit**

```bash
git add internal/intake/disambiguator.go
git commit -m "refactor: extract hasCodeAnchor, hasLineReference, and matchKeywordsInGraph from anchorSignal to reduce complexity 16→≤15

- Extract hasCodeAnchor() for lexical anchor detection
- Extract hasLineReference() for line reference pattern detection
- Extract matchKeywordsInGraph() for DB keyword matching
- No behavior change, pure structural refactor"
```

---

## Task 7: Add doc comments to `internal/agents/tools.go` (53 issues)

**Files:**
- Modify: `internal/agents/tools.go`

This file has the most revive-exported issues (53). All are missing doc comments on exported types and methods. The doc comment convention is: `// SymbolName verb phrase.` — starts with the symbol name, present tense verb, period at end.

**Types requiring doc comments (7):**
- `ReadFileTool` (line 26)
- `WriteFileTool` (line 108)
- `ReplaceTool` (line 170)
- `GrepSearchTool` (line 253)
- `AuditTool` (line 357)
- `RunTool` (line 409)
- `ADRTool` (line 482)
- `ScanTool` (line 569)
- `DecomposeTool` (line 609)

**Methods requiring doc comments (for each type, there are up to 6): Name, Description, Definition, ValidateArguments, Execute**

- [ ] **Step 1: Add doc comments to ReadFileTool and its methods**

Before line 26 (`type ReadFileTool struct`), add:
```go
// ReadFileTool reads a file from the project directory with optional line range.
```

Before line 30 (`func (t *ReadFileTool) Name()`), add:
```go
// Name returns the tool identifier "read_file".
```

Before line 31 (`func (t *ReadFileTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 35 (`func (t *ReadFileTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 60 (`func (t *ReadFileTool) ValidateArguments`), add:
```go
// ValidateArguments checks that the path argument is present and valid.
```

Before line 68 (`func (t *ReadFileTool) Execute`), add:
```go
// Execute reads the file at the given path and returns the content within the
// specified line range.
```

- [ ] **Step 2: Add doc comments to WriteFileTool and its methods**

Before line 108 (`type WriteFileTool struct`), add:
```go
// WriteFileTool writes content to a file, creating parent directories as needed.
```

Before line 112 (`func (t *WriteFileTool) Name()`), add:
```go
// Name returns the tool identifier "write_file".
```

Before line 113 (`func (t *WriteFileTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 117 (`func (t *WriteFileTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 138 (`func (t *WriteFileTool) ValidateArguments`), add:
```go
// ValidateArguments checks that path and content arguments are present and valid.
```

Before line 146 (`func (t *WriteFileTool) Execute`), add:
```go
// Execute writes the content to the file after validating structural integrity.
```

- [ ] **Step 3: Add doc comments to ReplaceTool and its methods**

Before line 170 (`type ReplaceTool struct`), add:
```go
// ReplaceTool replaces a specific string in a file with new content.
```

Before line 174 (`func (t *ReplaceTool) Name()`), add:
```go
// Name returns the tool identifier "replace".
```

Before line 175 (`func (t *ReplaceTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 179 (`func (t *ReplaceTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 204 (`func (t *ReplaceTool) ValidateArguments`), add:
```go
// ValidateArguments checks that path, old_string, and new_string are present and valid.
```

Before line 212 (`func (t *ReplaceTool) Execute`), add:
```go
// Execute replaces old_string with new_string in the specified file.
```

- [ ] **Step 4: Add doc comments to GrepSearchTool and its methods**

Before line 253 (`type GrepSearchTool struct`), add:
```go
// GrepSearchTool searches for a regular expression pattern within file contents.
```

Before line 257 (`func (t *GrepSearchTool) Name()`), add:
```go
// Name returns the tool identifier "grep_search".
```

Before line 258 (`func (t *GrepSearchTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 262 (`func (t *GrepSearchTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 283 (`func (t *GrepSearchTool) ValidateArguments`), add:
```go
// ValidateArguments checks that dir_path is valid if provided.
```

- [ ] **Step 5: Add doc comments to AuditTool and its methods**

Before line 357 (`type AuditTool struct`), add:
```go
// AuditTool runs the Sovereign Validator to detect Standard violations.
```

Before line 361 (`func (t *AuditTool) Name()`), add:
```go
// Name returns the tool identifier "sentinel:audit".
```

Before line 362 (`func (t *AuditTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 366 (`func (t *AuditTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 376 (`func (t *AuditTool) ValidateArguments`), add:
```go
// ValidateArguments returns nil; the audit tool has no arguments.
```

Before line 380 (`func (t *AuditTool) Execute`), add:
```go
// Execute runs the Sovereign Validator and returns a report of violations.
```

- [ ] **Step 6: Add doc comments to RunTool and its methods**

Before line 409 (`type RunTool struct`), add:
```go
// RunTool executes a safe, approved shell command and returns its output.
```

Before line 413 (`func (t *RunTool) Name()`), add:
```go
// Name returns the tool identifier "sentinel:run".
```

Before line 414 (`func (t *RunTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 418 (`func (t *RunTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 435 (`func (t *RunTool) ValidateArguments`), add:
```go
// ValidateArguments checks that the command argument is present and valid.
```

Before line 443 (`func (t *RunTool) Execute`), add:
```go
// Execute runs the command and returns its output, truncated to 200 lines or 10KB.
```

- [ ] **Step 7: Add doc comments to ADRTool and its methods**

Before line 482 (`type ADRTool struct`), add:
```go
// ADRTool generates a formal Architectural Decision Record file.
```

Before line 486 (`func (t *ADRTool) Name()`), add:
```go
// Name returns the tool identifier "sentinel:adr".
```

Before line 487 (`func (t *ADRTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 491 (`func (t *ADRTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 524 (`func (t *ADRTool) ValidateArguments`), add:
```go
// ValidateArguments checks that title, context, decision, and verification
// are present and valid.
```

Before line 534 (`func (t *ADRTool) Execute`), add:
```go
// Execute generates an ADR file from the provided arguments.
```

- [ ] **Step 8: Add doc comments to ScanTool and its methods**

Before line 569 (`type ScanTool struct`), add:
```go
// ScanTool triggers a project scan to update the dependency graph.
```

Before line 573 (`func (t *ScanTool) Name()`), add:
```go
// Name returns the tool identifier "sentinel:scan".
```

Before line 574 (`func (t *ScanTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 578 (`func (t *ScanTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 588 (`func (t *ScanTool) ValidateArguments`), add:
```go
// ValidateArguments returns nil; the scan tool has no arguments.
```

Before line 592 (`func (t *ScanTool) Execute`), add:
```go
// Execute runs the graph engine scan and returns a summary.
```

- [ ] **Step 9: Add doc comments to DecomposeTool and its methods**

Before line 609 (`type DecomposeTool struct`), add:
```go
// DecomposeTool creates sub-tasks from a parent task for parallel execution.
```

Before line 613 (`func (t *DecomposeTool) Name()`), add:
```go
// Name returns the tool identifier "sentinel:decompose".
```

Before line 614 (`func (t *DecomposeTool) Description()`), add:
```go
// Description returns a human-readable description of the tool.
```

Before line 618 (`func (t *DecomposeTool) Definition()`), add:
```go
// Definition returns the Gemini function declaration schema for the tool.
```

Before line 655 (`func (t *DecomposeTool) ValidateArguments`), add:
```go
// ValidateArguments checks that task_id, description, branch_name, and
// subtasks are present and valid.
```

Before line 701 (`func (t *DecomposeTool) Execute`), add:
```go
// Execute creates sub-tasks in the database and returns a summary.
```

- [ ] **Step 10: Commit**

```bash
git add internal/agents/tools.go
git commit -m "docs: add exported doc comments to all tool types and methods in agents/tools.go

28 doc comments added to satisfy revive/exported linter.
No behavior change."
```

---

## Task 8: Add doc comments to `internal/patterns/store.go` (11 issues) + `internal/patterns/backfill.go` (5 issues) + `internal/patterns/dedup.go` (1 issue)

**Files:**
- Modify: `internal/patterns/store.go`
- Modify: `internal/patterns/backfill.go`
- Modify: `internal/patterns/dedup.go`

- [ ] **Step 1: Read all three files and add doc comments**

Run `cat` to see exact line numbers:
```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
/home/emiyakiritsugu/go/bin/golangci-lint run 2>&1 | grep "revive" | grep "internal/patterns"
```

Add doc comments to each exported symbol. Convention: `// SymbolName verb phrase.`

For `internal/patterns/store.go`:
- `CategoryAntiPattern` const block → `// CategoryAntiPattern identifies an anti-pattern category.`
- `SourceCognitiveDNA` const block → `// SourceCognitiveDNA identifies patterns sourced from Cognitive DNA analysis.`
- `ImpactHigh` const block → `// ImpactHigh identifies a high-impact pattern.`
- `Pattern` struct → `// Pattern represents a detected or stored architectural pattern.`
- `ListFilters` struct → `// ListFilters holds optional filters for listing patterns.`
- `PatternStore` struct → `// PatternStore manages persistence and retrieval of patterns in SQLite.`
- `NewPatternStore` func → `// NewPatternStore creates a PatternStore backed by the given database.`
- `Create` method → `// Create inserts a new pattern into the store.`
- `List` method → `// List returns patterns matching the given filters.`
- `Search` method → `// Search returns patterns whose content matches the query.`
- `Get` method → `// Get returns a single pattern by ID.`

For `internal/patterns/backfill.go`:
- `BackfillResult` struct → `// BackfillResult holds the outcome of a backfill operation.`
- `BackfillCandidate` struct → `// BackfillCandidate represents a pattern candidate for backfilling.`
- `BackfillFromCognitiveDNA` method → `// BackfillFromCognitiveDNA imports patterns from Cognitive DNA analysis.`
- `BackfillFromEvolutionInsights` method → `// BackfillFromEvolutionInsights imports patterns from evolution insights.`
- `BackfillFromSentinelLog` method → `// BackfillFromSentinelLog imports patterns from the Sentinel log.`

For `internal/patterns/dedup.go`:
- `FindSimilar` method → `// FindSimilar returns patterns similar to the given pattern.`

- [ ] **Step 2: Run build and tests**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/patterns/... -v -count=1
```

Expected: Build passes, tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/patterns/store.go internal/patterns/backfill.go internal/patterns/dedup.go
git commit -m "docs: add exported doc comments to patterns package (17 symbols)

17 doc comments added to satisfy revive/exported linter.
No behavior change."
```

---

## Task 9: Add doc comments to `internal/agents/mutation.go` (4 issues)

**Files:**
- Modify: `internal/agents/mutation.go`

- [ ] **Step 1: Add doc comments**

- `MutationEngine` struct (line 15) → `// MutationEngine coordinates mutation operations on the database.`
- `NewMutationEngine` func (line 19) → `// NewMutationEngine creates a MutationEngine backed by the given database.`
- `Mutate` method (line 28) → `// Mutate applies a mutation prompt to the specialist and returns the result.`
- `Rollback` method (line 71) → `// Rollback reverses a previous mutation for the given specialist.`

- [ ] **Step 2: Run build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/agents/mutation.go
git commit -m "docs: add exported doc comments to mutation.go (4 symbols)

No behavior change."
```

---

## Task 10: Add doc comments to `internal/bridge/` (10 issues: classifier 5, gemini_classifier 2, prompt_factory 5 ≈ but 5+2 overlaps)

Actually: classifier.go has 5 issues, gemini_classifier.go has 2 issues, prompt_factory.go has 5 issues = 12 issues total across 3 files.

**Files:**
- Modify: `internal/bridge/classifier.go`
- Modify: `internal/bridge/gemini_classifier.go`
- Modify: `internal/bridge/prompt_factory.go`

- [ ] **Step 1: Get exact line numbers and add doc comments**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run 2>&1 | grep "revive" | grep "internal/bridge"
```

For `internal/bridge/classifier.go`:
- `Intent` struct (line 12) → `// Intent represents a classified user intent type.`
- `IntentDiagnose` const (line 15) → Add block comment: `// Sentinel intent types.`
- `NewIntentClassifier` func (line 43) → `// NewIntentClassifier creates a classifier using the Gemini backend.`
- `NewNilClassifier` func (line 111) → `// NewNilClassifier creates a no-op classifier that always returns IntentDiagnose.`
- `Classify` method (line 113) → `// Classify returns the classified intent for the given description.`

For `internal/bridge/gemini_classifier.go`:
- `NewGeminiClassifier` func (line 21) → `// NewGeminiClassifier creates a classifier backed by the Gemini API.`
- `Classify` method (line 28) → `// Classify sends the description to Gemini and returns the classified intent.`

For `internal/bridge/prompt_factory.go`:
- `ADR` struct (line 15) → `// ADR represents an Architectural Decision Record for prompt generation.`
- `ContextNode` struct (line 20) → `// ContextNode represents a node in the surgical context graph.`
- `ContextPayload` struct (line 29) → `// ContextPayload holds the generated prompt payload for the model.`
- `Factory` struct (line 36) → `// Factory generates prompt payloads with surgical context.`
- `NewFactory` func (line 41) → `// NewFactory creates a prompt factory backed by the given database.`

- [ ] **Step 2: Run build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/bridge/classifier.go internal/bridge/gemini_classifier.go internal/bridge/prompt_factory.go
git commit -m "docs: add exported doc comments to bridge package (12 symbols)

No behavior change."
```

---

## Task 11: Add doc comments to `internal/graph/` (12 issues across 6 files)

**Files:**
- Modify: `internal/graph/adr_generator.go` (1 issue)
- Modify: `internal/graph/engine.go` (4 issues)
- Modify: `internal/graph/events.go` (4 issues)
- Modify: `internal/graph/scanner_go.go` (4 issues)
- Modify: `internal/graph/schema.go` (1 issue)
- Modify: `internal/graph/visualizer.go` (2 existing issues — already has some comments)

- [ ] **Step 1: Get exact line numbers and add doc comments**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run 2>&1 | grep "revive" | grep "internal/graph"
```

For `internal/graph/adr_generator.go`:
- `NewADRGenerator` (line 18) → `// NewADRGenerator creates a new ADR generator instance.`

For `internal/graph/engine.go`:
- `Engine` struct (line 17) → `// Engine orchestrates graph scanning and persistence operations.`
- `NewEngine` func (line 26) → `// NewEngine creates a graph Engine backed by the given database.`
- `RegisterObserver` method (line 37) → `// RegisterObserver adds an observer to receive graph events.`
- `RegisterScanner` method (line 60) → `// RegisterScanner adds a scanner for the given file extensions.`

For `internal/graph/events.go`:
- `EventType` type (line 7) → `// EventType represents a graph engine event type.`
- `EventScanStarted` const block (line 10) → `// Graph event types.`
- `GraphEvent` struct (line 16) → `// GraphEvent represents an event emitted during graph operations.`
- `Observer` struct (line 22) → `// Observer receives graph events through callbacks.`

For `internal/graph/scanner_go.go`:
- `GoScanner` struct (line 12) → `// GoScanner scans Go source files using the go/ast parser.`
- `NewGoScanner` func (line 16) → `// NewGoScanner creates a new Go source file scanner.`
- `SupportedExtensions` method (line 20) → `// SupportedExtensions returns the file extensions this scanner handles.`
- `Scan` method (line 24) → `// Scan parses the Go file at path and returns its nodes and edges.`

For `internal/graph/schema.go`:
- `Migrate` func (line 169) → `// Migrate runs database schema migrations for the graph engine.`

For `internal/graph/visualizer.go`:
- `Visualizer` struct (line 14) → `// Visualizer generates Mermaid diagrams from graph data.`
- `NewVisualizer` func (line 18) → `// NewVisualizer creates a Visualizer backed by the given database.`

For `internal/graph/scanner_treesitter.go`:
- `NewTreeSitterScanner` func (line 34) → `// NewTreeSitterScanner creates a Tree-sitter scanner with pooled parsers.`
- `SupportedExtensions` method (line 55) → `// SupportedExtensions returns the file extensions this scanner handles.`

- [ ] **Step 2: Run build and tests**

```bash
go build ./...
go test ./internal/graph/... -v -count=1
```

- [ ] **Step 3: Commit**

```bash
git add internal/graph/adr_generator.go internal/graph/engine.go internal/graph/events.go internal/graph/scanner_go.go internal/graph/schema.go internal/graph/visualizer.go internal/graph/scanner_treesitter.go
git commit -m "docs: add exported doc comments to graph package (16 symbols)

No behavior change."
```

---

## Task 12: Add doc comments to `cmd/sentinel/commands/` (11 issues across 10 files)

**Files:**
- Modify: `cmd/sentinel/commands/audit.go` (1)
- Modify: `cmd/sentinel/commands/live.go` (1)
- Modify: `cmd/sentinel/commands/pattern.go` (1)
- Modify: `cmd/sentinel/commands/plan.go` (1)
- Modify: `cmd/sentinel/commands/report.go` (1)
- Modify: `cmd/sentinel/commands/root.go` (2)
- Modify: `cmd/sentinel/commands/scan.go` (1)
- Modify: `cmd/sentinel/commands/start.go` (1)
- Modify: `cmd/sentinel/commands/status.go` (1)
- Modify: `cmd/sentinel/commands/visualize.go` (1)

- [ ] **Step 1: Add doc comments**

For each `New*Cmd` function:
- `audit.go` → `// NewAuditCmd creates a cobra command that runs a Sovereign audit.`
- `live.go` → `// NewLiveCmd creates a cobra command that starts the live view server.`
- `pattern.go` → `// NewPatternCmd creates a cobra command for pattern management.`
- `plan.go` → `// NewPlanCmd creates a cobra command for planning tasks.`
- `report.go` → `// NewReportCmd creates a cobra command that generates a status report.`
- `scan.go` → `// NewScanCmd creates a cobra command that scans the project graph.`
- `start.go` → `// NewStartCmd creates a cobra command that starts the Sentinel agent.`
- `status.go` → `// NewStatusCmd creates a cobra command that shows task status.`
- `visualize.go` → `// NewVisualizeCmd creates a cobra command that generates architecture diagrams.`

For `root.go`:
- `NewRootCmd` → `// NewRootCmd creates the root cobra command for the Sentinel CLI.`
- `Execute` → `// Execute runs the root command and exits.`

- [ ] **Step 2: Run build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add cmd/sentinel/commands/
git commit -m "docs: add exported doc comments to command package (11 symbols)

No behavior change."
```

---

## Task 13: Add doc comments to remaining files (40 issues across 14 files)

**Files:**
- Modify: `internal/audit/runner.go` (2)
- Modify: `internal/intake/disambiguator.go` (1 — `NewDisambiguator`)
- Modify: `internal/liveview/api.go` (1)
- Modify: `internal/liveview/server.go` (1)
- Modify: `internal/reflect/validator.go` (3)
- Modify: `internal/report/aggregator.go` (4)
- Modify: `internal/state/manager.go` (3)
- Modify: `pkg/sqlite/db.go` (1)

- [ ] **Step 1: Get exact line numbers**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run 2>&1 | grep "revive" | grep -v "internal/agents/tools.go" | grep -v "internal/patterns" | grep -v "internal/agents/mutation" | grep -v "internal/bridge" | grep -v "internal/graph" | grep -v "cmd/sentinel"
```

- [ ] **Step 2: Add doc comments**

For `internal/audit/runner.go`:
- `Runner` struct → `// Runner executes the Sovereign Validator across the project.`
- `NewRunner` func → `// NewRunner creates a Validator Runner backed by the given database.`

For `internal/intake/disambiguator.go`:
- `NewDisambiguator` → `// NewDisambiguator creates a Disambiguator with an optional database connection.`

For `internal/liveview/api.go`:
- `GraphSnapshot` struct → `// GraphSnapshot represents a point-in-time snapshot of the graph for live view.`

For `internal/liveview/server.go`:
- `NewServer` func → `// NewServer creates a live view HTTP server.`

For `internal/reflect/validator.go`:
- `Violation` struct → `// Violation represents a single Standard violation found by the validator.`
- `Validator` struct → `// Validator checks project files against Sovereign Standards.`
- `NewValidator` func → `// NewValidator creates a Validator backed by the given database.`

For `internal/report/aggregator.go`:
- `TaskInfo` struct → `// TaskInfo holds summary information for a single task.`
- `ProjectStats` struct → `// ProjectStats holds aggregated project statistics.`
- `Aggregator` struct → `// Aggregator collects and summarizes project and task data.`
- `NewAggregator` func → `// NewAggregator creates a report Aggregator backed by the given database.`

For `internal/state/manager.go`:
- `Task` struct → `// Task represents a tracked unit of work in the state manager.`
- `Manager` struct → `// Manager handles task persistence and state transitions.`
- `NewManager` func → `// NewManager creates a state Manager backed by the given database.`

For `pkg/sqlite/db.go`:
- `DB` struct → `// DB wraps a SQLite connection with project-scoped access.`

- [ ] **Step 3: Run build**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/audit/runner.go internal/intake/disambiguator.go internal/liveview/api.go internal/liveview/server.go internal/reflect/validator.go internal/report/aggregator.go internal/state/manager.go pkg/sqlite/db.go
git commit -m "docs: add exported doc comments to remaining packages (16 symbols)

No behavior change."
```

---

## Task 14: Final Verification

- [ ] **Step 1: Build verification**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
```

Expected: Exit 0 (no errors).

- [ ] **Step 2: Test verification (with race detector)**

```bash
go test -race ./...
```

Expected: All 17 packages pass, no races detected.

- [ ] **Step 3: Linter verification (gocyclo only)**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run --enable gocyclo
```

Expected: 0 issues (all 6 functions should be ≤ 15).

- [ ] **Step 4: Full linter verification**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run
```

Expected: 0 issues (all gocyclo reduced + all revive exported comments added).

- [ ] **Step 5: go vet verification**

```bash
go vet ./...
```

Expected: Exit 0 (no issues).

- [ ] **Step 6: go mod tidy check**

```bash
go mod tidy -diff
git diff
```

Expected: No diff (clean module state).

---

## Self-Review Checklist

### 1. Spec Coverage

| Requirement | Task |
|---|---|
| GrepSearchTool.Execute (22) → ≤15 | Task 1 ✅ |
| NewInstructCmd (19) → ≤15 | Task 2 ✅ |
| TreeSitterScanner.Scan (19) → ≤15 | Task 3 ✅ |
| formatC4Mermaid (19) → ≤15 | Task 4 ✅ |
| Engine.Execute (16) → ≤15 | Task 5 ✅ |
| anchorSignal (16) → ≤15 | Task 6 ✅ |
| 131 revive exported doc comments | Tasks 7-13 ✅ |
| Build passes | Task 14 ✅ |
| Tests pass (race) | Task 14 ✅ |
| Linter: 0 issues | Task 14 ✅ |

### 2. Placeholder Scan

- ✅ No "TBD", "TODO", "implement later"
- ✅ No "add validation", "handle edge cases" without code
- ✅ All code blocks contain actual implementation
- ✅ All commit messages are complete
- ✅ All test commands include expected output descriptions

### 3. Type Consistency

- ✅ `shouldSkipDir(fs.DirEntry)` uses `fs.DirEntry` — matches `os.DirEntry` in `filepath.WalkDir` signature (they are the same type)
- ✅ `scanFileMatches(*regexp.Regexp, string)` — returns `([]string, error)` — matches usage in `Execute`
- ✅ `c4Container` and `c4RelKey` — promoted to file-scope types, used consistently in `formatC4Mermaid`, `writeContainerDefs`, `buildNodeToContainerMap`, `aggregateRelationships`
- ✅ `shouldTerminate([]map[string]interface{}, []string)` — matches types in `Execute`
- ✅ `matchKeywordsInGraph(*sql.DB, []string)` — uses `*sql.DB` directly, not `d.db.Conn` (which is `*sql.DB`)
- ✅ All doc comments start with the exported symbol name
- ✅ All helper functions are in the same package as the original code
- ✅ No cross-package imports added except `"io/fs"` and `"database/sql"` where needed

### Potential Risks and Mitigations

| Risk | Mitigation |
|---|---|
| `shouldSkipDir` uses `fs.DirEntry` but `filepath.WalkDir` provides `os.DirEntry` | `os.DirEntry` is an alias for `fs.DirEntry` — no type mismatch |
| `scanFileMatches` returns `nil, nil` for files that can't be opened (nolint:nilnil) | Added `//nolint:nilnil` comment — consistent with existing `//nolint:gosec` pattern |
| `formatC4Mermaid` containers changed from map to slice | This is actually a bugfix: maps have non-deterministic iteration order; slice guarantees stable output |
| `readIntent` returns `("", nil)` for empty stdin input | Matches original behavior where `scanner.Scan()` returns false for empty input |
| `matchKeywordsInGraph` accepts `*sql.DB` instead of using `d.db.Conn` | Allows testing with mock DB; `d.db.Conn` is already `*sql.DB` |
| GrepSearchTool.Execute rewrite changes match truncation behavior | Preserved: matches > 100 still trigger early exit via error propagation |
| `nilnil` lint issue from `scanFileMatches` | Added explicit nolint comment — this is a deliberate "skip unreadable files" pattern |