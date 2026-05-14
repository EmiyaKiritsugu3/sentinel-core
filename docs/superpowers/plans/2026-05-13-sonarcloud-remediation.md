# SonarCloud Remediation: English Comments + Code Quality

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all PR-introduced SonarCloud issues: translate ~99 Portuguese comments to English, reduce cognitive complexity in 2 refactored functions, fix duplicate literal detection, remove accidentally committed files.

**Architecture:** Batch file translations by package, refactor 2 functions with complexity > 15 back down to ≤ 15, fix literal duplication with constants.

**Tech Stack:** Go 1.26.2, golangci-lint v2.12.2, SonarCloud (S3776 cognitive complexity, S1192 literals, S8209 params)

---

## Pre-Flight Check

- [ ] **Step 0: Verify clean baseline**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./...
/home/emiyakiritsugu/go/bin/golangci-lint run
```

Expected: Build passes, all tests pass, 0 golangci-lint issues. If build or tests fail, STOP and fix before proceeding.

---

## Task 1: Remove accidentally committed files

**Files:**
- Delete: `install.sh` (third-party golangci-lint downloader, not project code)
- Delete: `coverage.out` (build artifact)
- Modify: `.gitignore` — add entries to prevent re-committing

- [ ] **Step 1: Remove install.sh**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
git rm install.sh
```

- [ ] **Step 2: Remove coverage.out**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
git rm coverage.out
```

- [ ] **Step 3: Update .gitignore (idempotent — skip if already present)**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
grep -qx "coverage.out" .gitignore 2>/dev/null || echo "coverage.out" >> .gitignore
grep -qx "install.sh" .gitignore 2>/dev/null || echo "install.sh" >> .gitignore
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove accidentally committed files (install.sh, coverage.out)"
```

---

## Task 2: Fix duplicate literal "too many matches found" (go:S1192)

**Files:**
- Modify: `internal/agents/tools.go`

**Issue:** Literal `"too many matches found"` appears 3 times in `scanFileMatches` and `GrepSearchTool.Execute`.

- [ ] **Step 1: Add a constant**

Replace the `scanFileMatches` function's literal and the `Execute` method's references.

In `internal/agents/tools.go`, add this constant above `textFileExtensions`:

```go
// errTooManyMatches is the sentinel error message for the match limit.
const errTooManyMatches = "too many matches found"
```

- [ ] **Step 2: Update scanFileMatches**

Replace the error return at line ~319:
```go
return matches, fmt.Errorf("too many matches found")
```
With:
```go
return matches, fmt.Errorf(errTooManyMatches)
```

- [ ] **Step 3: Update GrepSearchTool.Execute references**

Replace both occurrences (lines ~401 and ~409):
- `scanErr.Error() == "too many matches found"` → `scanErr.Error() == errTooManyMatches`
- `err.Error() != "too many matches found"` → `err.Error() != errTooManyMatches`

- [ ] **Step 4: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/agents/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/agents/tools.go
git commit -m "fix: extract 'too many matches found' literal to errTooManyMatches constant (S1192)"
```

---

## Task 3: Group same-type parameters in collectADRDetails (godre:S8209)

**Files:**
- Modify: `cmd/sentinel/commands/instruct.go`

**Issue:** `func collectADRDetails(choice string, intent string, evidence string)` — three consecutive `string` parameters.

- [ ] **Step 1: Group parameters**

Replace line 54:
```go
func collectADRDetails(choice string, intent string, evidence string) graph.ADRData {
```
With:
```go
func collectADRDetails(choice, intent, evidence string) graph.ADRData {
```

- [ ] **Step 2: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./cmd/sentinel/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 3: Commit**

```bash
git add cmd/sentinel/commands/instruct.go
git commit -m "fix: group consecutive string parameters in collectADRDetails (S8209)"
```

---

## Task 4: Reduce executeQuery complexity (31 → ≤15)

**Files:**
- Modify: `internal/graph/scanner_treesitter.go`

**Issue:** `executeQuery` method has cognitive complexity 31. The method was extracted from `Scan` but the extraction moved ALL the complexity into one method. Need to further decompose.

**Strategy:** Extract the import capture handler and the symbol capture handler into separate methods.

- [ ] **Step 1: Extract `handleImportCapture` method**

Add this method before `executeQuery`:

```go
// handleImportCapture extracts import paths from an import statement node
// and adds them to the scan result.
func (s *TreeSitterScanner) handleImportCapture(node *sitter.Node, sourceCode []byte, path, fileID string, res *ScanResult) {
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
}
```

- [ ] **Step 2: Extract `handleSymbolCapture` method**

Add after `handleImportCapture`:

```go
// handleSymbolCapture extracts symbol names from a declaration node
// and dispatches to processSymbol for type classification.
func (s *TreeSitterScanner) handleSymbolCapture(node *sitter.Node, captureName string, sourceCode []byte, path, fileID string, res *ScanResult) {
	var name string
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "identifier" || child.Type() == "type_identifier" {
			name = child.Content(sourceCode)
			break
		}
	}
	if name != "" {
		s.processSymbol(node, captureName, name, path, fileID, res)
	}
}
```

- [ ] **Step 3: Rewrite executeQuery to use helpers**

Replace the switch body inside `executeQuery` (the `for _, capture := range match.Captures` block):

```go
		for _, capture := range match.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			node := capture.Node

			switch captureName {
			case "import":
				s.handleImportCapture(node, sourceCode, path, fileID, &res)
			case "interface", "class", "function", "variable":
				s.handleSymbolCapture(node, captureName, sourceCode, path, fileID, &res)
			}
		}
