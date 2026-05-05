# Prompt Intelligence System — Design Spec
**Date:** 2026-05-04  
**Status:** Ready for audit round 2 (Section 2 critical issues corrected)  
**Estimated implementation:** ~4h  
**Audit budget spent:** ~1.5h (within 30% rule)

---

## Problem

The sentinel pipeline has two vagueness problems at different points in time:

- **T1 (input):** `sentinel plan "fix the auth bug"` — task description is ambiguous before being saved. AI gets imprecise task.
- **T2 (execution):** `loadSurgicalContext` always returns top 10 nodes by `last_indexed`, regardless of what the task actually needs. A diagnosis task gets the same context as a refactor task.

Both degrade AI output quality without any signal to the user.

---

## Scope

### In scope
- **Subsystem B:** Smart Context Routing — intent-aware payload construction (T2)
- **Subsystem A:** Input Disambiguation — vagueness detection and suggestion before task save (T1)

### Out of scope (documented for future)
- **Subsystem C:** Standalone generic prompt refiner for non-sentinel AI conversations. Excluded from MVP: without project graph grounding, refinement is templateization with no real signal gain.

---

## Architecture

### Temporal separation

```
T1 — CLI (sentinel plan)
──────────────────────────────────────────────────────────
sentinel plan "fix the auth bug" [--refine] [--no-suggest]
      │
      ▼
internal/intake/disambiguator.go
  ├── VaguenessScore(description) → score
  ├── score > 0.50?
  │     ├── No  → save as-is
  │     └── Yes → QueryGraph(keywords) → []Suggestion
  │                    ├── --no-suggest → save original, silent
  │                    ├── default      → save original + print suggestion (yellow)
  │                    └── --refine     → prompt user → save chosen description
  └── task persisted in DB

T2 — Agent execution (GeneratePayload)
──────────────────────────────────────────────────────────
GeneratePayload(ctx, taskID, persona)
      │
      ▼
internal/bridge/classifier.go   [deps: db, optional GeminiClassifier]
  ├── cache.Load(taskID) → hit? return cached Intent
  ├── heuristicClassify(description) → (Intent, confidence)
  ├── confidence < 0.60 && ai != nil?
  │     └── ai.Classify(ctx, description) → Intent
  │         (failure → keep heuristic best-guess, log to stderr)
  └── cache.Store(taskID, intent)

      │ Intent: diagnose | implement | refactor | review | unknown
      ▼
internal/bridge/router.go
  └── strategyByIntent[intent] → ContextStrategy

      │ ContextStrategy
      ▼
prompt_factory.go (modified)
  └── loadContextByStrategy(taskID, strategy) → []ContextNode → ContextPayload → AI
```

---

## New files

```
internal/
├── bridge/
│   ├── classifier.go     ← IntentClassifier + AIClassifier interface + GeminiClassifier
│   ├── router.go         ← ContextRouter + ContextStrategy + strategyByIntent
│   └── prompt_factory.go ← modified: accepts classifier, uses strategy
└── intake/
    └── disambiguator.go  ← Disambiguator + VaguenessScore + graph query
```

---

## Subsystem B: Smart Context Routing

### classifier.go

```go
type Intent string

const (
    IntentDiagnose  Intent = "diagnose"
    IntentImplement Intent = "implement"
    IntentRefactor  Intent = "refactor"
    IntentReview    Intent = "review"
    IntentUnknown   Intent = "unknown"
)

type AIClassifier interface {
    Classify(ctx context.Context, description string) (Intent, error)
}

type IntentClassifier struct {
    ai        AIClassifier  // nil = heuristic-only, graceful degradation
    threshold float64       // confidence below this triggers AI fallback
    cache     sync.Map      // goroutine-safe: taskID → Intent
}
```

**Heuristic keywords (lowercase, language-agnostic where possible):**

| Intent | Keywords |
|---|---|
| diagnose | fix, bug, error, broken, failing, crash, debug, investigate, corrigir, erro |
| implement | add, create, build, implement, new, adicionar, criar, implementar |
| refactor | refactor, cleanup, reorganize, extract, move, simplify, refatorar |
| review | review, audit, check, verify, analyze, validate, revisar, auditar |

**Confidence algorithm — presence + ambiguity, not word ratio:**

```
keywords_matched = keywords found in description (any category)
categories_hit   = number of distinct Intent categories with at least 1 match

if categories_hit == 0 → confidence = 0.00  (no match → AI fallback)
if categories_hit == 1 → confidence = 0.85  (unambiguous → heuristic wins)
if categories_hit >= 2 → confidence = 0.30  (ambiguous → AI fallback)
```

