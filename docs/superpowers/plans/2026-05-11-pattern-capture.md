# Pattern Capture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `patterns` table + FTS5 index to graph.db, with CRUD Go package, CLI commands, backfill from existing docs, and Epiphany Protocol Filtro D extension.

**Architecture:** New `internal/patterns/` package with `PatternStore` receiving `*sqlite.DB` via injection. Schema migration in `internal/graph/schema.go`. CLI subcommands registered via existing `internal/registry` pattern. FTS5 external-content virtual table with triggers for full-text search. Levenshtein dedup on `pattern add`.

**Tech Stack:** Go 1.21+, SQLite (modernc.org/sqlite), cobra CLI, google/uuid

---

## File Structure

| Action | Path | Responsibility |
|---|---|---|
| Modify | `internal/graph/schema.go` | Add patterns table + FTS5 + triggers to schema const + pragmaTableInfo |
| Create | `internal/patterns/store.go` | Pattern struct, PatternStore, CRUD methods (Create, List, Search, Get, FindSimilar) |
| Create | `internal/patterns/store_test.go` | Unit tests for all PatternStore methods |
| Create | `internal/patterns/backfill.go` | BackfillFromCognitiveDNA, BackfillFromEvolutionInsights, BackfillFromSentinelLog (dry-run) |
| Create | `internal/patterns/backfill_test.go` | Tests for backfill parsing + idempotency |
| Create | `internal/patterns/dedup.go` | Levenshtein distance + tag overlap dedup logic |
| Create | `internal/patterns/dedup_test.go` | Tests for Levenshtein and FindSimilar |
| Create | `cmd/sentinel/commands/pattern.go` | sentinel pattern {add,list,search,get,backfill} CLI subcommands |
| Modify | `GEMINI.md` | Add Filtro D to Epiphany Protocol section |

---

### Task 1: Schema Migration — patterns table + FTS5

**Files:**
- Modify: `internal/graph/schema.go:10-117` (schema const + Migrate + pragmaTableInfo)
- Test: `internal/graph/schema_test.go`

- [ ] **Step 1: Add patterns table DDL to schema const**

In `internal/graph/schema.go`, append to the `schema` const (after the `agent_trust` table, before the closing backtick):

```go
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

CREATE VIRTUAL TABLE IF NOT EXISTS patterns_fts USING fts5(
    title,
    description,
    tags,
    content=patterns,
    content_rowid=rowid
);

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

- [ ] **Step 2: Add patterns to pragmaTableInfo map**

In `internal/graph/schema.go`, add entry to the `pragmaTableInfo` map (after the `performance_logs` entry):

```go
"patterns": "PRAGMA table_info(patterns)",
```

- [ ] **Step 3: Run existing tests to verify migration doesn't break anything**

Run: `go test ./internal/graph/... -v -run TestMigrate`

Expected: All existing tests PASS. The new table is created alongside existing ones.

- [ ] **Step 4: Add a test for the patterns table creation**

In `internal/graph/schema_test.go`, add to the `tables` slice in `TestMigrate`:

```go
tables := []string{
    "specialist_registry",
    "sub_tasks",
    "performance_logs",
    "agent_trust",
    "patterns",
}
```

Also add a FTS5 check:

```go
// Verify FTS5 virtual table exists
var ftsName string
err = sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='patterns_fts'").Scan(&ftsName)
if err != nil {
    t.Errorf("patterns_fts virtual table was not created: %v", err)
}
```

Run: `go test ./internal/graph/... -v -run TestMigrate`

Expected: PASS with patterns table and patterns_fts created.

- [ ] **Step 5: Commit**

```bash
git add internal/graph/schema.go internal/graph/schema_test.go
git commit -m "feat(patterns): add patterns table + FTS5 index to schema migration"
```

---

### Task 2: PatternStore — CRUD + Search

**Files:**
- Create: `internal/patterns/store.go`
- Create: `internal/patterns/store_test.go`

- [ ] **Step 1: Write failing tests for PatternStore**

Create `internal/patterns/store_test.go`:

```go
package patterns

import (
    "testing"

    "github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
)

func setupStore(t *testing.T) *PatternStore {
    t.Helper()
    db := testutil.SetupTestDB(t)
    t.Cleanup(func() { db.Close() })
    if err := graph.Migrate(db); err != nil {
        t.Fatalf("migration failed: %v", err)
    }
    store, err := NewPatternStore(db)
    if err != nil {
        t.Fatalf("NewPatternStore failed: %v", err)
    }
    return store
}

func TestNewPatternStore_NilDB(t *testing.T) {
    _, err := NewPatternStore(nil)
    if err == nil {
        t.Fatal("expected error for nil db")
    }
}

func TestCreate(t *testing.T) {
    store := setupStore(t)

    id, err := store.Create(&Pattern{
        Title:       "Diagnóstico sem dado empírico = loop",
        Description: "Quando o agente diagnostica sem dados empíricos, entra em loop de hipóteses",
        Category:    "anti-pattern",
        Source:      "manual",
        Tags:        "loop,diagnosis,empirical",
        Impact:      "high",
    })
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }
    if id == "" {
        t.Fatal("expected non-empty id")
    }
}

func TestCreate_InvalidCategory(t *testing.T) {
    store := setupStore(t)

    _, err := store.Create(&Pattern{
        Title:       "Test",
        Description: "Test",
        Category:    "invalid-category",
        Source:      "manual",
        Tags:        "",
        Impact:      "medium",
    })
    if err == nil {
        t.Fatal("expected error for invalid category")
    }
}

func TestCreate_InvalidSource(t *testing.T) {
    store := setupStore(t)

    _, err := store.Create(&Pattern{
        Title:       "Test",
        Description: "Test",
        Category:    "anti-pattern",
        Source:      "invalid-source",
        Tags:        "",
        Impact:      "medium",
    })
    if err == nil {
        t.Fatal("expected error for invalid source")
    }
}

