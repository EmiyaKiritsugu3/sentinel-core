# Prompt Intelligence System — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add intent-aware context routing (Subsystem B) and input disambiguation (Subsystem A) to sentinel's AI pipeline.

**Architecture:** Two independent subsystems operating at different points in time. B intercepts `GeneratePayload` at agent execution time to select context based on task intent. A intercepts `sentinel plan` at CLI time to detect vague descriptions and suggest graph-anchored alternatives.

**Tech Stack:** Go 1.26, SQLite (modernc), Gemini API (google/generative-ai-go), Cobra CLI, standard library only (no new deps).

**Spec:** `docs/superpowers/specs/2026-05-04-prompt-intelligence-design.md`

---

> **Note:** Subsystems B and A are fully independent. B (Tasks 1–4) can be merged before A (Tasks 5–6) is started. Each block produces working, testable software on its own.

---

## File Map

**Create:**
- `internal/bridge/classifier.go` — Intent type, IntentClassifier, AIClassifier interface, NilClassifier
- `internal/bridge/classifier_test.go` — heuristic + cache tests
- `internal/bridge/gemini_classifier.go` — GeminiClassifier (Gemini API implementation)
- `internal/bridge/router.go` — ContextStrategy, strategyByIntent, StrategyFor()
- `internal/bridge/router_test.go` — strategy mapping tests
- `internal/intake/disambiguator.go` — Disambiguator, VaguenessScore, Suggestion
- `internal/intake/disambiguator_test.go` — score + suggestion tests

**Modify:**
- `internal/bridge/prompt_factory.go` — add classifier field, update NewFactory, add loadContextByStrategy
- `internal/agents/engine.go` — wire GeminiClassifier + IntentClassifier into NewEngine, pass to NewFactory
- `cmd/sentinel/commands/start.go:34` — update NewFactory call signature
- `internal/agents/engine_test.go:21` — update NewFactory with NilClassifier
- `cmd/sentinel/commands/plan.go` — add --refine / --no-suggest flags + Disambiguator call

---

## BLOCK B — Smart Context Routing

---

### Task 1: Intent types, heuristic classifier, NilClassifier

**Files:**
- Create: `internal/bridge/classifier.go`
- Create: `internal/bridge/classifier_test.go`

- [ ] **Step 1.1: Write failing tests**

```go
// internal/bridge/classifier_test.go
package bridge_test

import (
	"context"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
)

func TestHeuristic_Diagnose(t *testing.T) {
	c := bridge.NewIntentClassifier(nil, 0.60)
	got := c.Classify(context.Background(), "t1", "fix the broken JWT validation")
	if got != bridge.IntentDiagnose {
		t.Errorf("want diagnose, got %s", got)
	}
}

func TestHeuristic_Implement(t *testing.T) {
	c := bridge.NewIntentClassifier(nil, 0.60)
	got := c.Classify(context.Background(), "t2", "add OAuth2 support to auth module")
	if got != bridge.IntentImplement {
		t.Errorf("want implement, got %s", got)
	}
}

func TestHeuristic_Ambiguous_ReturnsUnknown(t *testing.T) {
	// "fix" (diagnose) + "review" (review) = 2 categories = confidence 0.30 < 0.60
	// AI is nil → IntentUnknown
	c := bridge.NewIntentClassifier(nil, 0.60)
	got := c.Classify(context.Background(), "t3", "fix and review the auth module")
	if got != bridge.IntentUnknown {
		t.Errorf("want unknown for ambiguous+nil AI, got %s", got)
	}
}

func TestHeuristic_NoMatch_ReturnsUnknown(t *testing.T) {
	c := bridge.NewIntentClassifier(nil, 0.60)
	got := c.Classify(context.Background(), "t4", "the JWT module")
	if got != bridge.IntentUnknown {
		t.Errorf("want unknown for no match, got %s", got)
	}
}

func TestCache_ReturnsCachedOnSecondCall(t *testing.T) {
	c := bridge.NewIntentClassifier(nil, 0.60)
	first := c.Classify(context.Background(), "t5", "fix the bug")
	second := c.Classify(context.Background(), "t5", "completely different description")
	if first != second {
		t.Errorf("want cache hit (same intent), got %s vs %s", first, second)
	}
}

func TestNilClassifier_ReturnsUnknown(t *testing.T) {
	n := bridge.NewNilClassifier()
	got, err := n.Classify(context.Background(), "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bridge.IntentUnknown {
		t.Errorf("want unknown, got %s", got)
	}
}
```