Intent = category with most keyword matches. Tie → lower confidence (0.30), AI decides.

Rationale: `matched/total` always produces low confidence for descriptive tasks
("fix JWT validation in internal/agents/auth_provider.go" → 1/7 = 0.14), defeating
the heuristic path entirely. Presence+ambiguity is semantically correct.

**GeminiClassifier — prompt and parsing:**

```
System prompt: "You are a task classifier. Respond with exactly one word."
User prompt:   "Classify this software task: diagnose, implement, refactor, or review.\nTask: {description}"
Parse:         strings.ToLower(strings.TrimSpace(response))
Validation:    if parsed value not in {diagnose, implement, refactor, review} → return IntentUnknown, nil
               (not an error — unknown is a valid graceful result)
```

**Construction in `NewEngine`** (reuses existing `genai.Client`):

```go
// engine.go — after creating genai.Client:
geminiClassifier := bridge.NewGeminiClassifier(client)
classifier := bridge.NewIntentClassifier(geminiClassifier, 0.60)
factory := bridge.NewFactory(db, classifier)
```

`Factory` signature change: `NewFactory(db *sqlite.DB, classifier *IntentClassifier) *Factory`

**Breaking call sites — must update:**
- `cmd/sentinel/commands/start.go:34` → `bridge.NewFactory(db)` → `bridge.NewFactory(db, classifier)`
- `internal/agents/engine_test.go:21` → `bridge.NewFactory(nil)` → `bridge.NewFactory(nil, bridge.NewNilClassifier())`

`NilClassifier` is a null object implementing `AIClassifier` that always returns `IntentUnknown`.
This preserves test isolation without panics.

### router.go

`ContextStrategy` fields split by data source — DB-queryable vs file-based:

```go
type ContextStrategy struct {
    // DB-queryable (nodes + edges tables)
    HighCoupling  bool  // SELECT to_node_id, COUNT(*) FROM edges GROUP BY to_node_id ORDER BY COUNT DESC
    RecentChanges bool  // ORDER BY last_indexed DESC (already default — acts as tie-breaker weight)
    IncludeTests  bool  // WHERE file_path LIKE '%_test.go'
    NodeLimit     int

    // File-based (direct read, not DB)
    IncludeADRs        bool  // reads docs/architecture/adr/*.md
    IncludeDebtMarkers bool  // reads TECHNICAL-DEBT.md, extracts sections matching task keywords
}

var strategyByIntent = map[Intent]ContextStrategy{
    IntentDiagnose:  {HighCoupling: true,  RecentChanges: true,  NodeLimit: 15},
    IntentImplement: {IncludeTests: true,  IncludeADRs: true,    NodeLimit: 10},
    IntentRefactor:  {HighCoupling: true,  IncludeDebtMarkers: true, NodeLimit: 12},
    IntentReview:    {IncludeADRs: true,   NodeLimit: 8},
    IntentUnknown:   {},  // empty → Factory uses current behavior (top 10 by last_indexed)
}
```

---

## Subsystem A: Input Disambiguation

### disambiguator.go

```go
// Weights — named constants, not magic numbers. Adjust after real-world calibration.
const (
    weightLength  = 0.25
    weightVerb    = 0.20
    weightPronoun = 0.15
    weightAnchor  = 0.40
    scoreThreshold = 0.50
)

type Disambiguator struct {
    db *sqlite.DB
}

type Suggestion struct {
    NodeName string
    FilePath string
    MatchScore float64
}
```

### VaguenessScore algorithm

**Four independent signals, summed and clamped to [0.0, 1.0]:**

```
Score = LengthSignal + VerbSignal + PronounSignal + AnchorSignal
```

**Signal 1 — Length (max 0.25):**
```
< 3 words  → 0.25
3–5 words  → 0.18
6–10 words → 0.08
> 10 words → 0.00
```

**Signal 2 — Generic Verb (max 0.20):**  
Keywords: `fix, improve, update, change, make, handle, check, corrigir, melhorar, atualizar, mudar`  
Match without specific object/condition → +0.20. "fix X where Y" → +0.00.

**Signal 3 — Pronoun (max 0.15):**  
Keywords: `it, this, the issue, the bug, the problem, isso, ele, o problema, o erro`  
Any match → +0.15.

**Signal 4 — Technical Anchor (max 0.40) — two-phase:**

Phase 1 — lexical (zero DB, <1ms):
- Contains path pattern (`internal/`, `pkg/`, `.go`) → 0.00
- Contains line reference (`:` + digit) → 0.00
- Contains error literal (capitalized word + `Error`/`Err`) → 0.00
- None of the above → proceed to Phase 2

