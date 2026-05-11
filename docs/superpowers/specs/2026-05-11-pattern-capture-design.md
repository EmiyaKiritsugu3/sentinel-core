# Pattern Capture — Design Spec
**Date:** 2026-05-11
**Status:** Pending user approval
**Estimated implementation:** ~6h
**ROADMAP reference:** Fase Futura — Pattern Capture

---

## 1. Executive Summary

Pattern Capture é um subsistema de indexação e captura de padrões arquiteturais e cognitivos. O problema real não é falta de captura — o projeto já tem 4 mecanismos (COGNITIVE-DNA.md, EVOLUTION-INSIGHTS.md, sentinel-log.md, Epiphany Protocol). O problema é **discoverability e deduplicação**: padrões estão presos em narrativas de sessão, sem query unificada, sem detecção de duplicata.

Este spec define a tabela `patterns` no SQLite existente, comandos CLI (`sentinel pattern`), backfill dos padrões já documentados, e extensão do Epiphany Protocol com Filtro D (Decision Routing).

---

## 2. Problem

| Mecanismo | Conteúdo | Problema |
|---|---|---|
| `COGNITIVE-DNA.md` | 3 anti-patterns (AP-01 a AP-03) + 3 PMOs | Estático, sem busca, sem relação com outros padrões |
| `EVOLUTION-INSIGHTS.md` | Gaps estruturais + 1 padrão cognitivo | Misturado com icebox, não-queryável |
| `sentinel-log.md` | 569 linhas, 15+ sessões de epifanias (Filtros A/B/C) | Narrativo, não-indexável, deduplicação impossível |
| Epiphany Protocol | Workflow de captura em tempo real | Sem Filtro D para princípios de roteamento |

**Consequência**: Ao começar uma sessão, o agente lê 4 arquivos diferentes para encontrar padrões relevantes. Não pode responder "quais padrões temos sobre diagnóstico?" ou "esse anti-pattern já foi capturado?". Semantic Search (fase futura) depende de uma tabela indexável como feeder.

---

## 3. Scope

### In scope (v1)
- Tabela `patterns` no SQLite (schema.go)
- CRUD interno em Go (Create, List, Search, Get)
- Comandos CLI: `sentinel pattern add/list/search/get`
- Comando de backfill: `sentinel pattern backfill`
- Filtro D no Epiphany Protocol (GEMINI.md)
- Deduplicação heurística no `pattern add`

### Out of scope (documented for future)
- **Embeddings / similarity search vetorial**: Pré-requisito é Semantic Search (fase futura). v1 usa FTS5 + Levenshtein.
- **Auto-extração de padrões de sessões**: Precisa de LLM. Sentinel é determinístico. v1 = captura manual + backfill.
- **UI Web para patterns**: WebUI (web/) não recebe mudanças. v1 = CLI-only.
- **Migração dos .md files para DB-only**: COGNITIVE-DNA, EVOLUTION-INSIGHTS, sentinel-log continuam como source-of-truth narrativo. A tabela `patterns` é um **índice** sobre eles, não um substituto.

---

## 4. Architecture

### 4.1 Data flow

```
┌─────────────────────────────────────────────────────────────┐
│ patterns (SQLite — graph.db)                                │
│ id | title | description | category | source | tags | impact│
├─────────────────────────────────────────────────────────────┤
│ ↑ Backfill          ↑ CLI add           ↑ Filtro D          │
│ (3 .md files)       (manual)            (Epiphany Protocol) │
└─────────────────────────────────────────────────────────────┘
        │
        ▼
  sentinel pattern {add|list|search|get|backfill}
```

### 4.2 Schema

Adicionado ao `internal/graph/schema.go` como nova tabela no `schema` const:

```sql
CREATE TABLE IF NOT EXISTS patterns (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL CHECK(category IN (
        'anti-pattern',
        'cognitive-pattern',
        'structural-principle',
        'routing-principle'
    )),
    source TEXT NOT NULL CHECK(source IN (
        'cognitive-dna',
        'evolution-insights',
        'sentinel-log',
        'manual',
        'epiphany'
    )),
    source_ref TEXT,
    tags TEXT NOT NULL DEFAULT '',
    impact TEXT NOT NULL DEFAULT 'medium' CHECK(impact IN ('high', 'medium', 'low')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**CHECK constraints** garantem categorias e sources válidos sem tabela auxiliar. SQLite não suporta ALTER TABLE ADD CHECK — a tabela é criada com constraints desde o início. **Nota**: Tabelas existentes no schema usam comments para documentar valores válidos (ex: `status TEXT NOT NULL -- PENDING, IN_PROGRESS, ...`). Esta tabela usa CHECK constraints — melhoria intencional sobre o padrão existente, aplicável porque é uma tabela nova sem necessidade de migration incremental sobre dados legados.

**FTS5 index** para busca full-text:

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS patterns_fts USING fts5(
    title,
    description,
    tags,
    content=patterns,
    content_rowid=rowid
);
```