```

- [ ] **Step 4: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/graph/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/graph/scanner_treesitter.go
git commit -m "refactor: extract handleImportCapture and handleSymbolCapture from executeQuery (S3776)"
```

---

---

## Task 5: Reduce buildNodeToContainerMap complexity (19 → ≤15)

**Files:**
- Modify: `internal/graph/visualizer.go`

**Issue:** `buildNodeToContainerMap` has cognitive complexity 19. Extracted from `formatC4Mermaid` but the edge-to-container classification loop carries high nesting.

**Strategy:** Extract the edge endpoint classification into a separate helper function.

- [ ] **Step 1: Extract `classifyEdgeEndpoints` helper**

Add this function before `buildNodeToContainerMap` in visualizer.go:

```go
// classifyEdgeEndpoints maps edge From/To IDs to containers when the ID
// has a "file:" prefix. Only maps endpoints not already in the container map.
func classifyEdgeEndpoints(edges []Edge, nodeToContainer map[string]string) {
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
}
```

- [ ] **Step 2: Rewrite buildNodeToContainerMap to use helper**

Replace the entire edge loop in `buildNodeToContainerMap` — find this block:

```go
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
```

Replace with:

```go
	classifyEdgeEndpoints(edges, nodeToContainer)
```

- [ ] **Step 3: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/graph/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/graph/visualizer.go
git commit -m "refactor: extract classifyEdgeEndpoints from buildNodeToContainerMap (S3776)"
```

---

## Task 6: Translate Portuguese comments in `pkg/` (18 comments across 4 files)

**Files:**
- Modify: `pkg/utils/filter.go`
- Modify: `pkg/utils/hash.go`
- Modify: `pkg/utils/text.go`
- Modify: `pkg/sqlite/db.go`

- [ ] **Step 1: Translate `pkg/utils/filter.go`**

| Line | Portuguese | English |
|---|---|---|
| 11 | `// IgnoreFilter gerencia as regras de exclusão baseadas no .gitignore` | `// IgnoreFilter manages exclusion rules based on .gitignore` |
| 16 | `// NewIgnoreFilter carrega os padrões de um diretório raiz` | `// NewIgnoreFilter loads patterns from a root directory` |
| 55 | `// 2. Padrões do .gitignore` | `// 2. .gitignore patterns` |
| 59 | `// Match exato de arquivo ou pasta` | `// Exact file or folder match` |
| 64 | `// Match de diretório (prefixo ou conteúdo)` | `// Directory match (prefix or content)` |
| 70 | `// Match de sufixo (extensões como *.log ou caminhos específicos)` | `// Suffix match (extensions like *.log or specific paths)` |
| 76 | `// 3. Arquivos ocultos por padrão` | `// 3. Hidden files by default` |

- [ ] **Step 2: Translate `pkg/utils/hash.go`**

| Line | Portuguese | English |
|---|---|---|
| 10 | `// CalculateHash gera um hash SHA256 do conteúdo de um arquivo para detecção de mudanças` | `// CalculateHash generates a SHA256 hash of file content for change detection` |