- [ ] **Step 1.2: Run tests — verify they fail**

```bash
cd /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core
go test ./internal/bridge/... 2>&1 | head -20
```

Expected: `cannot find package` or `undefined: bridge.IntentDiagnose`

- [ ] **Step 1.3: Create classifier.go**

```go
// internal/bridge/classifier.go
package bridge

import (
	"context"
	"strings"
	"sync"
)

type Intent string

const (
	IntentDiagnose  Intent = "diagnose"
	IntentImplement Intent = "implement"
	IntentRefactor  Intent = "refactor"
	IntentReview    Intent = "review"
	IntentUnknown   Intent = "unknown"
)

var intentKeywords = map[Intent][]string{
	IntentDiagnose:  {"fix", "bug", "error", "broken", "failing", "crash", "debug", "investigate", "corrigir", "erro"},
	IntentImplement: {"add", "create", "build", "implement", "new", "adicionar", "criar", "implementar"},
	IntentRefactor:  {"refactor", "cleanup", "reorganize", "extract", "move", "simplify", "refatorar"},
	IntentReview:    {"review", "audit", "check", "verify", "analyze", "validate", "revisar", "auditar"},
}

// AIClassifier is the interface for AI-powered intent classification.
// The zero value (nil) means heuristic-only mode.
type AIClassifier interface {
	Classify(ctx context.Context, description string) (Intent, error)
}

// IntentClassifier classifies task intent using a tiered strategy:
// heuristic first, AI fallback when confidence is below threshold.
type IntentClassifier struct {
	ai        AIClassifier
	threshold float64
	cache     sync.Map // taskID → Intent, goroutine-safe
}

func NewIntentClassifier(ai AIClassifier, threshold float64) *IntentClassifier {
	return &IntentClassifier{ai: ai, threshold: threshold}
}

// Classify returns the Intent for a task. Results are cached by taskID.
func (c *IntentClassifier) Classify(ctx context.Context, taskID, description string) Intent {
	if v, ok := c.cache.Load(taskID); ok {
		return v.(Intent)
	}
	intent, confidence := heuristicClassify(description)
	if confidence < c.threshold && c.ai != nil {
		if aiIntent, err := c.ai.Classify(ctx, description); err == nil {
			intent = aiIntent
		} else {
			fmt.Fprintf(os.Stderr, "warning: classifier: gemini fallback failed: %v\n", err)
		}
	}
	c.cache.Store(taskID, intent)
	return intent
}

func heuristicClassify(description string) (Intent, float64) {
	lower := strings.ToLower(description)
	words := strings.Fields(lower)

	hits := map[Intent]int{}
	for _, word := range words {
		for intent, keywords := range intentKeywords {
			for _, kw := range keywords {
				if strings.Contains(word, kw) {
					hits[intent]++
				}
			}
		}
	}

	categoriesHit := 0
	var bestIntent Intent
	bestCount := 0
	for intent, count := range hits {
		if count > 0 {
			categoriesHit++
		}
		if count > bestCount {
			bestCount = count
			bestIntent = intent
		}
	}

	switch categoriesHit {
	case 0:
		return IntentUnknown, 0.00
	case 1:
		return bestIntent, 0.85
	default:
		return bestIntent, 0.30
	}
}

// NilClassifier is a null object for AIClassifier. Use in tests and
// when no AI key is configured.
type NilClassifier struct{}

func NewNilClassifier() *NilClassifier { return &NilClassifier{} }

func (n *NilClassifier) Classify(_ context.Context, _ string) (Intent, error) {
	return IntentUnknown, nil
}
```