FTS5 `content=patterns` usa a tabela principal como conteúdo (external content FTS). Triggers mantêm sincronia:

```sql
CREATE TRIGGER IF NOT EXISTS patterns_ai AFTER INSERT ON patterns BEGIN
    INSERT INTO patterns_fts(rowid, title, description, tags)
    VALUES (new.rowid, new.title, new.description, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS patterns_ad AFTER DELETE ON patterns BEGIN
    INSERT INTO patterns_fts(patterns_fts, rowid, title, description, tags)
    VALUES ('delete', old.rowid, old.title, old.description, old.tags);
END;

CREATE TRIGGER IF NOT EXISTS patterns_au AFTER UPDATE ON patterns BEGIN
    INSERT INTO patterns_fts(patterns_fts, rowid, title, description, tags)
    VALUES ('delete', old.rowid, old.title, old.description, old.tags);
    INSERT INTO patterns_fts(rowid, title, description, tags)
    VALUES (new.rowid, new.title, new.description, new.tags);
END;
```

**pragmaTableInfo update** em `schema.go`:

```go
var pragmaTableInfo = map[string]string{
    // ... existing tables ...
    "patterns": "PRAGMA table_info(patterns)",
}
```

### 4.3 Categories

| Categoria | Descrição | Origem primária |
|---|---|---|
| `anti-pattern` | Comportamento recorrente que degrada qualidade | COGNITIVE-DNA (AP-01 a AP-03) |
| `cognitive-pattern` | Modos de operação cognitiva do agente | EVOLUTION-INSIGHTS, sentinel-log |
| `structural-principle` | Princípios arquiteturais do projeto | EVOLUTION-INSIGHTS, ADRs |
| `routing-principle` | Princípios de roteamento de decisões | Filtro D (novo) |

### 4.4 Sources

| Source | Descrição |
|---|---|
| `cognitive-dna` | Extraído de `docs/process/COGNITIVE-DNA.md` |
| `evolution-insights` | Extraído de `docs/process/EVOLUTION-INSIGHTS.md` |
| `sentinel-log` | Extraído de `docs/process/sentinel-log.md` |
| `manual` | Captura manual via `sentinel pattern add` |
| `epiphany` | Captura via Filtro D do Epiphany Protocol |

---

## 5. New files

```
internal/
├── patterns/
│   ├── store.go        ← PatternStore + CRUD methods
│   ├── store_test.go   ← Unit tests for CRUD + search + dedup
│   └── backfill.go     ← Backfill logic (parse .md files → insert patterns)
cmd/sentinel/commands/
└── pattern.go          ← sentinel pattern {add|list|search|get|backfill}
```

`internal/patterns/` é um novo pacote. Não existe dependência circular com `internal/graph/` — `PatternStore` recebe `*sqlite.DB` via injeção, como `state.Manager` e outros componentes.

---

## 6. Implementation detail

### 6.1 PatternStore (internal/patterns/store.go)

```go
package patterns

type Pattern struct {
    ID          string
    Title       string
    Description string
    Category    string
    Source      string
    SourceRef   string
    Tags        string
    Impact      string
    CreatedAt   string
    UpdatedAt   string
}

type PatternStore struct {
    db *sqlite.DB
}

func NewPatternStore(db *sqlite.DB) (*PatternStore, error)  // validates db != nil via sqlite.ValidateDB
func (s *PatternStore) Create(p *Pattern) (string, error)   // UUID, insert, return id
func (s *PatternStore) List(filters ListFilters) ([]Pattern, error)
func (s *PatternStore) Search(query string) ([]Pattern, error)  // FTS5 MATCH
func (s *PatternStore) Get(id string) (*Pattern, error)
func (s *PatternStore) FindSimilar(title string, tags []string) ([]Pattern, error)  // dedup helper

type ListFilters struct {
    Category string
    Source   string
    Impact   string
    Limit    int
}
```

**Nil Guard**: `NewPatternStore` chama `sqlite.ValidateDB(db, "pattern-store")` seguindo o padrão CG-02 existente.

**UUID**: Usa `github.com/google/uuid` com `uuid.New().String()` — mesmo padrão de `state.Manager.CreateTask` e `internal/agents/tools.go`. IDs são UUIDs sem prefixo (ex: `a1b2c3d4-e5f6-7890-abcd-ef1234567890`).

### 6.2 Deduplication (FindSimilar)

No `pattern add`, antes de criar:

1. `FindSimilar(title, tags)` busca por:
   - `title` com Levenshtein distance ≤ 3 contra todos os `patterns.title` existentes
   - Overlap de tags (≥ 50% das tags em comum)
2. Se encontrado, CLI imprime:
   ```
   [SENTINEL] Similar pattern found: "Diagnóstico sem dado empírico = loop" (ID: a1b2c3d4-...)
   [SENTINEL] Use --force to create anyway, or link to existing instead.
   ```
3. Flag `--force` no comando `add` ignora dedup e cria.

**Levenshtein**: Implementação inline (sem dependência). Threshold 3 para títulos típicos de 3-8 palavras. Não é perfeito, mas v1 não requer embeddings.

### 6.3 Backfill (internal/patterns/backfill.go)

```go
type BackfillResult struct {
    Extracted  int
    Inserted   int
    Skipped    int  // duplicates
    Errors     []string
}

func (s *PatternStore) BackfillFromCognitiveDNA() (BackfillResult, error)
func (s *PatternStore) BackfillFromEvolutionInsights() (BackfillResult, error)
func (s *PatternStore) BackfillFromSentinelLog() (BackfillResult, error)
```

**Backfill é semi-automático para sentinel-log**: O comando `sentinel pattern backfill` extrai candidatos de cada fonte, mas o usuário valida. Fluxo:

1. `sentinel pattern backfill --source cognitive-dna` → parseia AP-01 a AP-03 + PMOs, insere automaticamente (formato estruturado, baixo risco de erro)
2. `sentinel pattern backfill --source evolution-insights` → parseia seções estruturadas, insere automaticamente
3. `sentinel pattern backfill --source sentinel-log --dry-run` → lista candidatos extraídos sem inserir. Usuário revisa e seleciona quais inserir via `sentinel pattern add`

**Sem `--dry-run`** = backfill automático (cognitive-dna e evolution-insights apenas). **Com `--dry-run`** = lista candidatos sem persistir.

**Formato de extração COGNITIVE-DNA**: Seções com header `## AP-XX` ou `## PMO-XX` são parseadas. Title = nome do anti-pattern/PMO. Description = corpo da seção. Category = `anti-pattern` ou `structural-principle`.

### 6.4 CLI (cmd/sentinel/commands/pattern.go)

Registro via `registry.Register(NewPatternCmd)` como os demais comandos.

```go
func init() {
    registry.Register(NewPatternCmd)
}

func NewPatternCmd(db *sqlite.DB) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "pattern",
        Short: "Capture and query architectural and cognitive patterns",
    }

    cmd.AddCommand(patternAddCmd(db))
    cmd.AddCommand(patternListCmd(db))
    cmd.AddCommand(patternSearchCmd(db))
    cmd.AddCommand(patternGetCmd(db))
    cmd.AddCommand(patternBackfillCmd(db))

    return cmd
}
```

#### `sentinel pattern add`

```bash
sentinel pattern add \
    --title "Diagnóstico sem dado empírico = loop" \
    --desc "Quando o agente diagnostica sem dados empíricos, entra em loop de hipóteses" \
    --category anti-pattern \
    --tags "loop,diagnosis,empirical" \
    --impact high \
    [--source manual] \     # default: manual
    [--source-ref "COGNITIVE-DNA.md:AP-01"] \
    [--force]               # skip dedup check
```

**Campos obrigatórios**: `--title`, `--desc`, `--category`. `--impact` default = `medium`. `--source` default = `manual`. `--tags` default = `""`.

**Output**: `✅ PATTERN CAPTURED [ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890]: Diagnóstico sem dado empírico = loop`

Ou se duplicata encontrada (sem `--force`):
```
[SENTINEL] Similar pattern found: "Diagnóstico sem dado empírico = loop" (ID: x9y8z7w6-...)
[SENTINEL] Use --force to create anyway.
```

#### `sentinel pattern list`

```bash
sentinel pattern list [--category anti-pattern] [--impact high] [--source cognitive-dna]
```

**Output**: Tabela com ID, Title, Category, Impact, Source. Ordenado por `created_at DESC`. Limit default = 20.

#### `sentinel pattern search`

```bash
sentinel pattern search "diagnosis"
```

**Output**: Tabela com ID, Title, Category, Impact. Usa FTS5 MATCH sobre title + description + tags. Ranking por relevância FTS5 (`bm25`).

#### `sentinel pattern get`

```bash
sentinel pattern get <uuid>
```

**Output**: Detalhe completo — todos os campos, incluindo description completa e source_ref.

#### `sentinel pattern backfill`