Phase 2 — graph-anchored (DB query, only if Phase 1 = no anchor):
- Guard: if `SELECT COUNT(*) FROM nodes = 0` → skip Phase 2, return 0.40 (graph not indexed)
- Extract nouns/identifiers from description
- `SELECT name, file_path FROM nodes WHERE name LIKE '%keyword%' LIMIT 5`
- `matched_ratio = matches / total_keywords`
- `anchor_score = 0.40 × (1 - matched_ratio)`

**Calibration examples:**

| Description | L | V | P | A | Total | Decision |
|---|---|---|---|---|---|---|
| `"fix bug"` | 0.25 | 0.20 | 0.00 | 0.40 | **0.85** | suggest |
| `"fix the auth bug"` | 0.18 | 0.20 | 0.15 | 0.28* | **0.81** | suggest |
| `"fix JWT validation in internal/agents/auth_provider.go"` | 0.00 | 0.20 | 0.00 | 0.00 | **0.20** | pass |
| `"refactor loadSurgicalContext to use graph-aware ranking"` | 0.00 | 0.00 | 0.00 | 0.18* | **0.18** | pass |

*`auth` and `loadSurgicalContext` match graph nodes → `matched_ratio` reduces anchor score.

### CLI flags — sentinel/cmd/plan.go

```go
var flagRefine, flagNoSuggest bool

func init() {
    planCmd.Flags().BoolVarP(&flagRefine,    "refine",     "r", false, "interactive disambiguation")
    planCmd.Flags().BoolVar (&flagNoSuggest, "no-suggest", false,      "skip suggestion (for scripts)")
}
```

**Mode behavior:**

| Mode | Trigger | Behavior |
|---|---|---|
| Default | `sentinel plan "..."` | Save original + print `[SUGGEST] did you mean: <node> in <file>?` |
| Interactive | `--refine` / `-r` | Prompt user, save chosen description |
| Silent | `--no-suggest` | Save original, no output (CI-safe) |

**Color output:** no color library — use `[SUGGEST]` prefix, consistent with existing `[SENTINEL]`/`warning:` patterns.

**Flag conflict:** `--refine` takes precedence over `--no-suggest` when both are passed. One validation line in `RunE`:
```go
if flagRefine && flagNoSuggest { flagNoSuggest = false }
```

**Note:** isatty auto-detection for CI excluded from MVP. Operators use `--no-suggest` explicitly.

---

## Error handling

| Scenario | Behavior |
|---|---|
| Gemini unavailable | Silent fallback to heuristic. `fmt.Fprintf(os.Stderr, "warning: classifier: ...")` |
| Graph not indexed | Skip Phase 2. AnchorSignal = 0.40. No error. |
| Graph query returns empty | No suggestion displayed. Task saved as-is. |
| Intent = unknown | ContextRouter returns empty strategy → Factory uses current behavior. |
| `--refine` + no suggestion | Interactive mode falls back to saving original. |

---

## Known limitations

**KNOWN_LIMITATION_01 — Intent cache is process-scoped.**  
`sync.Map` cache lives for the duration of the process. If task description changes between runs (user edits), next run re-classifies. Accepted for MVP.

**KNOWN_LIMITATION_02 — VaguenessScore weights are uncalibrated.**  
Values chosen by heuristic reasoning, not empirical data. Named constants in code enable fast tuning without refactoring. Target: calibrate after 20+ real task descriptions.

**KNOWN_LIMITATION_03 — isatty not detected in --refine mode.**  
If `--refine` is passed in a non-interactive context (piped input), the prompt will hang. Mitigation: document in CLI help text. Proper fix: `golang.org/x/term` check in a future iteration.

---

## Audit Depth Rule (process documentation)

Defined during this design session. Applies to all future sentinel feature design:

```
Audit Rounds = f(BlastRadius, Reversibility)

              │ Easy to revert  │ Hard to revert
──────────────┼─────────────────┼────────────────
Low blast     │   0 rounds      │   1 round
Medium blast  │   1 round       │   2 rounds
High blast    │   1 round       │   2 rounds + sign-off
```

**Termination condition (overrides matrix):**  
If audit round N finds 0 critical and 0 major issues → stop.

**Time budget:**  
If design time > 30% of estimated implementation time → commit design and implement. Implementation reveals problems faster than additional design discussion.

---

## Future: Subsystem C

Generic prompt refinement for non-sentinel AI conversations. Excluded from MVP because without graph grounding, refinement is templateization without signal gain. Revisit when knowledge base (`~/knowledge/`) has sufficient domain documents to serve as grounding context.

---

## Verification gate

```bash
go build ./...
go test ./internal/bridge/... ./internal/intake/...
go vet ./...
```

All must pass before PR.