Add missing imports at the top of classifier.go:

```go
import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)
```

- [ ] **Step 1.4: Run tests — verify they pass**

```bash
go test ./internal/bridge/... -run TestHeuristic -run TestCache -run TestNilClassifier -v
```

Expected: all PASS

- [ ] **Step 1.5: Commit**

```bash
git add internal/bridge/classifier.go internal/bridge/classifier_test.go
git commit -m "feat(bridge): add IntentClassifier with heuristic+cache and NilClassifier"
```

---

### Task 2: GeminiClassifier

**Files:**
- Create: `internal/bridge/gemini_classifier.go`

No unit test for GeminiClassifier — it calls a live API. Covered by integration test in Task 4.

- [ ] **Step 2.1: Create gemini_classifier.go**

```go
// internal/bridge/gemini_classifier.go
package bridge

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// GeminiClassifier implements AIClassifier using the Gemini API.
type GeminiClassifier struct {
	client *genai.Client
}

func NewGeminiClassifier(client *genai.Client) *GeminiClassifier {
	return &GeminiClassifier{client: client}
}

func (g *GeminiClassifier) Classify(ctx context.Context, description string) (Intent, error) {
	model := g.client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text("You are a task classifier. Respond with exactly one word.")},
	}
	prompt := fmt.Sprintf(
		"Classify this software task into exactly one word: diagnose, implement, refactor, or review.\nTask: %s",
		description,
	)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return IntentUnknown, fmt.Errorf("gemini classifier: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return IntentUnknown, nil
	}
	raw := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	parsed := Intent(strings.ToLower(strings.TrimSpace(raw)))
	switch parsed {
	case IntentDiagnose, IntentImplement, IntentRefactor, IntentReview:
		return parsed, nil
	default:
		return IntentUnknown, nil
	}
}
```

- [ ] **Step 2.2: Verify build**

```bash
go build ./internal/bridge/...
```

Expected: exits 0, no output.

- [ ] **Step 2.3: Commit**

```bash
git add internal/bridge/gemini_classifier.go
git commit -m "feat(bridge): add GeminiClassifier as AIClassifier implementation"
```

---

### Task 3: ContextStrategy and StrategyFor

**Files:**
- Create: `internal/bridge/router.go`
- Create: `internal/bridge/router_test.go`

- [ ] **Step 3.1: Write failing tests**

```go
// internal/bridge/router_test.go
package bridge_test

import (
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
)

func TestStrategyFor_KnownIntents_HaveNodeLimit(t *testing.T) {
	for _, intent := range []bridge.Intent{
		bridge.IntentDiagnose,
		bridge.IntentImplement,
		bridge.IntentRefactor,
		bridge.IntentReview,
	} {
		s := bridge.StrategyFor(intent)
		if s.NodeLimit == 0 {
			t.Errorf("intent %s: NodeLimit must be > 0", intent)
		}
	}
}

func TestStrategyFor_Unknown_ReturnsEmpty(t *testing.T) {
	s := bridge.StrategyFor(bridge.IntentUnknown)
	if s.NodeLimit != 0 || s.HighCoupling || s.IncludeTests || s.IncludeADRs {
		t.Error("IntentUnknown must return zero-value ContextStrategy")
	}
}

func TestStrategyFor_Diagnose_HasHighCoupling(t *testing.T) {
	s := bridge.StrategyFor(bridge.IntentDiagnose)
	if !s.HighCoupling {
		t.Error("diagnose strategy must include high coupling nodes")
	}
}

func TestStrategyFor_Implement_HasTests(t *testing.T) {
	s := bridge.StrategyFor(bridge.IntentImplement)
	if !s.IncludeTests {
		t.Error("implement strategy must include test files")
	}
}
```