- [ ] **Step 3: Translate `pkg/utils/text.go`**

| Line | Portuguese | English |
|---|---|---|
| 20 | `// Slugify transforma uma string em um formato amigável para nomes de arquivos` | `// Slugify transforms a string into a file-name-friendly format` |
| 25 | `// 2. Remove caracteres especiais (mantém apenas letras, números e espaços)` | `// 2. Remove special characters (keep only letters, numbers and spaces)` |
| 29 | `// 3. Substitui espaços e underscores por hífens` | `// 3. Replace spaces and underscores with hyphens` |
| 33 | `// 4. Remove hífens duplicados` | `// 4. Remove duplicate hyphens` |
| 37 | `// 5. Trim hífens nas extremidades` | `// 5. Trim hyphens at edges` |
| 40 | `// Fallback para caso o slug resulte em vazio` | `// Fallback in case the slug results in empty string` |
| 47 | `// EscapeYAML prepara uma string para ser usada com segurança dentro de aspas duplas no YAML` | `// EscapeYAML prepares a string for safe use inside YAML double quotes` |

- [ ] **Step 4: Translate `pkg/sqlite/db.go`**

| Line | Portuguese | English |
|---|---|---|
| 19 | `// Init inicializa a conexão com o SQLite e configura as Pragmas de Elite` | `// Init establishes SQLite connection and configures Elite Pragmas` |
| 24 | `// InitAtPath inicializa a conexão com o SQLite em um caminho específico` | `// InitAtPath establishes SQLite connection at a specific path` |
| 39 | `// Configuração de Pragmas para Performance e Integridade` | `// Pragma Configuration for Performance and Integrity` |
| 55 | `// Configuração de Pool para Concorrência (WAL permite múltiplos leitores)` | `// Pool Configuration for Concurrency (WAL allows multiple readers)` |
| 66 | `// Close fecha a conexão com o banco` | `// Close closes the database connection` |

- [ ] **Step 5: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./pkg/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 6: Commit**

```bash
git add pkg/utils/filter.go pkg/utils/hash.go pkg/utils/text.go pkg/sqlite/db.go
git commit -m "docs: translate Portuguese comments to English in pkg/ (filter, hash, text, sqlite)"
```

---

## Task 7: Translate Portuguese comments in `internal/report/` and `internal/reflect/` (11 comments)

**Files:**
- Modify: `internal/report/aggregator.go`
- Modify: `internal/reflect/validator.go`

- [ ] **Step 1: Translate `internal/report/aggregator.go`**

| Line | Portuguese | English |
|---|---|---|
| 48 | `// FetchStats consolida todos os dados do SQLite` | `// FetchStats consolidates all SQLite data` |
| 52 | `// 1. Contagem de Nós` | `// 1. Node count` |
| 66 | `// 2. Contagem de Tasks` | `// 2. Task count` |
| 77 | `// 3. Cálculo de Success Rate e SME` | `// 3. Success Rate and SME calculation` |
| 90 | `// 4. Listagem Detalhada de Tasks (Sovereign Link Discovery)` | `// 4. Detailed Task Listing (Sovereign Link Discovery)` |
| 102 | `// Tenta encontrar o ADR via padrão no disco` | `// Attempts to find ADR via pattern on disk` |
| 114 | `// GenerateMarkdown gera o arquivo de dashboard persistence` | `// GenerateMarkdown generates the persistence dashboard file` |

- [ ] **Step 2: Translate `internal/reflect/validator.go`**

| Line | Portuguese | English |
|---|---|---|
| 35 | `// ValidateProject varre o projeto em busca de violações de Standards` | `// ValidateProject scans the project for Standards violations` |
| 62 | `// ValidatePath garante que o caminho fornecido pelo agente é seguro (Standard #10).` | `// ValidatePath ensures the agent-provided path is safe (Standard #10).` |
| 71 | `// 2. Bloqueia tentativa de sair do diretório do projeto (Path Traversal)` | `// 2. Blocks attempts to escape the project directory (Path Traversal)` |
| 79 | `// ValidateCommand valida se o commando shell é permitido e não contém injeções.` | `// ValidateCommand validates whether the shell command is allowed and injection-free.` |

