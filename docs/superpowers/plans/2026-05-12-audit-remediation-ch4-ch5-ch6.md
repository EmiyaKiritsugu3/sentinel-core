# Post-Audit Remediation: Ch4 (CG-02), Ch5 (Estado Global), Ch6 (CG-01 FP Tests)

**Branch:** `refactor/cg02-nil-guards` (Ch4) → `refactor/global-state-cleanup` (Ch5) → `refactor/cg01-fp-tests` (Ch6)
**Date:** 2026-05-12
**Status:** Pending
**Predecessor:** Ch3 pattern flag localization (PR #10, merged as `47a07d0`)

## Audit Findings Summary

Três eixos de auditoria executados após Chapters 1-3:

| Eixo | Violações | Severidade |
|------|-----------|------------|
| CG-02: nil validation | 24 construtores sem nil check (de 25) | 🚨 CRITICAL |
| Estado global | 3 flag vars em plan.go + 1 RootCmd morto | ⚠️ HIGH |
| CG-01: FP tests | 8 classificações sem teste de falso positivo | MEDIUM |

---

## Chapter 4: CG-02 Nil Guards

### Problem

CG-02: "Todo componente deve validar `nil` em dependências, independente do wiring global."

Apenas `NewPatternStore` é compliant. Todos os outros 24 construtores recebem dependências (`*sqlite.DB`, `*genai.Client`, interfaces) sem nenhuma validação nil. Nil dereference panic é o resultado inevitável se qualquer dependência for nil.

### Two Patterns Required

Construtores CLI e internos têm assinaturas diferentes — dois padrões são necessários.

#### Pattern A: CLI Command Constructors (retornam `*cobra.Command`)

CLI constructors não podem mudar o retorno para `(*cobra.Command, error)` porque `CommandFactory` é `func(*sqlite.DB) *cobra.Command`. A solução: validar no construtor e, se falhar, retornar um command cujo `RunE` sempre retorna o erro de validação. Isso é o que `NewPatternCmd` já faz.

```go
func NewAuditCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Run the verification gate for the active task",
	}

	if err := sqlite.ValidateDB(db, "audit-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// ... lógica real movida pra cá ...
	}
	return cmd
}
```

**Vantagem**: backward-compatible com `CommandFactory`. Cobra executa `RunE` — se o db for nil, o usuário recebe o erro em runtime ao invocar o comando, não um panic.

#### Pattern B: Internal Constructors (retornam `(*T, error)`)

Constructores internos podem e devem retornar erro. Seguir o padrão de `NewPatternStore`.

```go
func NewEngine(db *sqlite.DB) (*Engine, error) {
	if err := sqlite.ValidateDB(db, "graph-engine"); err != nil {
		return nil, err
	}
	return &Engine{db: db, scanners: make(map[string]FileScanner)}, nil
}
```

**Vantagem**: fail-fast. O chamador sabe imediatamente que a construção falhou. Sem estado inválido circulando.

### Constructors to Modify

#### Group 1: CLI Commands — Pattern A (11 construtores)

Todos seguem o mesmo molde: receber `db *sqlite.DB`, validar, mover `RunE` pra dentro do bloco de sucesso.

| # | Constructor | Arquivo | Param nil-check |
|---|------------|---------|-----------------|
| 1 | `NewAuditCmd` | `cmd/sentinel/commands/audit.go` | `db` |
| 2 | `NewInstructCmd` | `cmd/sentinel/commands/instruct.go` | `db` |
| 3 | `NewLiveCmd` | `cmd/sentinel/commands/live.go` | `db` |
| 4 | `NewPatternCmd` | `cmd/sentinel/commands/pattern.go` | `db` (JÁ COMPLIANT ✅ — apenas verificar consistência) |
| 5 | `NewPlanCmd` | `cmd/sentinel/commands/plan.go` | `db` |
| 6 | `NewReportCmd` | `cmd/sentinel/commands/report.go` | `db` |
| 7 | `NewRootCmd` | `cmd/sentinel/commands/root.go` | `db` |
| 8 | `NewScanCmd` | `cmd/sentinel/commands/scan.go` | `db` |
| 9 | `NewStartCmd` | `cmd/sentinel/commands/start.go` | `db` |
| 10 | `NewStatusCmd` | `cmd/sentinel/commands/status.go` | `db` |
| 11 | `NewVisualizeCmd` | `cmd/sentinel/commands/visualize.go` | `db` |

**NewPatternCmd já é compliant** — usa `sqlite.ValidateDB(db, "pattern-cmd")` e seta `cmd.RunE` com erro se falhar. Os outros 10 precisam seguir o mesmo padrão.

#### Group 2: Internal Constructors — Pattern B (13 construtores)

Mudar retorno de `*T` para `(*T, error)`. Adicionar `sqlite.ValidateDB` ou nil check manual.

| # | Constructor | Arquivo | Params nil-check | Callers to Update |
|---|------------|---------|------------------|-------------------|
| 1 | `NewDispatcher` | `internal/agents/dispatcher.go` | `registry`, `shield`, `db` (3 params) | `start.go:40` |
| 2 | `NewEngine` (agents) | `internal/agents/engine.go` | `r`, `auth`, `v`, `db` (4 params) — JÁ retorna `(*Engine, error)` mas sem nil check explícito | `start.go:47` (já trata erro) |
| 3 | `NewMutationEngine` | `internal/agents/mutation.go` | `db` | Verificar callers |
| 4 | `NewRegistryManager` | `internal/agents/registry.go` | `db` | `start.go:39` |
| 5 | `NewEngine` (graph) | `internal/graph/engine.go` | `db` | `scan.go`, `live.go` |
| 6 | `NewVisualizer` | `internal/graph/visualizer.go` | `db` | `visualize.go` |
| 7 | `NewRunner` | `internal/audit/runner.go` | `db` | `audit.go:53` |
| 8 | `NewGeminiClassifier` | `internal/bridge/gemini_classifier.go` | `client` (*genai.Client) — não é *sqlite.DB | `engine.go:68` |
| 9 | `NewFactory` | `internal/bridge/prompt_factory.go` | `db`, `classifier` (2 params) | `engine.go:70` |
| 10 | `NewValidator` | `internal/reflect/validator.go` | `db` | `audit.go:32`, `start.go:33` |
| 11 | `NewAggregator` | `internal/report/aggregator.go` | `db` | `report.go` |
| 12 | `NewManager` | `internal/state/manager.go` | `db` | Múltiplos callers |
| 13 | `NewDisambiguator` | `internal/intake/disambiguator.go` | `db` | `plan.go:41` |

### Implementation Order

**Sub-action 4.1**: CLI Commands (10 construtores — NewPatternCmd já ok)
**Sub-action 4.2**: Internal constructors com poucos callers (NewRunner, NewVisualizer, NewGeminiClassifier, NewFactory)
**Sub-action 4.3**: Internal constructors com muitos callers (NewManager, NewValidator, NewAggregator, NewDisambiguator)
**Sub-action 4.4**: Internal constructors multi-param (NewDispatcher, NewEngine agents, NewMutationEngine, NewRegistryManager, NewEngine graph)
**Sub-action 4.5**: Atualizar todos os call sites para tratar erro
**Sub-action 4.6**: Adicionar testes `TestNewXXX_NilDB` para cada construtor

### Special Cases

- **NewGeminiClassifier**: recebe `*genai.Client`, não `*sqlite.DB`. Precisa de nil check manual: `if client == nil { return nil, fmt.Errorf("gemini-classifier: nil client") }`. Adicionar `ErrNilClient` sentinel ou usar `errors.New`.
- **NewEngine (agents)**: JÁ retorna `(*Engine, error)` mas confia em `auth.GetAPIKey()` para implicitamente falhar. Adicionar nil checks explícitos para `r`, `auth`, `v`, `db` ANTES de chamar `GetAPIKey`.
- **NewFactory**: recebe `*IntentClassifier` — nil check manual. Se `classifier` pode ser nil por design (optional), documentar explicitamente.
- **NewDispatcher**: recebe 3 pointers. Todos devem ser nil-checkados.

### Tests Required

Para cada construtor modificado:

```go
func TestNewXXX_NilDB(t *testing.T) {
	_, err := NewXXX(nil)
	if err == nil {
		t.Fatal("expected error for nil db")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}
```

Para CLI commands:

```go
func TestNewXXXCmd_NilDB(t *testing.T) {
	cmd := NewXXXCmd(nil)
	if cmd == nil {
		t.Fatal("expected command even with nil db")
	}
	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when executing command with nil db")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}
```

---

## Chapter 5: Global State Cleanup

### Problem

Mesma classe de hazard resolvida no Ch3 — variáveis de flag no escopo do pacote.

### 5a: plan.go Flag Localization

**3 variáveis globais** em `cmd/sentinel/commands/plan.go:20-24`:

```go
var (
	planTier   string
	flagRefine bool
	flagNoSuggest bool
)
```

Aplicar o mesmo padrão do Ch3: mover para variáveis locais dentro de `NewPlanCmd`.

```go
func NewPlanCmd(db *sqlite.DB) *cobra.Command {
	var tier string
	var refine bool
	var noSuggest bool

	cmd := &cobra.Command{...}
	cmd.Flags().StringVar(&tier, "tier", "T2", ...)
	cmd.Flags().BoolVarP(&refine, "refine", "r", false, ...)
	cmd.Flags().BoolVar(&noSuggest, "no-suggest", false, ...)
	...
}
```

**Impacto**: Sem mudança na CLI. Flags idênticas. Sem reset helpers pra deletar (plan_test.go pode não existir ou não ter resets).

### 5b: RootCmd Dead Code Removal

`root.go:12` declara `var RootCmd = &cobra.Command{...}` mas `NewRootCmd` (linha 17) cria um NOVO command do zero. `RootCmd` nunca é usado — `Execute()` chama `NewRootCmd(db)`, não `RootCmd`.

**Ação**: Remover `RootCmd` (linhas 12-15). É código morto e estado mutável desnecessário.

### Verification

- `go vet ./...`
- `go test ./... -count=1`
- Verificar que `grep -r "RootCmd" cmd/` só retorna `NewRootCmd` (uso legítimo)

---

## Chapter 6: CG-01 FP Tests

### Problem

CG-01: "Proibido o uso de `strings.Contains` para classificação sem teste de falso positivo."

8 classificações em 2 arquivos sem nenhum teste de falso positivo.

### 6a: visualizer.go — 7 FP Tests

**Arquivo**: `internal/graph/visualizer.go`
**Uso**: 7 `strings.Contains(path, "...")` para classificar containers C4

As 7 classificações (linhas 171-183):

| Container | Substring buscada | FP Risk |
|-----------|-------------------|---------|
| CLI | `"cmd/sentinel"` | Path `docs/cmd/sentinel/README.md` seria falso positivo |
| Agents | `"internal/agents"` | Path `test/internal/agents/mock.go` seria FP |
| Graph | `"internal/graph"` | Path `docs/internal/graph/diagram.png` seria FP |
| Audit | `"internal/audit"` | Path `docs/internal/audit/report.md` seria FP |
| State | `"internal/state"` | Path `docs/internal/state/schema.md` seria FP |
| DB | `"pkg/sqlite"` | Path `docs/pkg/sqlite/setup.md` seria FP |
| Frontend | `"web/"` | Path `docs/web/api.md` seria FP |

**Test pattern**: Para cada container, criar um teste que passa um path contendo a substring mas NÃO pertencente ao container. O teste verifica se o classificador retorna o container errado (prova que o FP é possível — ou prova que o código é seguro se houver checagem adicional).

```go
func TestContainerClassification_FalsePositive_XYZ(t *testing.T) {
	// Path contém a substring mas NÃO é do container
	path := "docs/cmd/sentinel/README.md"
	result := classifyContainer(path)
	if result == "CLI" {
		t.Errorf("FP: path %q falsamente classificado como CLI", path)
	}
}
```

### 6b: instruct.go — 1 FP Test

**Arquivo**: `cmd/sentinel/commands/instruct.go:60`
**Uso**: `strings.Contains(strings.ToLower(intent), "performance")` para detectar vagueza

FP risk: "performance review", "performance metrics", "performance testing" não são vagueza — são tasks específicos.

```go
func TestVaguenessDetection_FalsePositive_PerformanceInContext(t *testing.T) {
	// "performance testing" NÃO é vago — é um task específico
	intent := "Add performance testing to the API endpoints"
	vague := isVague(intent)
	if vague {
		t.Errorf("FP: %q falsamente classificado como vago", intent)
	}
}
```

### Verification

- `go test ./internal/graph/... -v -run FalsePositive`
- `go test ./cmd/sentinel/commands/... -v -run FalsePositive`
- Full suite: `go test ./... -count=1`

---

## Execution Sequence

1. **Ch4** (CG-02) — branch `refactor/cg02-nil-guards` → PR
2. **Ch5** (Estado global) — branch `refactor/global-state-cleanup` → PR
3. **Ch6** (CG-01 FP tests) — branch `refactor/cg01-fp-tests` → PR

Cada chapter = 1 branch = 1 PR = 1 merge. Sem acumular. Sem pular.

## Risk Assessment

- **Ch4**: MAIOR RISCO — mudar assinaturas de construtores quebra callers. Testar exhaustivamente. A abordagem Pattern A (CLI) é non-breaking; Pattern B (internal) requer atualização de call sites.
- **Ch5**: BAIXO RISCO — mesmo padrão do Ch3, já provado. Refatoração mecânica.
- **Ch6**: BAIXO RISCO — adicionar testes, sem mudança em código de produção. Se FP tests falharem (provar que o classification é buggy), o fix é separado e escopo menor.

## Notes

- `NewPatternCmd` (Ch4 Group 1) já é compliant — servir como referência para os outros 10 CLI constructors.
- `NewEngine` (agents) já retorna `(*Engine, error)` — adicionar nil checks explícitos antes da lógica existente.
- `NewAggregator` e `NewDisambiguator` têm guards parciais em métodos — mover validação para o construtor (fail-fast) e remover guards de método que se tornam redundantes.
- `NewGeminiClassifier` usa `*genai.Client` — não usar `sqlite.ValidateDB`, usar nil check manual com sentinel error próprio.