- [ ] **Step 3.2: Run tests — verify they fail**

```bash
go test ./internal/bridge/... -run TestStrategyFor -v 2>&1 | head -10
```

Expected: `undefined: bridge.StrategyFor`

- [ ] **Step 3.3: Create router.go**

```go
// internal/bridge/router.go
package bridge

// ContextStrategy defines what context to inject into the AI payload
// based on the classified task intent.
type ContextStrategy struct {
	// DB-queryable fields (nodes + edges tables)
	HighCoupling  bool // nodes with highest fan-in via edges COUNT(*)
	RecentChanges bool // weight toward higher last_indexed
	IncludeTests  bool // nodes where file_path LIKE '%_test.go'
	NodeLimit     int  // max nodes to inject (0 = use current default)

	// File-based fields (direct filesystem read)
	IncludeADRs        bool // reads docs/architecture/adr/*.md
	IncludeDebtMarkers bool // reads TECHNICAL-DEBT.md, filters by task keywords
}

var strategyByIntent = map[Intent]ContextStrategy{
	IntentDiagnose: {
		HighCoupling:  true,
		RecentChanges: true,
		NodeLimit:     15,
	},
	IntentImplement: {
		IncludeTests: true,
		IncludeADRs:  true,
		NodeLimit:    10,
	},
	IntentRefactor: {
		HighCoupling:       true,
		IncludeDebtMarkers: true,
		NodeLimit:          12,
	},
	IntentReview: {
		IncludeADRs: true,
		NodeLimit:   8,
	},
	IntentUnknown: {}, // zero value → Factory uses existing default behavior
}

// StrategyFor returns the ContextStrategy for a given intent.
// IntentUnknown returns a zero-value strategy (no routing).
func StrategyFor(intent Intent) ContextStrategy {
	return strategyByIntent[intent]
}
```

- [ ] **Step 3.4: Run tests — verify they pass**

```bash
go test ./internal/bridge/... -run TestStrategyFor -v
```

Expected: all PASS

- [ ] **Step 3.5: Commit**

```bash
git add internal/bridge/router.go internal/bridge/router_test.go
git commit -m "feat(bridge): add ContextStrategy and StrategyFor router"
```

---

### Task 4: Wire classifier into Factory and Engine

**Files:**
- Modify: `internal/bridge/prompt_factory.go`
- Modify: `internal/agents/engine.go`
- Modify: `cmd/sentinel/commands/start.go`
- Modify: `internal/agents/engine_test.go`

- [ ] **Step 4.1: Update prompt_factory.go — add classifier, update NewFactory, add loadContextByStrategy**

In `internal/bridge/prompt_factory.go`, replace the `Factory` struct and `NewFactory`:

```go
// Replace:
type Factory struct {
	db *sqlite.DB
}

func NewFactory(db *sqlite.DB) *Factory {
	return &Factory{db: db}
}

// With:
type Factory struct {
	db         *sqlite.DB
	classifier *IntentClassifier
}

func NewFactory(db *sqlite.DB, classifier *IntentClassifier) *Factory {
	return &Factory{db: db, classifier: classifier}
}
```

In `GeneratePayload`, add intent classification before building the payload. Replace the call to `f.loadSurgicalContext(taskID)` with:

```go
// After retrieving task and before building systemOut:
intent := IntentUnknown
if f.classifier != nil {
    intent = f.classifier.Classify(ctx, taskID, task.Description)
}
strategy := StrategyFor(intent)
nodes, err := f.loadContextByStrategy(taskID, strategy)
if err != nil {
    return nil, fmt.Errorf("bridge: failed to load context: %w", err)
}
```

Add `loadContextByStrategy` below `loadSurgicalContext` (keep old function — it becomes the fallback):