- [ ] **Step 3: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/report/... ./internal/reflect/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/report/aggregator.go internal/reflect/validator.go
git commit -m "docs: translate Portuguese comments to English in report and reflect packages"
```

---

## Task 8: Translate Portuguese comments in `internal/graph/` (31 comments across 7 files)

**Files:**
- Modify: `internal/graph/engine.go`
- Modify: `internal/graph/adr_generator.go`
- Modify: `internal/graph/linker.go`
- Modify: `internal/graph/scanner_go.go`
- Modify: `internal/graph/scanner_treesitter.go`
- Modify: `internal/graph/types.go`
- Modify: `internal/graph/visualizer.go`

- [ ] **Step 1: Translate `internal/graph/engine.go`**

| Line | Portuguese | English |
|---|---|---|
| 50 | `// Notifica de forma assíncrona com backpressure protection` | `// Notifies asynchronously with backpressure protection` |
| 70 | `// ScanProject varre o diretório e coordena os scanners registrados` | `// ScanProject scans the directory and coordinates registered scanners` |
| 75 | `// Inicializa o filtro soberano baseado no .gitignore` | `// Initializes the sovereign filter based on .gitignore` |
| 82 | `// 1. Inicia o Worker Pool` | `// 1. Start the Worker Pool` |
| 95 | `// Verificação de Hash Incremental movida para o Engine para ser global` | `// Incremental Hash Verification moved to Engine to be global` |
| 102 | `// 2. Coletor de Resultados (Escrita no DB serializada)` | `// 2. Result Collector (Serialized DB writes)` |
| 123 | `// 3. File Walker utilizando o novo filtro dinâmico` | `// 3. File Walker using the new dynamic filter` |
| 167 | `// Garante que o nó do arquivo tenha o hash atualizado` | `// Ensures the file node has the updated hash` |
| 212 | `// Notifica observadores após commit bem sucedido` | `// Notifies observers after successful commit` |

- [ ] **Step 2: Translate `internal/graph/adr_generator.go`**

| Line | Portuguese | English |
|---|---|---|
| 13 | `// ADRGenerator gerencia a criação física de Architectural Decision Records` | `// ADRGenerator manages the physical creation of Architectural Decision Records` |
| 25 | `// ADRData contém todas as informações necessárias para gerar um registro de decisão` | `// ADRData contains all information needed to generate a decision record` |
| 36 | `// Generate cria um novo arquivo de ADR baseado nos dados fornecidos` | `// Generate creates a new ADR file based on the provided data` |
| 39 | `// Limita o slug para não estourar o nome do arquivo` | `// Limits the slug to avoid overflowing the filename` |
| 47 | `// Template Smart ADR com Frontmatter Blindado` | `// Smart ADR Template with Hardened Frontmatter` |
| 88 | `// Garante que o diretório existe` | `// Ensures the directory exists` |

- [ ] **Step 3: Translate `internal/graph/linker.go`**

| Line | Portuguese | English |
|---|---|---|
| 12 | `// LinkDependencies resolve imports temporários para referências reais entre arquivos.` | `// LinkDependencies resolves temporary imports to real cross-file references.` |
| 14 | `// 1. Busca todos os imports pendentes` | `// 1. Fetch all pending imports` |
| 45 | `// Remove o nó temporário após resolução bem sucedida` | `// Remove the temporary node after successful resolution` |
| 74 | `// Possíveis extensões em ordem de prioridade` | `// Possible extensions in priority order` |

- [ ] **Step 4: Translate `internal/graph/scanner_go.go`**

| Line | Portuguese | English |
|---|---|---|
| 14 | `// Scanner Go não precisa mais do DB diretamente` | `// Go Scanner no longer needs the DB directly` |
| 44 | `// Extrai Imports do arquivo Go` | `// Extracts imports from the Go file` |

- [ ] **Step 5: Translate `internal/graph/scanner_treesitter.go`**

| Line | Portuguese | English |
|---|---|---|
| 18 | `// TreeSitterScanner agora utiliza o motor real Tree-sitter com suporte a concorrência segura` | `// TreeSitterScanner uses the real Tree-sitter engine with concurrency support` |
| 19 | `// e extração semântica via Queries.` | `// and semantic extraction via Queries.` |
| 172 | `// Determina o tipo real (Heurística de Componente React)` | `// Determine the actual type (React Component Heuristic)` |
| 187 | `// Encontra o nó pai para pegar o range real (ex: interface_declaration inteiro)` | `// Find the parent node to get the real range (e.g. full interface_declaration)` |

