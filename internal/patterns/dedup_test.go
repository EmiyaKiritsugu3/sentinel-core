package patterns

import (
	"context"
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestLevenshteinDistance(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}

	_, err = store.Create(context.Background(), &Pattern{
		Title: "Diagnóstico sem dado empírico = loop", Description: "Agent loops when diagnosing without data",
		Category: "anti-pattern", Source: "manual", Tags: "loop,diagnosis,empirical", Impact: "high",
	})
	if err != nil {
		t.Fatalf("seed Create failed: %v", err)
	}

	results, err := store.FindSimilar(context.Background(), "Diagnóstico sem dado empírico", []string{"loop", "diagnosis"})
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected similar pattern to be found")
	}

	results, err = store.FindSimilar(context.Background(), "Refactor module structure", []string{"refactor", "structure"})
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no similar patterns, got %d", len(results))
	}
}

// CG-02: FindSimilar should return ErrNilDB when store has no DB

func TestFindSimilar_NilDB(t *testing.T) {
	t.Parallel()
	s := &PatternStore{}
	_, err := s.FindSimilar(context.Background(), "test", []string{"a"})
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}

// Coverage: FindSimilar — tag overlap branch (no levenshtein match, but overlap ≥ 0.5)

func TestFindSimilar_TagOverlap(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}

	_, err = store.Create(context.Background(), &Pattern{
		Title:       "Título completamente diferente",
		Description: "desc",
		Category:    "anti-pattern",
		Source:      "manual",
		Tags:        "loop,diagnosis,empirical",
		Impact:      "high",
	})
	if err != nil {
		t.Fatalf("seed Create failed: %v", err)
	}

	results, err := store.FindSimilar(context.Background(), "Outro título sem similaridade", []string{"loop", "diagnosis"})
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected pattern found via tag overlap path")
	}
}