```go
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
		debtContent, _ := extractLines("TECHNICAL-DEBT.md", 1, 100)
		if debtContent != "" {
			nodes = append(nodes, ContextNode{
				Name:        "TECHNICAL-DEBT",
				Type:        "doc",
				FilePath:    "TECHNICAL-DEBT.md",
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
```

Also add `"os"` and `"strings"` to the import block if not already present.

Update `GeneratePayload` signature to accept `ctx context.Context` as first parameter:

```go
// Replace:
func (f *Factory) GeneratePayload(taskID string, personaPrompt string) (*ContextPayload, error) {

// With:
func (f *Factory) GeneratePayload(ctx context.Context, taskID string, personaPrompt string) (*ContextPayload, error) {
```

- [ ] **Step 4.2: Update engine.go — wire classifier into NewEngine**

In `internal/agents/engine.go`, inside `NewEngine`, after `client` is created:

```go
// After: client, err := genai.NewClient(...)
geminiClassifier := bridge.NewGeminiClassifier(client)
classifier := bridge.NewIntentClassifier(geminiClassifier, 0.60)
```

Replace `factory` parameter usage when stored:

```go
// The factory parameter is now unused — NewFactory is called by the caller.
// Remove factory from NewEngine parameters and construct it internally:
// OLD signature: NewEngine(r *Registry, auth AuthProvider, factory *bridge.Factory, v *reflect.Validator, db *sqlite.DB)
// NEW signature: NewEngine(r *Registry, auth AuthProvider, v *reflect.Validator, db *sqlite.DB)
```

Inside `NewEngine`, construct factory internally:

```go
factory := bridge.NewFactory(db, classifier)
```

Update the `Engine` struct initialization accordingly:

```go
return &Engine{
    Registry:      r,
    genaiClient:   client,
    authProvider:  auth,
    promptFactory: factory,
    validator:     v,
    Dispatcher:    nil,
    DB:            db,
}, nil
```

Also update all calls to `GeneratePayload` inside engine.go to pass `ctx` as first argument.

- [ ] **Step 4.3: Update start.go — fix call site**

In `cmd/sentinel/commands/start.go:34`, remove the standalone `NewFactory` call (factory is now constructed inside `NewEngine`):

```go
// Remove this line:
factory := bridge.NewFactory(db)

// And remove factory from the NewEngine call:
// OLD:
engine, err := agents.NewEngine(registry, authProvider, factory, validator, db)
// NEW:
engine, err := agents.NewEngine(registry, authProvider, validator, db)
```

- [ ] **Step 4.4: Update engine_test.go — replace NewFactory(nil) with NilClassifier**

In `internal/agents/engine_test.go:21`:

```go
// OLD:
factory := bridge.NewFactory(nil)

// NEW:
factory := bridge.NewFactory(nil, bridge.NewNilClassifier())
```

- [ ] **Step 4.5: Build and test**

```bash
go build ./...
go test ./internal/bridge/... ./internal/agents/... -v 2>&1 | tail -20
```

Expected: build succeeds, all existing tests pass.

- [ ] **Step 4.6: Commit**

```bash
git add internal/bridge/prompt_factory.go internal/agents/engine.go \
        cmd/sentinel/commands/start.go internal/agents/engine_test.go
git commit -m "feat(bridge): wire IntentClassifier into Factory and Engine for smart context routing"
```

---

## BLOCK A — Input Disambiguation

---

### Task 5: Disambiguator and VaguenessScore

**Files:**
- Create: `internal/intake/disambiguator.go`
- Create: `internal/intake/disambiguator_test.go`

- [ ] **Step 5.1: Write failing tests**

