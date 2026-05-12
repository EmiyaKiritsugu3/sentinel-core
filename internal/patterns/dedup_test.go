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
		Title: "Diagnóstico sem dado empírico = loop", Description: "Agent loops when diagnosing without data",
		Category: "anti-pattern", Source: "manual", Tags: "loop,diagnosis,empirical", Impact: "high",
	})

	results, err := store.FindSimilar("Diagnóstico sem dado empírico", []string{"loop", "diagnosis"})
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected similar pattern to be found")
	}

	results, err = store.FindSimilar("Refactor module structure", []string{"refactor", "structure"})
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no similar patterns, got %d", len(results))
	}
}

// Cobertura: FindSimilar — ramo de tag overlap (sem match levenshtein, mas overlap ≥ 0.5)

func TestFindSimilar_TagOverlap(t *testing.T) {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { db.Close() })
	if err := graph.Migrate(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, _ := NewPatternStore(db)

	store.Create(&Pattern{
		Title:       "Título completamente diferente",
		Description: "desc",
		Category:    "anti-pattern",
		Source:      "manual",
		Tags:        "loop,diagnosis,empirical",
		Impact:      "high",
	})

	results, err := store.FindSimilar("Outro título sem similaridade", []string{"loop", "diagnosis"})
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected pattern found via tag overlap path")
	}
}