- [ ] **Step 6: Translate `internal/graph/types.go`**

| Line | Portuguese | English |
|---|---|---|
| 7 | `// Node representa um símbolo ou arquivo no grafo` | `// Node represents a symbol or file in the graph` |
| 26 | `// ScanResult contém os dados extraídos de um único arquivo` | `// ScanResult contains data extracted from a single file` |
| 33 | `// FileScanner é a interface que cada driver de linguagem deve implementar` | `// FileScanner is the interface each language driver must implement` |

- [ ] **Step 7: Translate `internal/graph/visualizer.go`**

| Line | Portuguese | English |
|---|---|---|
| 27 | `// GenerateMasterDiagram gera o C4 holístico do projeto` | `// GenerateMasterDiagram generates the holistic C4 of the project` |
| 54 | `// GenerateTaskSnapshot gera um diagrama focado nos nós impactados por uma tarefa` | `// GenerateTaskSnapshot generates a diagram focused on nodes impacted by a task` |
| 116 | `// GenerateC4ContainerDiagram gera um diagrama C4 de Nível 2 (Container)` | `// GenerateC4ContainerDiagram generates a C4 Level 2 (Container) diagram` |

- [ ] **Step 8: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/graph/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 9: Commit**

```bash
git add internal/graph/
git commit -m "docs: translate Portuguese comments to English in internal/graph/ (7 files)"
```

---

## Task 9: Translate Portuguese comments in `internal/state/`, `internal/agents/`, `internal/audit/`, `internal/bridge/`, `internal/registry/` (14 comments)

**Files:**
- Modify: `internal/state/manager.go`
- Modify: `internal/agents/tools.go`
- Modify: `internal/agents/dispatcher.go`
- Modify: `internal/audit/runner.go`
- Modify: `internal/bridge/prompt_factory.go`
- Modify: `internal/registry/commands.go`

- [ ] **Step 1: Translate `internal/state/manager.go`**

| Line | Portuguese | English |
|---|---|---|
| 35 | `// CreateTask cria uma nova tarefa no banco` | `// CreateTask creates a new task in the database` |
| 46 | `// StartTask marca a tarefa como em progression` | `// StartTask marks the task as in progress` |
| 56 | `// GetTaskByID busca uma tarefa específica` | `// GetTaskByID fetches a specific task` |
| 78 | `// GetActiveTask retorna a tarefa que está em progression` | `// GetActiveTask returns the task that is in progress` |
| 105 | `// SQLite CURRENT_TIMESTAMP é "YYYY-MM-DD HH:MM:SS"` | `// SQLite CURRENT_TIMESTAMP is "YYYY-MM-DD HH:MM:SS"` |

- [ ] **Step 2: Translate `internal/agents/tools.go`**

| Line | Portuguese | English |
|---|---|---|
| 171 | `// Garante que o diretório pai exista` | `// Ensures the parent directory exists` |

- [ ] **Step 3: Translate `internal/agents/dispatcher.go`**

| Line | Portuguese | English |
|---|---|---|
| 61 | `// Persistir sub-task no Ledger Central (Apenas o Dispatcher escreve aqui)` | `// Persist sub-task in Central Ledger (Only the Dispatcher writes here)` |
| 122 | `// Atualização Atômica no Ledger (Standard #13)` | `// Atomic Update on Ledger (Standard #13)` |

- [ ] **Step 4: Translate `internal/audit/runner.go`**

| Line | Portuguese | English |
|---|---|---|
| 30 | `// ExecuteAudit roda o commando de verificação para uma tarefa específica com timeout e proteção de shell` | `// ExecuteAudit runs the verification command for a specific task with timeout and shell protection` |
| 61 | `// Uso do errors.As para detecção robusta de erro de saída` | `// Use errors.As for robust exit error detection` |

- [ ] **Step 5: Translate `internal/bridge/prompt_factory.go`**

| Line | Portuguese | English |
|---|---|---|
| 53 | `// GeneratePayload constrói o payload estruturado para a Engine` | `// GeneratePayload builds the structured payload for the Engine` |

- [ ] **Step 6: Translate `internal/registry/commands.go`**

