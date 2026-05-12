package patterns

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
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

// CG-02: Validação nil-DB em métodos exportados — cada método deve
// retornar ErrNilDB independente do wiring do construtor.

func TestCreate_NilDB(t *testing.T) {
	s := &PatternStore{}
	_, err := s.Create(&Pattern{Title: "x", Description: "x", Category: "anti-pattern", Source: "manual", Tags: "x", Impact: "high"})
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}

func TestCreate_NilPattern(t *testing.T) {
	store := setupStore(t)
	_, err := store.Create(nil)
	if err == nil {
		t.Fatal("expected error for nil pattern")
	}
}

func TestList_NilDB(t *testing.T) {
	s := &PatternStore{}
	_, err := s.List(ListFilters{})
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}

func TestSearch_NilDB(t *testing.T) {
	s := &PatternStore{}
	_, err := s.Search("test")
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}

func TestGet_NilDB(t *testing.T) {
	s := &PatternStore{}
	_, err := s.Get("id")
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
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

	_, err := store.Create(&Pattern{
		Title: "P1", Description: "D1", Category: "anti-pattern",
		Source: "manual", Tags: "a", Impact: "high",
	})
	if err != nil {
		t.Fatalf("seed Create P1 failed: %v", err)
	}
	_, err = store.Create(&Pattern{
		Title: "P2", Description: "D2", Category: "cognitive-pattern",
		Source: "epiphany", Tags: "b", Impact: "low",
	})
	if err != nil {
		t.Fatalf("seed Create P2 failed: %v", err)
	}

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

	_, err := store.Create(&Pattern{
		Title: "P1", Description: "D1", Category: "anti-pattern",
		Source: "manual", Tags: "a", Impact: "high",
	})
	if err != nil {
		t.Fatalf("seed Create P1 failed: %v", err)
	}
	_, err = store.Create(&Pattern{
		Title: "P2", Description: "D2", Category: "cognitive-pattern",
		Source: "epiphany", Tags: "b", Impact: "low",
	})
	if err != nil {
		t.Fatalf("seed Create P2 failed: %v", err)
	}

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

	id, err := store.Create(&Pattern{
		Title: "Test Pattern", Description: "Full description here",
		Category: "anti-pattern", Source: "cognitive-dna",
		SourceRef: "COGNITIVE-DNA.md:AP-01", Tags: "test", Impact: "medium",
	})
	if err != nil {
		t.Fatalf("seed Create failed: %v", err)
	}

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

	_, err := store.Create(&Pattern{
		Title: "Empirical diagnosis loop", Description: "Agent loops without empirical data",
		Category: "anti-pattern", Source: "manual", Tags: "loop,diagnosis", Impact: "high",
	})
	if err != nil {
		t.Fatalf("seed Create P1 failed: %v", err)
	}
	_, err = store.Create(&Pattern{
		Title: "Cognitive mode switching", Description: "Audit changes constructive to destructive",
		Category: "cognitive-pattern", Source: "manual", Tags: "audit,cognitive", Impact: "medium",
	})
	if err != nil {
		t.Fatalf("seed Create P2 failed: %v", err)
	}

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

// Cobertura: List — ramo LIMIT

func TestList_WithLimit(t *testing.T) {
	store := setupStore(t)

	_, err := store.Create(&Pattern{
		Title: "P1", Description: "D1", Category: "anti-pattern",
		Source: "manual", Tags: "a", Impact: "high",
	})
	if err != nil {
		t.Fatalf("seed Create P1 failed: %v", err)
	}
	_, err = store.Create(&Pattern{
		Title: "P2", Description: "D2", Category: "cognitive-pattern",
		Source: "epiphany", Tags: "b", Impact: "low",
	})
	if err != nil {
		t.Fatalf("seed Create P2 failed: %v", err)
	}
	_, err = store.Create(&Pattern{
		Title: "P3", Description: "D3", Category: "structural-principle",
		Source: "cognitive-dna", Tags: "c", Impact: "medium",
	})
	if err != nil {
		t.Fatalf("seed Create P3 failed: %v", err)
	}

	patterns, err := store.List(ListFilters{Limit: 2})
	if err != nil {
		t.Fatalf("List with limit failed: %v", err)
	}
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns with limit, got %d", len(patterns))
	}
}

// Cobertura: List — filtro por Impact

// CG-01: bm25 ASC retorna melhor match primeiro (FP: DESC retornaria pior primeiro)

func TestSearch_Ranking(t *testing.T) {
	store := setupStore(t)

	store.Create(&Pattern{
		Title:       "Loop diagnosis empirical",
		Description: "Diagnosis loop repeated diagnosis pattern with empirical data",
		Category:    "anti-pattern",
		Source:      "manual",
		Tags:        "diagnosis,loop,empirical",
		Impact:      "high",
	})
	store.Create(&Pattern{
		Title:       "Side observation",
		Description: "A diagnosis mention in passing",
		Category:    "cognitive-pattern",
		Source:      "manual",
		Tags:        "misc",
		Impact:      "low",
	})

	results, err := store.Search("diagnosis")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}

	if results[0].Title != "Loop diagnosis empirical" {
		t.Fatalf("expected best match 'Loop diagnosis empirical' first, got %q — bm25 ASC ordering broken", results[0].Title)
	}
}

func TestList_FilterByImpact(t *testing.T) {
	store := setupStore(t)

	store.Create(&Pattern{
		Title: "P1", Description: "D1", Category: "anti-pattern",
		Source: "manual", Tags: "a", Impact: "high",
	})
	store.Create(&Pattern{
		Title: "P2", Description: "D2", Category: "cognitive-pattern",
		Source: "epiphany", Tags: "b", Impact: "low",
	})

	patterns, err := store.List(ListFilters{Impact: "high"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(patterns) != 1 || patterns[0].Title != "P1" {
		t.Fatalf("expected 1 high-impact pattern P1, got %v", patterns)
	}
}