```go
// internal/intake/disambiguator_test.go
package intake_test

import (
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/intake"
)

func TestVaguenessScore_HighForShortGeneric(t *testing.T) {
	d := intake.NewDisambiguator(nil) // nil db = skip graph phase
	score := d.VaguenessScore("fix bug")
	if score <= 0.50 {
		t.Errorf("want score > 0.50 for 'fix bug', got %.2f", score)
	}
}

func TestVaguenessScore_LowForPrecise(t *testing.T) {
	d := intake.NewDisambiguator(nil)
	score := d.VaguenessScore("fix JWT validation in internal/agents/auth_provider.go")
	if score > 0.50 {
		t.Errorf("want score <= 0.50 for precise description, got %.2f", score)
	}
}

func TestVaguenessScore_LowForLongDescriptive(t *testing.T) {
	d := intake.NewDisambiguator(nil)
	score := d.VaguenessScore("refactor loadSurgicalContext to use graph-aware ranking based on edge count")
	if score > 0.50 {
		t.Errorf("want score <= 0.50 for long descriptive, got %.2f", score)
	}
}

func TestAnalyze_NotVague_NoSuggestions(t *testing.T) {
	d := intake.NewDisambiguator(nil)
	vague, suggestions := d.Analyze("fix JWT validation in internal/agents/auth_provider.go")
	if vague {
		t.Error("want not vague for precise description")
	}
	if len(suggestions) != 0 {
		t.Error("want no suggestions for non-vague description")
	}
}

func TestAnalyze_Vague_NilDB_ReturnsSuggestions(t *testing.T) {
	d := intake.NewDisambiguator(nil)
	vague, _ := d.Analyze("fix bug")
	if !vague {
		t.Error("want vague=true for 'fix bug'")
	}
}
```

- [ ] **Step 5.2: Run tests — verify they fail**

```bash
go test ./internal/intake/... 2>&1 | head -10
```

Expected: `cannot find package` or `no Go files`

- [ ] **Step 5.3: Create disambiguator.go**