func TestList_NoFilters(t *testing.T) {
    store := setupStore(t)

    store.Create(&Pattern{
        Title: "P1", Description: "D1", Category: "anti-pattern",
        Source: "manual", Tags: "a", Impact: "high",
    })
    store.Create(&Pattern{
        Title: "P2", Description: "D2", Category: "cognitive-pattern",
        Source: "epiphany", Tags: "b", Impact: "low",
    })

    patterns, err := store.List(ListFilters{})
    if err != nil {
        t.Fatalf("List failed: %v", err)
    }
    if len(patterns) != 2 {
        t.Fatalf("expected 2 patterns, got %d", len(patterns))
    }
}

func TestList_FilterByCategory(t *testing.T) {
    store := setupStore(t)

    store.Create(&Pattern{
        Title: "P1", Description: "D1", Category: "anti-pattern",
        Source: "manual", Tags: "a", Impact: "high",
    })
    store.Create(&Pattern{
        Title: "P2", Description: "D2", Category: "cognitive-pattern",
        Source: "epiphany", Tags: "b", Impact: "low",
    })

    patterns, err := store.List(ListFilters{Category: "anti-pattern"})
    if err != nil {
        t.Fatalf("List failed: %v", err)
    }
    if len(patterns) != 1 {
        t.Fatalf("expected 1 pattern, got %d", len(patterns))
    }
    if patterns[0].Title != "P1" {
        t.Fatalf("expected P1, got %s", patterns[0].Title)
    }
}

func TestGet(t *testing.T) {
    store := setupStore(t)

    id, _ := store.Create(&Pattern{
        Title: "Test Pattern", Description: "Full description here",
        Category: "anti-pattern", Source: "cognitive-dna",
        SourceRef: "COGNITIVE-DNA.md:AP-01", Tags: "test", Impact: "medium",
    })

    p, err := store.Get(id)
    if err != nil {
        t.Fatalf("Get failed: %v", err)
    }
    if p.Title != "Test Pattern" {
        t.Fatalf("expected 'Test Pattern', got %s", p.Title)
    }
    if p.SourceRef != "COGNITIVE-DNA.md:AP-01" {
        t.Fatalf("expected source_ref, got %s", p.SourceRef)
    }
}

func TestGet_NotFound(t *testing.T) {
    store := setupStore(t)

    _, err := store.Get("nonexistent-id")
    if err == nil {
        t.Fatal("expected error for nonexistent pattern")
    }
}