```bash
sentinel pattern backfill --source cognitive-dna
sentinel pattern backfill --source evolution-insights
sentinel pattern backfill --source sentinel-log --dry-run
sentinel pattern backfill --all    # cognitive-dna + evolution-insights (sem sentinel-log)
```

**Output**:
```
Backfill complete: 6 extracted, 6 inserted, 0 skipped
```

Ou com `--dry-run`:
```
[DRY-RUN] Would extract 8 candidates from sentinel-log:
  1. "Diagnóstico sem dado empírico = loop" (Filtro A, session 12)
  2. "Especificar artefato de saída reduz iterações" (Filtro C, session 8)
  ...
Use 'sentinel pattern add' to capture selected patterns.
```

---

## 7. Filtro D — Epiphany Protocol Extension

### Mudança em GEMINI.md

Na seção "PROTOCOLO DE EPIFANIA", adicionar Filtro D:

```markdown
4. **FILTRO D — Decision Routing**: Quando uma epifania revela um princípio sobre
   COMO rotear decisões (não apenas o que aconteceu), o agente DEVE capturá-lo
   via `sentinel pattern add --source epiphany --category routing-principle`.
   Exemplo: "Auditoria troca modo cognitivo de construtivo para destrutivo"
   → Filtro D → routing-principle.
```

### Categorias de Epifania atualizadas

| Filtro | Tipo | Destino | Ação |
|---|---|---|---|
| A | Behavioural | sentinel-log.md | `write_file` (já existe) |
| B | Structural | EVOLUTION-INSIGHTS.md | `write_file` (já existe) |
| C | Process | sentinel-log.md | `write_file` (já existe) |
| **D** | **Decision Routing** | **patterns table** | **`sentinel pattern add --source epiphany`** |

---

## 8. Error handling

| Scenario | Behavior |
|---|---|
| DB nil | `sqlite.ValidateDB` → `ErrNilDB` sentinel error |
| Invalid category | `CHECK` constraint fails → user-friendly error: "Invalid category. Must be: anti-pattern, cognitive-pattern, structural-principle, routing-principle" |
| Invalid source | Same as category — `CHECK` constraint error |
| Duplicate title (dedup) | Print similar pattern, suggest `--force` |
| Backfill parse error | Log warning, skip entry, continue. Return in `BackfillResult.Errors` |
| FTS5 search returns nothing | Print "No patterns found matching 'query'" |
| Pattern ID not found (get) | Print "Pattern not found: <id>" with exit code 1 |

---

## 9. Known limitations

**KNOWN_LIMITATION_01 — Levenshtein dedup is imprecise.**
Threshold of 3 may miss semantic duplicates with different wording. Accepted for v1. Future: embedding-based similarity when Semantic Search is implemented.

**KNOWN_LIMITATION_02 — Backfill is format-dependent.**
Parsing COGNITIVE-DNA and EVOLUTION-INSIGHTS assumes their current markdown structure (`## AP-XX`, `## PMO-XX`, section headers). If the format changes, backfill must be updated. Mitigation: backfill is idempotent (`INSERT OR IGNORE` by title).

**KNOWN_LIMITATION_03 — sentinel-log backfill is dry-run only by default.**
Free-form epiphanies in sentinel-log.md have no consistent structure. Automatic extraction would produce false positives. Semi-automated flow (dry-run → manual add) is the safe path.

**KNOWN_LIMITATION_04 — FTS5 is not semantic search.**
"diagnosis" matches "diagnosis" but not "diagnóstico". Portuguese/English synonyms not handled. Future: embedding-based Semantic Search.

---

## 10. Verification gate

```bash
go build ./...
go test ./internal/patterns/... -v
go vet ./...
```

All must pass before PR.

### Manual verification

```bash
# After implementation:
sentinel pattern add --title "Test pattern" --desc "Test description" --category anti-pattern --impact high
sentinel pattern list
sentinel pattern search "Test"
sentinel pattern get <id-from-list>
sentinel pattern backfill --source cognitive-dna
sentinel pattern backfill --source sentinel-log --dry-run
```

---

## 11. Implementation sequence

1. **Schema migration** — Adicionar tabela `patterns` + FTS5 + triggers em `internal/graph/schema.go`
2. **PatternStore** — CRUD interno em `internal/patterns/store.go` + testes
3. **CLI commands** — `cmd/sentinel/commands/pattern.go` com add/list/search/get
4. **Backfill** — `internal/patterns/backfill.go` + comando `backfill`
5. **Deduplication** — `FindSimilar` no `pattern add`
6. **Filtro D** — Atualizar `GEMINI.md` com Filtro D

Each step is independently testable. Step 1-2 have zero blast radius (new table, new package). Step 3 extends CLI (additive). Step 4-5 add features to existing commands. Step 6 is documentation-only.