```go
// internal/intake/disambiguator.go
package intake

import (
	"fmt"
	"math"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

const (
	weightLength   = 0.25
	weightVerb     = 0.20
	weightPronoun  = 0.15
	weightAnchor   = 0.40
	scoreThreshold = 0.50
)

var genericVerbs = []string{
	"fix", "improve", "update", "change", "make", "handle", "check",
	"corrigir", "melhorar", "atualizar", "mudar",
}

var vaguePronouns = []string{
	"it", "this", "the issue", "the bug", "the problem",
	"isso", "ele", "o problema", "o erro",
}

// Suggestion is a graph-anchored alternative for a vague task description.
type Suggestion struct {
	NodeName string
	FilePath string
}

// Disambiguator analyzes task descriptions for vagueness and suggests
// graph-anchored alternatives.
type Disambiguator struct {
	db *sqlite.DB // nil = skip graph phase (Phase 2)
}

func NewDisambiguator(db *sqlite.DB) *Disambiguator {
	return &Disambiguator{db: db}
}

// Analyze returns whether the description is vague and any graph suggestions.
func (d *Disambiguator) Analyze(description string) (vague bool, suggestions []Suggestion) {
	score := d.VaguenessScore(description)
	if score <= scoreThreshold {
		return false, nil
	}
	if d.db != nil {
		suggestions = d.queryGraph(description)
	}
	return true, suggestions
}

// VaguenessScore returns a score in [0.0, 1.0]. Values > 0.50 trigger suggestion.
func (d *Disambiguator) VaguenessScore(description string) float64 {
	score := lengthSignal(description) +
		verbSignal(description) +
		pronounSignal(description) +
		d.anchorSignal(description)
	return math.Min(score, 1.0)
}

func lengthSignal(description string) float64 {
	n := len(strings.Fields(description))
	switch {
	case n < 3:
		return weightLength // 0.25
	case n <= 5:
		return 0.18
	case n <= 10:
		return 0.08
	default:
		return 0.00
	}
}

func verbSignal(description string) float64 {
	lower := strings.ToLower(description)
	for _, v := range genericVerbs {
		if strings.Contains(lower, v) {
			return weightVerb // 0.20
		}
	}
	return 0.00
}

func pronounSignal(description string) float64 {
	lower := strings.ToLower(description)
	for _, p := range vaguePronouns {
		if strings.Contains(lower, p) {
			return weightPronoun // 0.15
		}
	}
	return 0.00
}

func (d *Disambiguator) anchorSignal(description string) float64 {
	lower := strings.ToLower(description)

	// Phase 1: lexical anchors (zero DB)
	if strings.Contains(lower, "internal/") ||
		strings.Contains(lower, "pkg/") ||
		strings.Contains(lower, ".go") {
		return 0.00
	}
	// line reference: colon followed by digit
	for i, ch := range description {
		if ch == ':' && i+1 < len(description) && description[i+1] >= '0' && description[i+1] <= '9' {
			return 0.00
		}
	}

	// Phase 2: graph-anchored (DB query)
	if d.db == nil {
		return weightAnchor // 0.40 — no graph available
	}

	var count int
	if err := d.db.Conn.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count); err != nil || count == 0 {
		return weightAnchor // graph not indexed
	}

	keywords := extractKeywords(description)
	if len(keywords) == 0 {
		return weightAnchor
	}

	matched := 0
	for _, kw := range keywords {
		var n int
		_ = d.db.Conn.QueryRow(
			"SELECT COUNT(*) FROM nodes WHERE LOWER(name) LIKE ?",
			fmt.Sprintf("%%%s%%", strings.ToLower(kw)),
		).Scan(&n)
		if n > 0 {
			matched++
		}
	}

	matchedRatio := float64(matched) / float64(len(keywords))
	return weightAnchor * (1.0 - matchedRatio)
}

func (d *Disambiguator) queryGraph(description string) []Suggestion {
	keywords := extractKeywords(description)
	var suggestions []Suggestion
	seen := map[string]bool{}

	for _, kw := range keywords {
		rows, err := d.db.Conn.Query(
			"SELECT name, file_path FROM nodes WHERE LOWER(name) LIKE ? LIMIT 3",
			fmt.Sprintf("%%%s%%", strings.ToLower(kw)),
		)
		if err != nil {
			continue
		}
		for rows.Next() {
			var s Suggestion
			if err := rows.Scan(&s.NodeName, &s.FilePath); err == nil && !seen[s.NodeName] {
				suggestions = append(suggestions, s)
				seen[s.NodeName] = true
			}
		}
		rows.Close()
		if len(suggestions) >= 5 {
			break
		}
	}
	return suggestions
}

func extractKeywords(description string) []string {
	lower := strings.ToLower(description)
	// Remove common stop words and return remaining tokens >= 3 chars
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "in": true, "of": true,
		"to": true, "and": true, "or": true, "for": true, "with": true,
		"fix": true, "add": true, "new": true, "o": true, "a": true,
	}
	var keywords []string
	for _, w := range strings.Fields(lower) {
		w = strings.Trim(w, ".,!?")
		if len(w) >= 3 && !stopWords[w] {
			keywords = append(keywords, w)
		}
	}
	return keywords
}
```

- [ ] **Step 5.4: Run tests — verify they pass**

```bash
go test ./internal/intake/... -v
```

Expected: all PASS

- [ ] **Step 5.5: Build check**

```bash
go build ./...
```

Expected: exits 0.

- [ ] **Step 5.6: Commit**

```bash
git add internal/intake/disambiguator.go internal/intake/disambiguator_test.go
git commit -m "feat(intake): add Disambiguator with VaguenessScore and graph-anchored suggestions"
```

---

### Task 6: CLI flags — sentinel plan --refine / --no-suggest

**Files:**
- Modify: `cmd/sentinel/commands/plan.go`

- [ ] **Step 6.1: Read current plan.go**