| Line | Portuguese | English |
|---|---|---|
| 19 | `// Register adiciona uma factory ao registry global de forma thread-safe.` | `// Register adds a factory to the global registry in a thread-safe way.` |
| 26 | `// GetCommands retorna uma cópia defensiva das factories registradas.` | `// GetCommands returns a defensive copy of registered factories.` |
| 35 | `// ResetForTesting limpa o registry global. Apenas para uso em testes.` | `// ResetForTesting clears the global registry. Only for test use.` |

- [ ] **Step 7: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/state/... ./internal/agents/... ./internal/audit/... ./internal/bridge/... ./internal/registry/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 8: Commit**

```bash
git add internal/state/manager.go internal/agents/tools.go internal/agents/dispatcher.go internal/audit/runner.go internal/bridge/prompt_factory.go internal/registry/commands.go
git commit -m "docs: translate Portuguese comments to English in state, agents, audit, bridge, registry packages"
```

---

## Task 10: Translate Portuguese comments in `cmd/sentinel/` and remaining files (9 comments)

**Files:**
- Modify: `cmd/sentinel/main.go`
- Modify: `cmd/sentinel/commands/root.go`
- Modify: `cmd/sentinel/commands/scan.go`
- Modify: `cmd/sentinel/commands/audit.go`
- Modify: `cmd/sentinel/commands/instruct.go`
- Modify: `internal/liveview/server.go`

- [ ] **Step 1: Translate `cmd/sentinel/main.go`**

| Line | Portuguese | English |
|---|---|---|
| 12 | `// 1. Inicializa o Cérebro (SQLite)` | `// 1. Initialize the Brain (SQLite)` |
| 20 | `// 2. Executa o CLI injetando o banco` | `// 2. Execute the CLI with injected database` |

- [ ] **Step 2: Translate `cmd/sentinel/commands/root.go`**

| Line | Portuguese | English |
|---|---|---|
| 25 | `// Agrega todos os subcomandos registrados dinamicamente` | `// Aggregates all dynamically registered subcommands` |

- [ ] **Step 3: Translate `cmd/sentinel/commands/scan.go`**

| Line | Portuguese | English |
|---|---|---|
| 31 | `// Auto-Migração: garante que o banco esteja pronto` | `// Auto-Migration: ensures the database is ready` |
| 39 | `// Inicializa o Engine Multi-Linguagem` | `// Initializes the Multi-Language Engine` |

- [ ] **Step 4: Translate `cmd/sentinel/commands/audit.go`**

| Line | Portuguese | English |
|---|---|---|
| 42 | `// 1. Sovereign Gate: Validação de Padrões` | `// 1. Sovereign Gate: Standards Validation` |
| 62 | `// 2. Technical Gate: Build & Tests` | `// 2. Technical Gate: Build & Tests` |

- [ ] **Step 5: Translate `cmd/sentinel/commands/instruct.go`**

| Line | Portuguese | English |
|---|---|---|
| 169 | `// Query real no banco para encontrar os 3 arquivos mais complexos (God Objects)` | `// Real database query to find the 3 most complex files (God Objects)` |

- [ ] **Step 6: Translate `internal/liveview/server.go`**

| Line | Portuguese | English |
|---|---|---|
| 195 | `// Servir o build do Vite` | `// Serve the Vite build` |

- [ ] **Step 7: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./cmd/sentinel/... ./internal/liveview/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 8: Commit**

```bash
git add cmd/sentinel/ internal/liveview/server.go
git commit -m "docs: translate Portuguese comments to English in cmd/sentinel/ and liveview"
```

---

## Task 11: Translate Portuguese comments in `internal/patterns/` test files + remaining test files (22 comments)

**Files:**
- Modify: `cmd/sentinel/commands/pattern_test.go`
- Modify: `internal/patterns/backfill_test.go`
- Modify: `internal/patterns/dedup_test.go`
- Modify: `internal/patterns/store_test.go`

- [ ] **Step 1: Translate `cmd/sentinel/commands/pattern_test.go`**

| Line | Portuguese | English |
|---|---|---|
| 355 | `// CG-02: NewPatternCmd com nil DB deve retornar ErrNilDB na execução` | `// CG-02: NewPatternCmd with nil DB should return ErrNilDB on execution` |

- [ ] **Step 2: Translate `internal/patterns/backfill_test.go`**

