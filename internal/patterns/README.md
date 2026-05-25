# internal/patterns

Architectural pattern storage with FTS5 full-text search, deduplication, and documentation backfill.

## Overview

The patterns package manages a catalog of architectural patterns (anti-patterns, cognitive patterns, structural principles, routing principles) in SQLite with FTS5-powered search. It supports extraction from existing documentation (COGNITIVE-DNA.md, EVOLUTION-INSIGHTS.md, sentinel-log.md) via backfill pipelines.

## Key Types

### `Pattern`
Represents an architectural pattern with ID, title, description, category, source, tags, impact level, and timestamps. Categories: `anti-pattern`, `cognitive-pattern`, `structural-principle`, `routing-principle`. Sources: `cognitive-dna`, `evolution-insights`, `sentinel-log`, `manual`, `epiphany`.

### `PatternStore`
CRUD operations backed by SQLite with nil-guard validation:
- `NewPatternStore(db)` — constructor with `ValidateDB`
- `Create(ctx, pattern)` — inserts new pattern with UUID
- `List(ctx, filters)` — filtered listing with category/source/impact filters
- `Search(ctx, query)` — FTS5 search via `patterns_fts` virtual table with BM25 ranking
- `Get(ctx, id)` — single pattern lookup
- `FindSimilar(ctx, title, tags)` — Levenshtein distance (threshold ≤ 3) and tag overlap (≥ 50%)

### `ListFilters`
Filter parameters: `Category`, `Source`, `Impact`, `Limit`.

## Backfill

Three backfill functions extract patterns from documentation:
- `BackfillFromCognitiveDNA(ctx, baseDir)` — anti-patterns (AP-*) and structural principles (PMO-*)
- `BackfillFromEvolutionInsights(ctx, baseDir)` — structural gaps and cognitive patterns
- `BackfillFromSentinelLog(ctx, baseDir, dryRun)` — routing principles from Filtro A/B/C entries

Each backfill deduplicates by title (case-insensitive) before insertion.

## Dependencies

- `pkg/sqlite` — DB abstraction
- `github.com/google/uuid` — pattern ID generation
- FTS5 virtual table (created by `graph.Migrate`)

## Usage

```go
store, _ := patterns.NewPatternStore(db)

// Create
id, _ := store.Create(ctx, &patterns.Pattern{
    Title:       "AP-001: God Object",
    Description: "Struct exceeding cohesion threshold",
    Category:    patterns.CategoryAntiPattern,
    Source:      patterns.SourceManual,
    Impact:      patterns.ImpactHigh,
})

// Search
results, _ := store.Search(ctx, "god object cohesion")

// Backfill from docs
result, _ := store.BackfillFromCognitiveDNA(ctx, ".")
fmt.Printf("Extracted: %d, Inserted: %d\n", result.Extracted, result.Inserted)
```