func TestSearch(t *testing.T) {
    store := setupStore(t)

    store.Create(&Pattern{
        Title: "Empirical diagnosis loop", Description: "Agent loops without empirical data",
        Category: "anti-pattern", Source: "manual", Tags: "loop,diagnosis", Impact: "high",
    })
    store.Create(&Pattern{
        Title: "Cognitive mode switching", Description: "Audit changes constructive to destructive",
        Category: "cognitive-pattern", Source: "manual", Tags: "audit,cognitive", Impact: "medium",
    })

    results, err := store.Search("empirical")
    if err != nil {
        t.Fatalf("Search failed: %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("expected 1 result for 'empirical', got %d", len(results))
    }
    if results[0].Title != "Empirical diagnosis loop" {
        t.Fatalf("expected 'Empirical diagnosis loop', got %s", results[0].Title)
    }
}

func TestSearch_NoResults(t *testing.T) {
    store := setupStore(t)

    results, err := store.Search("xyznonexistent")
    if err != nil {
        t.Fatalf("Search failed: %v", err)
    }
    if len(results) != 0 {
        t.Fatalf("expected 0 results, got %d", len(results))
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/patterns/... -v`

Expected: Compilation error — `patterns` package doesn't exist yet.

- [ ] **Step 3: Implement PatternStore**

Create `internal/patterns/store.go`:

```go
package patterns

import (
    "fmt"
    "time"

    "github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
    "github.com/google/uuid"
)

// Valid category values enforced by SQLite CHECK constraint.
const (
    CategoryAntiPattern        = "anti-pattern"
    CategoryCognitivePattern   = "cognitive-pattern"
    CategoryStructuralPrinciple = "structural-principle"
    CategoryRoutingPrinciple   = "routing-principle"
)

// Valid source values enforced by SQLite CHECK constraint.
const (
    SourceCognitiveDNA     = "cognitive-dna"
    SourceEvolutionInsights = "evolution-insights"
    SourceSentinelLog      = "sentinel-log"
    SourceManual           = "manual"
    SourceEpiphany         = "epiphany"
)

// Valid impact values.
const (
    ImpactHigh   = "high"
    ImpactMedium = "medium"
    ImpactLow    = "low"
)

type Pattern struct {
    ID          string
    Title       string
    Description string
    Category    string
    Source      string
    SourceRef   string
    Tags        string
    Impact      string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type ListFilters struct {
    Category string
    Source   string
    Impact   string
    Limit    int
}

type PatternStore struct {
    db *sqlite.DB
}

func NewPatternStore(db *sqlite.DB) (*PatternStore, error) {
    if err := sqlite.ValidateDB(db, "pattern-store"); err != nil {
        return nil, err
    }
    return &PatternStore{db: db}, nil
}

func (s *PatternStore) Create(p *Pattern) (string, error) {
    id := uuid.New().String()
    query := `INSERT INTO patterns (id, title, description, category, source, source_ref, tags, impact)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
    _, err := s.db.Conn.Exec(query, id, p.Title, p.Description, p.Category, p.Source, p.SourceRef, p.Tags, p.Impact)
    if err != nil {
        return "", fmt.Errorf("patterns: failed to create pattern: %w", err)
    }
    return id, nil
}

func (s *PatternStore) List(filters ListFilters) ([]Pattern, error) {
    query := "SELECT id, title, description, category, source, source_ref, tags, impact, created_at, updated_at FROM patterns WHERE 1=1"
    var args []interface{}

    if filters.Category != "" {
        query += " AND category = ?"
        args = append(args, filters.Category)
    }
    if filters.Source != "" {
        query += " AND source = ?"
        args = append(args, filters.Source)
    }
    if filters.Impact != "" {
        query += " AND impact = ?"
        args = append(args, filters.Impact)
    }

    query += " ORDER BY created_at DESC"

    if filters.Limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", filters.Limit)
    }

    rows, err := s.db.Conn.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("patterns: failed to list patterns: %w", err)
    }
    defer rows.Close()

    return scanPatterns(rows)
}

func (s *PatternStore) Search(query string) ([]Pattern, error) {
    q := `SELECT p.id, p.title, p.description, p.category, p.source, p.source_ref, p.tags, p.impact, p.created_at, p.updated_at
          FROM patterns p
          JOIN patterns_fts fts ON p.rowid = fts.rowid
          WHERE patterns_fts MATCH ?
          ORDER BY bm25(patterns_fts) DESC
          LIMIT 20`
    rows, err := s.db.Conn.Query(q, query)
    if err != nil {
        return nil, fmt.Errorf("patterns: search failed: %w", err)
    }
    defer rows.Close()

    return scanPatterns(rows)
}

func (s *PatternStore) Get(id string) (*Pattern, error) {
    query := `SELECT id, title, description, category, source, source_ref, tags, impact, created_at, updated_at
              FROM patterns WHERE id = ?`
    var p Pattern
    var createdAt, updatedAt string
    err := s.db.Conn.QueryRow(query, id).Scan(
        &p.ID, &p.Title, &p.Description, &p.Category, &p.Source,
        &p.SourceRef, &p.Tags, &p.Impact, &createdAt, &updatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("patterns: pattern %s not found: %w", id, err)
    }
    p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
    p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
    return &p, nil
}

func scanPatterns(rows *sql.Rows) ([]Pattern, error) {
    var patterns []Pattern
    for rows.Next() {
        var p Pattern
        var createdAt, updatedAt string
        if err := rows.Scan(
            &p.ID, &p.Title, &p.Description, &p.Category, &p.Source,
            &p.SourceRef, &p.Tags, &p.Impact, &createdAt, &updatedAt,
        ); err != nil {
            return nil, fmt.Errorf("patterns: failed to scan pattern: %w", err)
        }
        p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
        p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
        patterns = append(patterns, p)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("patterns: row iteration error: %w", err)
    }
    return patterns, nil
}
```

**Fix**: The `scanPatterns` function uses `*sql.Rows` but needs the import. Add `"database/sql"` to imports. The function signature should be:

```go
import (
    "database/sql"
    "fmt"
    "time"

    "github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
    "github.com/google/uuid"
)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/patterns/... -v`

Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/patterns/store.go internal/patterns/store_test.go
git commit -m "feat(patterns): add PatternStore with CRUD and FTS5 search"
```

---

### Task 3: Deduplication — Levenshtein + tag overlap

**Files:**
- Create: `internal/patterns/dedup.go`
- Create: `internal/patterns/dedup_test.go`

- [ ] **Step 1: Write failing tests for dedup**

Create `internal/patterns/dedup_test.go`:

```go
package patterns

import (
    "testing"

    "github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
)

func TestLevenshteinDistance(t *testing.T) {
    tests := []struct {
        a, b     string
        expected int
    }{
        {"diagnosis loop", "diagnosis loop", 0},
        {"diagnosis loop", "diagnose loop", 2},
        {"abc", "axc", 1},
        {"", "abc", 3},
        {"abc", "", 3},
        {"kitten", "sitting", 3},
    }
    for _, tt := range tests {
        got := levenshteinDistance(tt.a, tt.b)
        if got != tt.expected {
            t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
        }
    }
}

func TestTagOverlap(t *testing.T) {
    tests := []struct {
        a, b     string
        expected float64
    }{
        {"loop,diagnosis,empirical", "loop,diagnosis,empirical", 1.0},
        {"loop,diagnosis", "loop,other", 0.5},
        {"a,b,c", "d,e,f", 0.0},
        {"", "a", 0.0},
        {"a", "", 0.0},
    }
    for _, tt := range tests {
        got := tagOverlap(tt.a, tt.b)
        if got != tt.expected {
            t.Errorf("tagOverlap(%q, %q) = %f, want %f", tt.a, tt.b, got, tt.expected)
        }
    }
}

func TestFindSimilar(t *testing.T) {
    db := testutil.SetupTestDB(t)
    t.Cleanup(func() { db.Close() })
    if err := graph.Migrate(db); err != nil {
        t.Fatalf("migration failed: %v", err)
    }
    store, _ := NewPatternStore(db)

    store.Create(&Pattern{
        Title: "Diagnóstico sem dado empírico = loop",
        Description: "Agent loops when diagnosing without data",
        Category: "anti-pattern", Source: "manual", Tags: "loop,diagnosis,empirical", Impact: "high",
    })

    // Similar title
    results, err := store.FindSimilar("Diagnóstico sem dado empírico", []string{"loop", "diagnosis"})
    if err != nil {
        t.Fatalf("FindSimilar failed: %v", err)
    }
    if len(results) == 0 {
        t.Fatal("expected similar pattern to be found")
    }

    // Completely different
    results, err = store.FindSimilar("Refactor module structure", []string{"refactor", "structure"})
    if err != nil {
        t.Fatalf("FindSimilar failed: %v", err)
    }
    if len(results) != 0 {
        t.Fatalf("expected no similar patterns, got %d", len(results))
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/patterns/... -v -run TestLevenshtein`

Expected: Compilation error — `levenshteinDistance` and `tagOverlap` not defined.

- [ ] **Step 3: Implement dedup logic**

Create `internal/patterns/dedup.go`:

```go
package patterns

import (
    "fmt"
    "strings"
)

// levenshteinDistance computes the edit distance between two strings.
// Standard Wagner-Fischer algorithm. O(m*n) time and space.
func levenshteinDistance(a, b string) int {
    la, lb := len(a), len(b)
    if la == 0 {
        return lb
    }
    if lb == 0 {
        return la
    }

    // Use two rows instead of full matrix for space efficiency
    prev := make([]int, lb+1)
    curr := make([]int, lb+1)

    for j := 0; j <= lb; j++ {
        prev[j] = j
    }

    for i := 1; i <= la; i++ {
        curr[0] = i
        for j := 1; j <= lb; j++ {
            cost := 1
            if a[i-1] == b[j-1] {
                cost = 0
            }
            curr[j] = min(
                prev[j]+1,     // deletion
                curr[j-1]+1,   // insertion
                prev[j-1]+cost, // substitution
            )
        }
        prev, curr = curr, prev
    }
    return prev[lb]
}

// tagOverlap computes the fraction of tags in a that also appear in b.
// Returns 0.0 if either has no tags. Case-insensitive comparison.
func tagOverlap(a, b string) float64 {
    tagsA := parseTags(a)
    tagsB := parseTags(b)
    if len(tagsA) == 0 || len(tagsB) == 0 {
        return 0.0
    }

    setB := make(map[string]bool, len(tagsB))
    for _, t := range tagsB {
        setB[strings.ToLower(t)] = true
    }

    matches := 0
    for _, t := range tagsA {
        if setB[strings.ToLower(t)] {
            matches++
        }
    }
    return float64(matches) / float64(len(tagsA))
}

func parseTags(s string) []string {
    if s == "" {
        return nil
    }
    parts := strings.Split(s, ",")
    result := make([]string, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" {
            result = append(result, p)
        }
    }
    return result
}

const (
    levenshteinThreshold = 3
    tagOverlapThreshold  = 0.5
)

// FindSimilar searches for patterns that are similar to the given title and tags.
// A pattern is considered similar if its title has Levenshtein distance <= 3
// or its tags overlap >= 50% with the given tags.
func (s *PatternStore) FindSimilar(title string, tags []string) ([]Pattern, error) {
    all, err := s.List(ListFilters{})
    if err != nil {
        return nil, fmt.Errorf("patterns: find similar: %w", err)
    }

    tagsStr := strings.Join(tags, ",")
    var similar []Pattern
    for _, p := range all {
        if levenshteinDistance(strings.ToLower(title), strings.ToLower(p.Title)) <= levenshteinThreshold {
            similar = append(similar, p)
            continue
        }
        if tagOverlap(tagsStr, p.Tags) >= tagOverlapThreshold {
            similar = append(similar, p)
        }
    }
    return similar, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/patterns/... -v`

Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/patterns/dedup.go internal/patterns/dedup_test.go
git commit -m "feat(patterns): add Levenshtein dedup and tag overlap detection"
```

---

### Task 4: Backfill from existing docs

**Files:**
- Create: `internal/patterns/backfill.go`
- Create: `internal/patterns/backfill_test.go`

- [ ] **Step 1: Write failing tests for backfill**

Create `internal/patterns/backfill_test.go`:

```go
package patterns

import (
    "testing"

    "github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
)

func TestBackfillFromCognitiveDNA(t *testing.T) {
    db := testutil.SetupTestDB(t)
    t.Cleanup(func() { db.Close() })
    if err := graph.Migrate(db); err != nil {
        t.Fatalf("migration failed: %v", err)
    }
    store, _ := NewPatternStore(db)

    result, err := store.BackfillFromCognitiveDNA()
    if err != nil {
        t.Fatalf("BackfillFromCognitiveDNA failed: %v", err)
    }

    // COGNITIVE-DNA.md has 3 APs + 3 PMOs = 6 patterns
    if result.Inserted == 0 {
        t.Fatal("expected at least 1 pattern inserted from COGNITIVE-DNA")
    }

    // Verify patterns were actually inserted
    patterns, err := store.List(ListFilters{Source: "cognitive-dna"})
    if err != nil {
        t.Fatalf("List failed: %v", err)
    }
    if len(patterns) != result.Inserted {
        t.Fatalf("expected %d patterns, got %d", result.Inserted, len(patterns))
    }
}

func TestBackfillFromCognitiveDNA_Idempotent(t *testing.T) {
    db := testutil.SetupTestDB(t)
    t.Cleanup(func() { db.Close() })
    if err := graph.Migrate(db); err != nil {
        t.Fatalf("migration failed: %v", err)
    }
    store, _ := NewPatternStore(db)

    // First backfill
    result1, _ := store.BackfillFromCognitiveDNA()

    // Second backfill — should skip all as duplicates
    result2, _ := store.BackfillFromCognitiveDNA()
    if result2.Inserted != 0 {
        t.Fatalf("expected 0 inserts on second run, got %d", result2.Inserted)
    }
    if result2.Skipped != result1.Inserted {
        t.Fatalf("expected %d skips, got %d", result1.Inserted, result2.Skipped)
    }
}

func TestBackfillFromEvolutionInsights(t *testing.T) {
    db := testutil.SetupTestDB(t)
    t.Cleanup(func() { db.Close() })
    if err := graph.Migrate(db); err != nil {
        t.Fatalf("migration failed: %v", err)
    }
    store, _ := NewPatternStore(db)

    result, err := store.BackfillFromEvolutionInsights()
    if err != nil {
        t.Fatalf("BackfillFromEvolutionInsights failed: %v", err)
    }
    if result.Inserted == 0 {
        t.Fatal("expected at least 1 pattern from EVOLUTION-INSIGHTS")
    }
}

func TestBackfillFromSentinelLog_DryRun(t *testing.T) {
    db := testutil.SetupTestDB(t)
    t.Cleanup(func() { db.Close() })
    if err := graph.Migrate(db); err != nil {
        t.Fatalf("migration failed: %v", err)
    }
    store, _ := NewPatternStore(db)

    candidates, err := store.BackfillFromSentinelLog(true) // dryRun = true
    if err != nil {
        t.Fatalf("BackfillFromSentinelLog dry-run failed: %v", err)
    }
    // Dry-run should return candidates without inserting
    patterns, _ := store.List(ListFilters{Source: "sentinel-log"})
    if len(patterns) != 0 {
        t.Fatalf("expected 0 patterns after dry-run, got %d", len(patterns))
    }
    _ = candidates // candidates list returned for inspection
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/patterns/... -v -run TestBackfill`

Expected: Compilation error — backfill methods not defined.

- [ ] **Step 3: Implement backfill**

Create `internal/patterns/backfill.go`:

```go
package patterns

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type BackfillResult struct {
    Extracted int
    Inserted  int
    Skipped   int
    Errors    []string
}

type BackfillCandidate struct {
    Title       string
    Description string
    Category    string
    Source      string
    SourceRef   string
    Tags        string
    Impact      string
}

// BackfillFromCognitiveDNA parses docs/process/COGNITIVE-DNA.md and inserts
// anti-patterns (AP-XX) and structural principles (PMO-XX) as patterns.
// Idempotent: skips patterns with same title.
func (s *PatternStore) BackfillFromCognitiveDNA() (BackfillResult, error) {
    var result BackfillResult
    candidates, err := parseCognitiveDNA("docs/process/COGNITIVE-DNA.md")
    if err != nil {
        return result, fmt.Errorf("patterns: backfill cognitive-dna: %w", err)
    }
    result.Extracted = len(candidates)

    for _, c := range candidates {
        // Check for duplicates by title
        existing, _ := s.List(ListFilters{})
        dup := false
        for _, p := range existing {
            if strings.EqualFold(p.Title, c.Title) {
                dup = true
                break
            }
        }
        if dup {
            result.Skipped++
            continue
        }

        _, err := s.Create(&Pattern{
            Title:       c.Title,
            Description: c.Description,
            Category:    c.Category,
            Source:      SourceCognitiveDNA,
            SourceRef:   c.SourceRef,
            Tags:        c.Tags,
            Impact:      c.Impact,
        })
        if err != nil {
            result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", c.Title, err))
            continue
        }
        result.Inserted++
    }
    return result, nil
}

// BackfillFromEvolutionInsights parses docs/process/EVOLUTION-INSIGHTS.md
// and inserts structural gaps and cognitive patterns.
func (s *PatternStore) BackfillFromEvolutionInsights() (BackfillResult, error) {
    var result BackfillResult
    candidates, err := parseEvolutionInsights("docs/process/EVOLUTION-INSIGHTS.md")
    if err != nil {
        return result, fmt.Errorf("patterns: backfill evolution-insights: %w", err)
    }
    result.Extracted = len(candidates)

    for _, c := range candidates {
        existing, _ := s.List(ListFilters{})
        dup := false
        for _, p := range existing {
            if strings.EqualFold(p.Title, c.Title) {
                dup = true
                break
            }
        }
        if dup {
            result.Skipped++
            continue
        }

        _, err := s.Create(&Pattern{
            Title:       c.Title,
            Description: c.Description,
            Category:    c.Category,
            Source:      SourceEvolutionInsights,
            SourceRef:   c.SourceRef,
            Tags:        c.Tags,
            Impact:      c.Impact,
        })
        if err != nil {
            result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", c.Title, err))
            continue
        }
        result.Inserted++
    }
    return result, nil
}

// BackfillFromSentinelLog extracts pattern candidates from sentinel-log.md.
// When dryRun=true, it returns candidates without inserting them.
// When dryRun=false, it is NOT recommended (free-form parsing produces false positives).
func (s *PatternStore) BackfillFromSentinelLog(dryRun bool) ([]BackfillCandidate, error) {
    candidates, err := parseSentinelLog("docs/process/sentinel-log.md")
    if err != nil {
        return nil, fmt.Errorf("patterns: backfill sentinel-log: %w", err)
    }
    if dryRun {
        return candidates, nil
    }
    // Non-dry-run: insert candidates
    for _, c := range candidates {
        s.Create(&Pattern{
            Title:       c.Title,
            Description: c.Description,
            Category:    c.Category,
            Source:      SourceSentinelLog,
            SourceRef:   c.SourceRef,
            Tags:        c.Tags,
            Impact:      c.Impact,
        })
    }
    return candidates, nil
}

// parseCognitiveDNA extracts AP-XX and PMO-XX entries from COGNITIVE-DNA.md.
// Parses the table format (AP-XX rows) and section format (PMO-XX headers).
func parseCognitiveDNA(path string) ([]BackfillCandidate, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    var candidates []BackfillCandidate
    scanner := bufio.NewScanner(f)
    var currentPMO BackfillCandidate
    var inPMO bool
    var pmoBody strings.Builder

    for scanner.Scan() {
        line := scanner.Text()

        // Parse AP entries from table rows: | **[AP-XX]** | Name | Failure Line | Motivation |
        if strings.Contains(line, "[AP-") {
            parts := strings.Split(line, "|")
            if len(parts) >= 5 {
                idPart := strings.TrimSpace(parts[1])
                namePart := strings.TrimSpace(parts[2])
                descPart := strings.TrimSpace(parts[3])
                // Clean markdown bold markers
                id := strings.TrimPrefix(strings.TrimSuffix(idPart, "**"), "**")
                name := strings.TrimPrefix(strings.TrimSuffix(namePart, "**"), "**")
                desc := strings.TrimPrefix(strings.TrimSuffix(descPart, "**"), "**")

                candidates = append(candidates, BackfillCandidate{
                    Title:       fmt.Sprintf("%s: %s", id, name),
                    Description: desc,
                    Category:    CategoryAntiPattern,
                    SourceRef:   fmt.Sprintf("COGNITIVE-DNA.md:%s", id),
                    Tags:        "anti-pattern,cognitive-dna",
                    Impact:      ImpactHigh,
                })
            }
        }

        // Parse PMO sections: ### PMO-XX: Title
        if strings.HasPrefix(line, "### PMO-") {
            if inPMO && pmoBody.Len() > 0 {
                currentPMO.Description = strings.TrimSpace(pmoBody.String())
                candidates = append(candidates, currentPMO)
            }
            title := strings.TrimPrefix(line, "### ")
            currentPMO = BackfillCandidate{
                Title:     title,
                Category:  CategoryStructuralPrinciple,
                SourceRef: fmt.Sprintf("COGNITIVE-DNA.md:%s", title),
                Tags:      "modus-operandi,cognitive-dna",
                Impact:    ImpactMedium,
            }
            inPMO = true
            pmoBody.Reset()
            continue
        }

        if inPMO {
            // Collect body lines (skip the rule/modus lines header)
            if strings.Contains(line, "- **Regra:**") {
                pmoBody.WriteString(strings.TrimPrefix(line, "- **Regra:**"))
                pmoBody.WriteString(" ")
            } else if strings.Contains(line, "- **Modus Operandi:**") {
                pmoBody.WriteString(strings.TrimPrefix(line, "- **Modus Operandi:**"))
            }
        }
    }

    // Don't forget the last PMO
    if inPMO && pmoBody.Len() > 0 {
        currentPMO.Description = strings.TrimSpace(pmoBody.String())
        candidates = append(candidates, currentPMO)
    }

    return candidates, scanner.Err()
}

// parseEvolutionInsights extracts gaps and cognitive patterns from EVOLUTION-INSIGHTS.md.
func parseEvolutionInsights(path string) ([]BackfillCandidate, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    var candidates []BackfillCandidate
    scanner := bufio.NewScanner(f)
    inGaps := false
    inCognitive := false

    for scanner.Scan() {
        line := scanner.Text()

        if strings.Contains(line, "Gaps Estruturais") {
            inGaps = true
            inCognitive = false
            continue
        }
        if strings.Contains(line, "Cognitive Patterns") {
            inGaps = false
            inCognitive = true
            continue
        }
        if strings.HasPrefix(line, "##") {
            inGaps = false
            inCognitive = false
            continue
        }

        // Parse list items: - [ ] Task Metadata Anemia: ...
        if (inGaps || inCognitive) && strings.HasPrefix(line, "- ") {
            // Extract title from the line
            clean := strings.TrimPrefix(line, "- ")
            clean = strings.TrimPrefix(clean, "[x] ")
            clean = strings.TrimPrefix(clean, "[ ] ")
            clean = strings.TrimPrefix(clean, "**")
            clean = strings.TrimSuffix(clean, "**")

            // Split on colon for title + description
            parts := strings.SplitN(clean, ":", 2)
            title := strings.TrimSpace(parts[0])
            desc := title
            if len(parts) > 1 {
                desc = strings.TrimSpace(parts[1])
            }

            if title == "" {
                continue
            }

            // Skip items marked as ~~strikethrough~~ (implemented)
            if strings.Contains(line, "~~") {
                continue
            }

            category := CategoryStructuralPrinciple
            if inCognitive {
                category = CategoryCognitivePattern
            }

            candidates = append(candidates, BackfillCandidate{
                Title:       title,
                Description: desc,
                Category:    category,
                SourceRef:   "EVOLUTION-INSIGHTS.md",
                Tags:        "evolution-insights",
                Impact:      ImpactMedium,
            })
        }
    }

    return candidates, scanner.Err()
}

// parseSentinelLog extracts epiphany candidates from sentinel-log.md.
// Returns candidates for dry-run review. Does not auto-insert.
func parseSentinelLog(path string) ([]BackfillCandidate, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    var candidates []BackfillCandidate
    scanner := bufio.NewScanner(f)

    for scanner.Scan() {
        line := scanner.Text()
        // Look for epiphany markers: **Filtro A/B/C**
        if strings.Contains(line, "Filtro A") || strings.Contains(line, "Filtro B") || strings.Contains(line, "Filtro C") {
            // Extract the line as a candidate title
            clean := strings.TrimPrefix(line, "- ")
            clean = strings.TrimPrefix(clean, "* ")
            clean = strings.TrimPrefix(clean, "**")

            if len(clean) > 10 { // Skip very short lines
                filtro := "unknown"
                if strings.Contains(line, "Filtro A") {
                    filtro = "A"
                } else if strings.Contains(line, "Filtro B") {
                    filtro = "B"
                } else if strings.Contains(line, "Filtro C") {
                    filtro = "C"
                }

                candidates = append(candidates, BackfillCandidate{
                    Title:       clean,
                    Description: clean,
                    Category:    CategoryRoutingPrinciple, // Default — user reclassifies on add
                    SourceRef:   fmt.Sprintf("sentinel-log.md:Filtro-%s", filtro),
                    Tags:        "epiphany,sentinel-log",
                    Impact:      ImpactMedium,
                })
            }
        }
    }

    return candidates, scanner.Err()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/patterns/... -v -run TestBackfill`

Expected: Tests PASS. COGNITIVE-DNA backfill inserts 6 patterns (3 APs + 3 PMOs). Idempotent second run skips all. EVOLUTION-INSIGHTS inserts gaps + cognitive patterns. Sentinel-log dry-run returns candidates without inserting.

Note: Tests depend on actual markdown files existing at the expected paths. If running from project root, paths should resolve. If not, adjust path logic.

- [ ] **Step 5: Commit**

```bash
git add internal/patterns/backfill.go internal/patterns/backfill_test.go
git commit -m "feat(patterns): add backfill from COGNITIVE-DNA, EVOLUTION-INSIGHTS, sentinel-log"
```

---

### Task 5: CLI commands — sentinel pattern {add,list,search,get,backfill}

**Files:**
- Create: `cmd/sentinel/commands/pattern.go`

- [ ] **Step 1: Implement the pattern CLI command**

Create `cmd/sentinel/commands/pattern.go`:

```go
package commands

import (
    "fmt"
    "os"
    "strings"

    "github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/patterns"
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
    "github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
    "github.com/spf13/cobra"
)

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

var (
    addTitle     string
    addDesc      string
    addCategory  string
    addSource    string
    addSourceRef string
    addTags      string
    addImpact    string
    addForce     bool
)

func patternAddCmd(db *sqlite.DB) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "add",
        Short: "Capture a new pattern",
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := graph.Migrate(db); err != nil {
                return fmt.Errorf("pattern add: migration failed: %w", err)
            }
            store, err := patterns.NewPatternStore(db)
            if err != nil {
                return err
            }

            // Dedup check (unless --force)
            if !addForce {
                tagSlice := strings.Split(addTags, ",")
                similar, _ := store.FindSimilar(addTitle, tagSlice)
                if len(similar) > 0 {
                    fmt.Printf("[SENTINEL] Similar pattern found: %q (ID: %s)\n", similar[0].Title, similar[0].ID)
                    fmt.Println("[SENTINEL] Use --force to create anyway.")
                    return nil
                }
            }

            if addSource == "" {
                addSource = patterns.SourceManual
            }
            if addImpact == "" {
                addImpact = patterns.ImpactMedium
            }

            id, err := store.Create(&patterns.Pattern{
                Title:       addTitle,
                Description: addDesc,
                Category:    addCategory,
                Source:      addSource,
                SourceRef:   addSourceRef,
                Tags:        addTags,
                Impact:      addImpact,
            })
            if err != nil {
                return fmt.Errorf("pattern add: %w", err)
            }

            fmt.Printf("✅ PATTERN CAPTURED [ID: %s]: %s\n", id, addTitle)
            return nil
        },
    }

    cmd.Flags().StringVar(&addTitle, "title", "", "Pattern title (required)")
    cmd.Flags().StringVar(&addDesc, "desc", "", "Pattern description (required)")
    cmd.Flags().StringVar(&addCategory, "category", "", "Category: anti-pattern, cognitive-pattern, structural-principle, routing-principle (required)")
    cmd.Flags().StringVar(&addSource, "source", "", "Source: cognitive-dna, evolution-insights, sentinel-log, manual, epiphany (default: manual)")
    cmd.Flags().StringVar(&addSourceRef, "source-ref", "", "Reference to original source location")
    cmd.Flags().StringVar(&addTags, "tags", "", "Comma-separated tags")
    cmd.Flags().StringVar(&addImpact, "impact", "", "Impact: high, medium, low (default: medium)")
    cmd.Flags().BoolVar(&addForce, "force", false, "Skip dedup check")

    cmd.MarkFlagRequired("title")
    cmd.MarkFlagRequired("desc")
    cmd.MarkFlagRequired("category")

    return cmd
}

var (
    listCategory string
    listSource   string
    listImpact   string
)

func patternListCmd(db *sqlite.DB) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List captured patterns",
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := graph.Migrate(db); err != nil {
                return fmt.Errorf("pattern list: migration failed: %w", err)
            }
            store, err := patterns.NewPatternStore(db)
            if err != nil {
                return err
            }

            result, err := store.List(patterns.ListFilters{
                Category: listCategory,
                Source:   listSource,
                Impact:   listImpact,
                Limit:    20,
            })
            if err != nil {
                return fmt.Errorf("pattern list: %w", err)
            }

            if len(result) == 0 {
                fmt.Println("No patterns found.")
                return nil
            }

            fmt.Printf("%-38s %-40s %-20s %-8s %-15s\n", "ID", "TITLE", "CATEGORY", "IMPACT", "SOURCE")
            for _, p := range result {
                title := p.Title
                if len(title) > 38 {
                    title = title[:35] + "..."
                }
                fmt.Printf("%-38s %-40s %-20s %-8s %-15s\n",
                    p.ID, title, p.Category, p.Impact, p.Source)
            }
            return nil
        },
    }

    cmd.Flags().StringVar(&listCategory, "category", "", "Filter by category")
    cmd.Flags().StringVar(&listSource, "source", "", "Filter by source")
    cmd.Flags().StringVar(&listImpact, "impact", "", "Filter by impact")

    return cmd
}

func patternSearchCmd(db *sqlite.DB) *cobra.Command {
    return &cobra.Command{
        Use:   "search [query]",
        Short: "Full-text search across patterns",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := graph.Migrate(db); err != nil {
                return fmt.Errorf("pattern search: migration failed: %w", err)
            }
            store, err := patterns.NewPatternStore(db)
            if err != nil {
                return err
            }

            result, err := store.Search(args[0])
            if err != nil {
                return fmt.Errorf("pattern search: %w", err)
            }

            if len(result) == 0 {
                fmt.Printf("No patterns found matching %q.\n", args[0])
                return nil
            }

            fmt.Printf("%-38s %-40s %-20s %-8s\n", "ID", "TITLE", "CATEGORY", "IMPACT")
            for _, p := range result {
                title := p.Title
                if len(title) > 38 {
                    title = title[:35] + "..."
                }
                fmt.Printf("%-38s %-40s %-20s %-8s\n", p.ID, title, p.Category, p.Impact)
            }
            return nil
        },
    }
}

func patternGetCmd(db *sqlite.DB) *cobra.Command {
    return &cobra.Command{
        Use:   "get [id]",
        Short: "Show full details of a pattern",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := graph.Migrate(db); err != nil {
                return fmt.Errorf("pattern get: migration failed: %w", err)
            }
            store, err := patterns.NewPatternStore(db)
            if err != nil {
                return err
            }

            p, err := store.Get(args[0])
            if err != nil {
                fmt.Fprintf(os.Stderr, "Pattern not found: %s\n", args[0])
                os.Exit(1)
            }

            fmt.Printf("ID:          %s\n", p.ID)
            fmt.Printf("Title:       %s\n", p.Title)
            fmt.Printf("Description: %s\n", p.Description)
            fmt.Printf("Category:    %s\n", p.Category)
            fmt.Printf("Source:      %s\n", p.Source)
            fmt.Printf("Source Ref:  %s\n", p.SourceRef)
            fmt.Printf("Tags:        %s\n", p.Tags)
            fmt.Printf("Impact:      %s\n", p.Impact)
            fmt.Printf("Created:     %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
            fmt.Printf("Updated:     %s\n", p.UpdatedAt.Format("2006-01-02 15:04:05"))
            return nil
        },
    }
}

var backfillAll bool

func patternBackfillCmd(db *sqlite.DB) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "backfill",
        Short: "Extract and insert patterns from existing documentation",
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := graph.Migrate(db); err != nil {
                return fmt.Errorf("pattern backfill: migration failed: %w", err)
            }
            store, err := patterns.NewPatternStore(db)
            if err != nil {
                return err
            }

            source := backfillSource

            if backfillAll || source == "cognitive-dna" {
                result, err := store.BackfillFromCognitiveDNA()
                if err != nil {
                    fmt.Fprintf(os.Stderr, "warning: cognitive-dna backfill: %v\n", err)
                } else {
                    fmt.Printf("Cognitive-DNA: %d extracted, %d inserted, %d skipped\n",
                        result.Extracted, result.Inserted, result.Skipped)
                }
            }

            if backfillAll || source == "evolution-insights" {
                result, err := store.BackfillFromEvolutionInsights()
                if err != nil {
                    fmt.Fprintf(os.Stderr, "warning: evolution-insights backfill: %v\n", err)
                } else {
                    fmt.Printf("Evolution-Insights: %d extracted, %d inserted, %d skipped\n",
                        result.Extracted, result.Inserted, result.Skipped)
                }
            }

            if source == "sentinel-log" {
                candidates, err := store.BackfillFromSentinelLog(true) // always dry-run
                if err != nil {
                    fmt.Fprintf(os.Stderr, "warning: sentinel-log backfill: %v\n", err)
                } else {
                    fmt.Printf("[DRY-RUN] %d candidates extracted from sentinel-log:\n", len(candidates))
                    for i, c := range candidates {
                        fmt.Printf("  %d. %q (%s)\n", i+1, c.Title, c.SourceRef)
                    }
                    fmt.Println("Use 'sentinel pattern add' to capture selected patterns.")
                }
            }

            return nil
        },
    }

    cmd.Flags().StringVar(&backfillSource, "source", "", "Source: cognitive-dna, evolution-insights, sentinel-log")
    cmd.Flags().BoolVar(&backfillAll, "all", false, "Backfill from cognitive-dna + evolution-insights (excludes sentinel-log)")

    return cmd
}

var backfillSource string
```

- [ ] **Step 2: Build to verify compilation**

Run: `go build ./...`

Expected: Build succeeds with no errors.

- [ ] **Step 3: Manual smoke test**

Run: `go run ./cmd/sentinel main.go pattern add --title "Test pattern" --desc "Test description" --category anti-pattern --impact high`

Expected: `✅ PATTERN CAPTURED [ID: ...]: Test pattern`

Run: `go run ./cmd/sentinel main.go pattern list`

Expected: Table with the test pattern.

Run: `go run ./cmd/sentinel main.go pattern search "Test"`

Expected: Table with the test pattern.

Run: `go run ./cmd/sentinel main.go pattern backfill --source cognitive-dna`

Expected: `Cognitive-DNA: 6 extracted, 6 inserted, 0 skipped` (or similar counts).

Run: `go run ./cmd/sentinel main.go pattern backfill --source sentinel-log`

Expected: `[DRY-RUN] N candidates extracted from sentinel-log:` followed by list.

- [ ] **Step 4: Commit**

```bash
git add cmd/sentinel/commands/pattern.go
git commit -m "feat(patterns): add sentinel pattern CLI with add/list/search/get/backfill"
```

---

### Task 6: Filtro D — Epiphany Protocol Extension

**Files:**
- Modify: `GEMINI.md:37-38` (Epiphany Protocol section)

- [ ] **Step 1: Add Filtro D to Epiphany Protocol in GEMINI.md**

In `GEMINI.md`, after line 37 (the existing Filtro C reference in the "PROTOCOLO DE EPIFANIA" section), add:

```markdown
4. **FILTRO D — Decision Routing**: Quando uma epifania revela um princípio sobre COMO rotear decisões (não apenas o que aconteceu), o agente DEVE capturá-lo via `sentinel pattern add --source epiphany --category routing-principle`. Exemplo: "Auditoria troca modo cognitivo de construtivo para destrutivo" → Filtro D → routing-principle.
```

The full updated section reads:

```markdown
- **🛡️ PROTOCOLO DE EPIFANIA (Sessão de Reflexão)**:
1. **RIGOR PROPORCIONAL**: A atualização de documentação deve seguir os Tiers do Standard #14. Tarefas Trivial (T1) não exigem logs. Mudanças de arquitetura (T3) exigem auditoria completa.
2. **DELTA CHECK**: O agente deve registrar insights significativos em `docs/process/EVOLUTION-INSIGHTS.md` sempre que uma nova "lição universal" for aprendida.
3. O `sentinel-log.md` deve ser atualizado com a síntese das decisões tomadas.
4. **FILTRO D — Decision Routing**: Quando uma epifania revela um princípio sobre COMO rotear decisões (não apenas o que aconteceu), o agente DEVE capturá-lo via `sentinel pattern add --source epiphany --category routing-principle`. Exemplo: "Auditoria troca modo cognitivo de construtivo para destrutivo" → Filtro D → routing-principle.
5. **OBRIGATÓRIO**: Toda entrega final deve conter o **Sovereign Audit Framework (Standard #08)**.
```

- [ ] **Step 2: Verify GEMINI.md is valid markdown**

Run: `cat GEMINI.md | head -80`

Expected: Filtro D appears in the Epiphany Protocol section.

- [ ] **Step 3: Commit**

```bash
git add GEMINI.md
git commit -m "docs: add Filtro D (Decision Routing) to Epiphany Protocol"
```

---

### Task 7: Full verification — build + test + vet

**Files:** No new files. Verification only.

- [ ] **Step 1: Run full build**

Run: `go build ./...`

Expected: Exit code 0. No compilation errors.

- [ ] **Step 2: Run all pattern tests**

Run: `go test ./internal/patterns/... -v`

Expected: All tests PASS.

- [ ] **Step 3: Run schema tests**

Run: `go test ./internal/graph/... -v`

Expected: All tests PASS including new patterns table check.

- [ ] **Step 4: Run go vet**

Run: `go vet ./...`

Expected: No issues reported.

- [ ] **Step 5: Run full test suite**

Run: `go test ./...`

Expected: All tests PASS. (Note: pre-existing failures in other packages are not our responsibility.)

- [ ] **Step 6: Final commit (if any fixes needed)**

```bash
git add -A
git commit -m "fix(patterns): address test/vet findings from full verification"
```

---

## Self-Review

### Spec coverage
- ✅ Schema migration (patterns table + FTS5 + triggers) → Task 1
- ✅ PatternStore CRUD (Create, List, Search, Get) → Task 2
- ✅ Deduplication (FindSimilar + Levenshtein) → Task 3
- ✅ Backfill from COGNITIVE-DNA, EVOLUTION-INSIGHTS, sentinel-log → Task 4
- ✅ CLI commands (add/list/search/get/backfill) → Task 5
- ✅ Filtro D in GEMINI.md → Task 6
- ✅ Verification gate → Task 7

### Placeholder scan
- No TBD, TODO, "implement later", "fill in details"
- No "add appropriate error handling" without showing the code
- All code steps include actual Go code
- All test steps include actual test code
- All commands specify expected output

### Type consistency
- `Pattern` struct fields match between store.go, backfill.go, and pattern.go
- `ListFilters` struct matches between store.go and pattern.go
- `BackfillCandidate` struct matches between backfill.go and backfill.go
- `BackfillResult` struct used consistently in backfill methods
- Constants (Category*, Source*, Impact*) defined once in store.go, used everywhere