| Line | Portuguese | English |
|---|---|---|
| 147 | `// CG-01: Testes de Falso Positivo — strings.Contains para classificação` | `// CG-01: False Positive Tests — strings.Contains for classification` |
| 148 | `// deve ser testado contra inputs que match a substring mas não são itens válidos.` | `// should be tested against inputs that match the substring but are not valid items.` |
| 152 | `// [AP- em comentário HTML sem pipes — len(parts) < 5 não gera candidato` | `// [AP- in HTML comment without pipes — len(parts) < 5 does not generate candidate` |
| 167 | `// Regra/MO antes de "### PMO-" — inPMO == false, não captura` | `// Rule/MO before "### PMO-" — inPMO == false, does not capture` |
| 197 | `// FP DOCUMENTADO: strings.Contains("Gaps Estruturais") em body ativa section detector` | `// DOCUMENTED FP: strings.Contains("Structural Gaps") in body triggers section detector` |
| 198 | `// Mecanismo: parseEvolutionInsights usa strings.Contains para detectar seção,` | `// Mechanism: parseEvolutionInsights uses strings.Contains to detect section,` |
| 199 | `// o que match substring em qualquer contexto. O FP é conhecido — a linha` | `// which matches substring in any context. The FP is known — the line` |
| 200 | `// "Veja Gaps Estruturais acima para contexto" vira candidato espúrio porque` | `// "See Structural Gaps above for context" becomes a spurious candidate because` |
| 201 | `// a seção está ativa quando o parser a encontra.` | `// the section is active when the parser finds it.` |
| 242 | `// FP DOCUMENTADO: "Filtro A" em texto narrativo sem prefix "- "/"* " —` | `// DOCUMENTED FP: "Filter A" in narrative text without "- "/"* " prefix —` |
| 243 | `// parseSentinelLine não exige prefixo de lista, apenas strings.Contains("Filtro A/B/C"),` | `// parseSentinelLine does not require list prefix, only strings.Contains("Filter A/B/C"),` |
| 244 | `// logo texto narrativo com substring vira candidato espúrio se len(clean)>10` | `// so narrative text with substring becomes spurious candidate if len(clean)>10` |
| 268 | `// "Filtro A" em linha curta — len(clean) > 10 protege` | `// "Filter A" in short line — len(clean) > 10 protects` |
| 281 | `// Cobertura: BackfillFromSentinelLog non-dry-run (caminho de inserção real)` | `// Coverage: BackfillFromSentinelLog non-dry-run (real insert path)` |
| 295 | `// Cria arquivo sentinel-log com conteúdo Filtro para testar inserção non-dry-run` | `// Create sentinel-log file with Filter content to test non-dry-run insert` |
| 358 | `// Cobertura: insertIfNew — caminho de erro (Create falha)` | `// Coverage: insertIfNew — error path (Create fails)` |
| 372 | `// Fecha o DB para forçar erro no Create dentro de insertIfNew` | `// Close the DB to force error on Create inside insertIfNew` |
| 389 | `// CG-02: Métodos BackfillFrom* devem retornar ErrNilDB quando store não tem DB` | `// CG-02: BackfillFrom* methods should return ErrNilDB when store has no DB` |

- [ ] **Step 3: Translate `internal/patterns/dedup_test.go`**

| Line | Portuguese | English |
|---|---|---|
| 91 | `// CG-02: FindSimilar deve retornar ErrNilDB quando store não tem DB` | `// CG-02: FindSimilar should return ErrNilDB when store has no DB` |
| 102 | `// Cobertura: FindSimilar — ramo de tag overlap (sem match levenshtein, mas overlap ≥ 0.5)` | `// Coverage: FindSimilar — tag overlap branch (no levenshtein match, but overlap ≥ 0.5)` |

- [ ] **Step 4: Translate `internal/patterns/store_test.go`**

| Line | Portuguese | English |
|---|---|---|
| 35 | `// CG-02: Validação nil-DB em métodos exportados — cada método deve` | `// CG-02: nil-DB validation on exported methods — each method should` |
| 36 | `// retornar ErrNilDB independence do wiring do construtor.` | `// return ErrNilDB regardless of constructor wiring.` |
| 312 | `// Cobertura: List — filtro por Impact` | `// Coverage: List — filter by Impact` |