```bash
cat cmd/sentinel/commands/plan.go
```

Understand the current `RunE` function structure before modifying.

- [ ] **Step 6.2: Add flags and disambiguation logic to plan.go**

At the top of the file, add the import for `intake` and `fmt`:

```go
import (
    // existing imports...
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/EmiyaKiritsugu3/sentinel-core/internal/intake"
)
```

Add flag variables at package level (above `init()`):

```go
var (
    flagRefine    bool
    flagNoSuggest bool
)
```

In `init()` (or equivalent `PersistentPreRun`), register the flags:

```go
planCmd.Flags().BoolVarP(&flagRefine,    "refine",     "r", false, "interactive disambiguation before saving")
planCmd.Flags().BoolVar (&flagNoSuggest, "no-suggest", false,      "skip suggestion output (for scripts and CI)")
```

At the start of `RunE`, before saving the task, add:

```go
// Flag conflict: --refine takes precedence
if flagRefine && flagNoSuggest {
    flagNoSuggest = false
}

if !flagNoSuggest {
    disambiguator := intake.NewDisambiguator(db)
    vague, suggestions := disambiguator.Analyze(description)

    if vague && len(suggestions) > 0 {
        if flagRefine {
            // Interactive mode: show options, prompt user
            fmt.Println("[SUGGEST] Task description may be vague. Did you mean:")
            for i, s := range suggestions {
                fmt.Printf("  [%d] %s  (%s)\n", i+1, s.NodeName, s.FilePath)
            }
            fmt.Printf("  [0] Keep original: %q\n", description)
            fmt.Print("Choice [0]: ")

            scanner := bufio.NewScanner(os.Stdin)
            if scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                if idx := parseChoice(line, len(suggestions)); idx > 0 {
                    description = suggestions[idx-1].NodeName + " — " + suggestions[idx-1].FilePath
                }
            }
        } else {
            // Default mode: print suggestion, save original
            fmt.Printf("[SUGGEST] did you mean: %s in %s?\n",
                suggestions[0].NodeName, suggestions[0].FilePath)
        }
    }
}
// continue with saving task using description (original or chosen)
```

Add helper at the bottom of the file:

```go
func parseChoice(s string, max int) int {
    if s == "" || s == "0" {
        return 0
    }
    n := 0
    for _, ch := range s {
        if ch < '0' || ch > '9' {
            return 0
        }
        n = n*10 + int(ch-'0')
    }
    if n < 1 || n > max {
        return 0
    }
    return n
}
```

- [ ] **Step 6.3: Build check**

```bash
go build ./...
```

Expected: exits 0.

- [ ] **Step 6.4: Manual smoke test**

```bash
# Build the binary
go build -o /tmp/sentinel-test ./cmd/sentinel

# Test default mode (should print [SUGGEST] if graph has matching nodes)
/tmp/sentinel-test plan "fix bug" --no-suggest
# Expected: task created, no [SUGGEST] output

# Test --no-suggest suppresses output
/tmp/sentinel-test plan "fix bug" --no-suggest
# Expected: task created silently
```

- [ ] **Step 6.5: Commit**

```bash
git add cmd/sentinel/commands/plan.go
git commit -m "feat(cli): add --refine and --no-suggest flags to sentinel plan"
```

---

## Final verification

- [ ] **Full build and test suite**

```bash
go build ./...
go test ./...
go vet ./...
```

Expected: all pass, zero vet warnings.

- [ ] **Run existing sentinel commands to verify no regression**

```bash
go build -o /tmp/sentinel ./cmd/sentinel
/tmp/sentinel scan
/tmp/sentinel status
```

Expected: commands work as before.

- [ ] **Final commit if any loose changes**

```bash
git status
# If clean, nothing to do.
```

---

## Verification gate (from spec)

```bash
go build ./...
go test ./internal/bridge/... ./internal/intake/...
go vet ./...
```

All must pass before opening PR.