- [ ] **Step 5: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/patterns/... ./cmd/sentinel/commands/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/patterns/ cmd/sentinel/commands/pattern_test.go
git commit -m "docs: translate Portuguese comments to English in patterns tests and cmd tests"
```

---

## Task 12: Fix "PAC angle analysis" literal duplication (go:S1192)

**Files:**
- Modify: `internal/agents/engine.go`

**Issue:** Literal `"PAC angle analysis"` appears 3 times (lines 546, 551, and one more in the same function).

- [ ] **Step 1: Add constant**

In `internal/agents/engine.go`, add this constant in the `const` block or before the function:

```go
const pacAngleLogMsg = "PAC angle analysis"
```

- [ ] **Step 2: Replace occurrences**

Replace all 3 occurrences of `slog.Info("PAC angle analysis", "angle", ...)` with `slog.Info(pacAngleLogMsg, "angle", ...)`.

- [ ] **Step 3: Build and test**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go build ./...
go test ./internal/agents/... -count=1
```

Expected: Build passes, all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/agents/engine.go
git commit -m "fix: extract 'PAC angle analysis' literal to pacAngleLogMsg constant (S1192)"
```

---

## Task 13: Final Verification

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

Expected: All packages pass, no races detected.

- [ ] **Step 3: Linter verification**

```bash
/home/emiyakiritsugu/go/bin/golangci-lint run
```

Expected: 0 issues.

- [ ] **Step 4: Confirm no Portuguese comments remain**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./.worktrees/*" -not -path "./web/node_modules/*" -exec grep -l "ção\|ões\|ência\|ância\|ário\|diretório\|vare\|consolida\|Garante\|Agrega\|Busca\|Extrai\|Cria\|Fecha\|Retorna\|Notifica\|Gerencia\|Configura" {} \; | grep -v vendor | grep -v ".worktrees" | grep -v "node_modules"
```

Expected: **No output** (zero files with Portuguese comments).

- [ ] **Step 5: Commit remaining changes**

```bash
git add -A
git commit -m "chore: final verification after SonarCloud remediation — 0 Portuguese comments"
```

---

## Self-Review

### 1. Spec Coverage

| Requirement | Task |
|---|---|
| Remove install.sh + coverage.out | Task 1 ✅ |
| Fix literal "too many matches found" | Task 2 ✅ |
| Group collectADRDetails params | Task 3 ✅ |
| Reduce executeQuery complexity (31→≤15) | Task 4 ✅ |
| Reduce buildNodeToContainerMap complexity (19→≤15) | Task 5 ✅ |
| Translate pkg/ comments | Task 6 ✅ |
| Translate report/reflect comments | Task 7 ✅ |
| Translate graph/ comments | Task 8 ✅ |
| Translate agents/state/audit/bridge/registry | Task 9 ✅ |
| Translate cmd/ + remaining | Task 10 ✅ |
| Translate test comments | Task 11 ✅ |
| Fix "PAC angle analysis" literal | Task 12 ✅ |
| Final verification | Task 13 ✅ |

### 2. Placeholder Scan

- ✅ No "TBD", "TODO", "implement later"
- ✅ All translations are exact Portuguese→English — no "(translate to English)" stubs
- ✅ Every code change step shows exact code
- ✅ Every commit message is complete
- ✅ All multi-line translations are individually listed (no combined rows)

### 3. Type Consistency

- ✅ Translation tables are exact file:line mappings
- ✅ Constant naming follows Go conventions (camelCase)
- ✅ All helper methods use consistent receiver type `(s *TreeSitterScanner)`

### Note on Pre-existing Issues NOT Covered

These issues existed before the PR and are out of scope for this plan:
- `ScanProject` complexity 26 (S3776) — pre-existing
- `Migrate` complexity 21 (S3776) — pre-existing
- `FetchStats` complexity 17 (S3776) — pre-existing
- `TestRunPACDeliberation` complexity 22 (S3776) — pre-existing
- `context.Context` field in struct (godre:S8242) — pre-existing
- Legacy TypeScript issues (legacy/ts/) — pre-existing

PR-introduced issues now ALL covered:
- `executeQuery` complexity 31 → Task 4 ✅
- `buildNodeToContainerMap` complexity 19 → Task 5 ✅
- `errTooManyMatches` literal → Task 2 ✅
- `pacAngleLogMsg` literal → Task 12 ✅
- `collectADRDetails` params → Task 3 ✅
- ~100 Portuguese comments → Tasks 6-11 ✅
