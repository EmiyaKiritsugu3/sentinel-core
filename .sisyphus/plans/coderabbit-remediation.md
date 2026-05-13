# CodeRabbit Review Remediation Plan — PR #9

**Branch**: `feat/audit-remediation`
**Total CodeRabbit Comments**: 24 (1 Critical, 14 Major, 9 Minor)
**Outdated (auto-resolved)**: 4
**Actionable**: 20

---

## Resolved by Previous Commits (4 outdated — skip)

| # | File | Issue | Status |
|---|------|-------|--------|
| M7 | backfill.go | `s.List` errors ignored | Fixed in rewrite |
| M9 | backfill.go | Section header substring matching | Fixed in rewrite |
| M10 | backfill.go | `parseSentinelLog` over-match | Fixed in rewrite |
| m6 | store.go | Timestamp parse failures swallowed | Fixed with `timeLayout` constant |

---

## Chapter 1: Bug Fix — bm25 ORDER BY Direction

**Severity**: Major (logic bug — search returns worst matches first)
**File**: `internal/patterns/store.go:114`
**Change**:
```go
// BEFORE
ORDER BY bm25(patterns_fts) DESC

// AFTER
ORDER BY bm25(patterns_fts) ASC
```
**Rationale**: SQLite FTS5 `bm25()` returns negative values where more negative = better match. `DESC` puts worst matches first; `ASC` puts best matches first.
**Test**: Add `TestSearch_Ranking` — insert 2 patterns with different relevance, verify order.

---

## Chapter 2: Nil-DB Guards in Exported Methods (CG-02 Compliance)

**Severity**: Critical (1) + Major (3) — 4 CodeRabbit comments
**Files**: `store.go`, `dedup.go`, `backfill.go`, `cmd/sentinel/commands/pattern.go`

### 2a. `internal/patterns/store.go` — Add per-method ValidateDB + nil Pattern guard

```go
// Add to Create, List, Search, Get:
func (s *PatternStore) Create(p *Pattern) (string, error) {
    if err := sqlite.ValidateDB(s.db, "pattern-store.Create"); err != nil {
        return "", err
    }
    if p == nil {
        return "", errors.New("pattern-store.Create: nil pattern")
    }
    // ... rest unchanged
}
```

**Test**: Add `TestCreate_NilDB`, `TestCreate_NilPattern`, `TestList_NilDB`, `TestSearch_NilDB`, `TestGet_NilDB` — construct store with nil db, verify `errors.Is(err, sqlite.ErrNilDB)`.

### 2b. `internal/patterns/dedup.go` — Add ValidateDB to FindSimilar

```go
func (s *PatternStore) FindSimilar(title string, tags []string) ([]Pattern, error) {
    if err := sqlite.ValidateDB(s.db, "pattern-store.FindSimilar"); err != nil {
        return nil, err
    }
    // ... rest unchanged
}
```

**Test**: Add `TestFindSimilar_NilDB`.

### 2c. `internal/patterns/backfill.go` — Add ValidateDB to BackfillFrom* methods

```go
func (s *PatternStore) BackfillFromCognitiveDNA(baseDir string) (BackfillResult, error) {
    if err := sqlite.ValidateDB(s.db, "pattern-store.BackfillFromCognitiveDNA"); err != nil {
        return BackfillResult{}, err
    }
    // ... rest unchanged
}
```
Same pattern for `BackfillFromEvolutionInsights` and `BackfillFromSentinelLog`.

**Test**: Add `TestBackfillFromCognitiveDNA_NilDB`, `TestBackfillFromEvolutionInsights_NilDB`, `TestBackfillFromSentinelLog_NilDB`.

### 2d. `cmd/sentinel/commands/pattern.go` — Add ValidateDB to NewPatternCmd

```go
func NewPatternCmd(db *sqlite.DB) *cobra.Command {
    if err := sqlite.ValidateDB(db, "pattern-cmd"); err != nil {
        // Return command that surfaces error on execution
        cmd := &cobra.Command{Use: "pattern", Short: "..."}
        cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
        return cmd
    }
    // ... rest unchanged
}
```

**Test**: Add `TestNewPatternCmd_NilDB` — verify command returns ErrNilDB on execution.

---

## Chapter 3: Error Handling Fixes

**Severity**: Major (3 comments)
**Files**: `pattern.go`, `backfill.go`

### 3a. `cmd/sentinel/commands/pattern.go:61` — Handle FindSimilar error

```go
// BEFORE
similar, _ := store.FindSimilar(addTitle, tagSlice)

// AFTER
similar, err := store.FindSimilar(addTitle, tagSlice)
if err != nil {
    return fmt.Errorf("pattern add: dedup check failed: %w", err)
}
```

### 3b. `cmd/sentinel/commands/pattern.go:218` — Replace os.Exit(1) with return error

```go
// BEFORE
p, err := store.Get(args[0])
if err != nil {
    fmt.Fprintf(os.Stderr, "Pattern not found: %s\n", args[0])
    os.Exit(1)
}

// AFTER
p, err := store.Get(args[0])
if err != nil {
    return fmt.Errorf("pattern get: pattern not found: %s", args[0])
}
```

**Test**: Add `TestPatternGetCmd_NotFound` — verify error returned (no os.Exit).

### 3c. `internal/patterns/backfill.go:89-98` — Handle Create errors in non-dry-run path

```go
// BEFORE
for _, c := range candidates {
    s.Create(&Pattern{...})
}

// AFTER
for _, c := range candidates {
    if _, err := s.Create(&Pattern{...}); err != nil {
        result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", c.Title, err))
        continue
    }
    result.Inserted++
}
```

Change `BackfillFromSentinelLog` return type to `(BackfillResult, error)` for consistency, returning the result with errors collected.

**Test**: Update existing `TestRunBackfillSentinelLog_Success` to verify BackfillResult.

---

## Chapter 4: Test Hardening — False Positive Assertions & Setup Guards

**Severity**: Major (1) + Minor (5) — 6 CodeRabbit comments
**Files**: `backfill_test.go`, `dedup_test.go`, `store_test.go`

### 4a. `backfill_test.go:171` — FP tests need assertions, not just t.Log

```go
// BEFORE
t.Logf("SHOULD NOT be a candidate: %q", c.Title)

// AFTER
t.Errorf("false positive: %q should not be a candidate", c.Title)
```

For each parser FP test, add negative assertion: verify no candidate with non-matching title exists in the results.

### 4b. `backfill_test.go:38` — Don't ignore setup errors

```go
// BEFORE
store, _ := patterns.NewPatternStore(db)

// AFTER
store, err := patterns.NewPatternStore(db)
if err != nil {
    t.Fatalf("NewPatternStore failed: %v", err)
}
```

### 4c. `backfill_test.go:110` — Assert candidates non-empty in dry-run test

```go
// AFTER dry-run call
if len(candidates) == 0 {
    t.Fatal("expected non-empty candidates from dry-run")
}
```

### 4d. `dedup_test.go:60,96` — Don't ignore setup/seed errors

Same pattern as 4b: replace `_` with proper error check + `t.Fatalf`.

### 4e. `store_test.go:92` — Don't ignore Create errors

Same pattern: check error on every `store.Create()` call.

---

## Chapter 5: Test Hardening — Capture Helpers & CG-01 Compliance

**Severity**: Minor (3) + Major (1) — 4 CodeRabbit comments
**Files**: `pattern_test.go`, `registry/commands_test.go`

### 5a. `pattern_test.go:53,69` — Defer-based restore for captureStderr/captureStdout

```go
// BEFORE
func captureStderr(t *testing.T, fn func()) string {
    old := os.Stderr
    r, w, err := os.Pipe()
    os.Stderr = w
    fn()
    w.Close()
    os.Stderr = old
    ...
}

// AFTER
func captureStderr(t *testing.T, fn func()) string {
    old := os.Stderr
    r, w, err := os.Pipe()
    if err != nil {
        t.Fatalf("pipe: %v", err)
    }
    os.Stderr = w
    defer func() { os.Stderr = old }()
    fn()
    w.Close()
    out, _ := io.ReadAll(r)
    r.Close()
    return string(out)
}
```

Same pattern for `captureStdout`.

### 5b. `pattern_test.go:107` — CG-01: strings.Contains without FP test

For `TestPatternAddCmd`, verify output does NOT contain unrelated markers:
```go
if strings.Contains(out, "Similar pattern found") {
    t.Fatal("false positive: new pattern should not trigger dedup warning")
}
```

Add FP assertions to: `TestPatternAddCmd`, `TestPatternListCmd`, `TestPatternSearchCmd`, `TestPatternGetCmd`.

### 5c. `registry/commands_test.go:16` — Race condition with global factories

Use `sync.Mutex` lock around factory manipulation, or refactor test to use a test-only registry instance. Evaluate if a `ResetForTesting()` helper is needed.

---

## Chapter 6 (Optional): Global Flag Refactor

**Severity**: Major (1 comment)
**File**: `cmd/sentinel/commands/pattern.go`

Move package-level flag vars into command constructors:
```go
// BEFORE
var addTitle string
func patternAddCmd(db *sqlite.DB) *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.Flags().StringVar(&addTitle, "title", "", ...)
}

// AFTER
func patternAddCmd(db *sqlite.DB) *cobra.Command {
    var title, desc, category, source, sourceRef, tags, impact string
    var force bool
    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // use local title, desc, etc.
        },
    }
    cmd.Flags().StringVar(&title, "title", "", ...)
    ...
}
```

**Risk**: Medium. Changes command wiring but improves test isolation.
**Recommendation**: Discuss with user. If accepted, remove `resetAddFlags()`/`resetListFlags()`/`resetBackfillFlags()` helpers from tests.

---

## Chapter 7: Documentation Fix

**Severity**: Minor (1 comment)
**File**: `docs/superpowers/specs/2026-05-11-pattern-capture-design.md:340`

Remove `--dry-run` flag from docs, or document that sentinel-log backfill is always dry-run by default.

---

## Execution Order

| Step | Chapter | Commits | CodeRabbit Comments Addressed | Risk |
|------|---------|---------|-------------------------------|------|
| 1 | Ch 1: bm25 fix | 1 | M12 | Low |
| 2 | Ch 2: Nil-DB guards | 1 | C1, M1, M6, M11 | Low |
| 3 | Ch 3: Error handling | 1 | M3, M4, M8 | Low |
| 4 | Ch 4: Test hardening (patterns) | 1 | M5, m2, m3, m4, m5, m7 | Low |
| 5 | Ch 5: Test hardening (commands) | 1 | M14, m8, m9 | Low |
| 6 | Ch 6: Flag refactor (optional) | 1 | M2 | Medium |
| 7 | Ch 7: Docs fix | 1 | m1 | Low |

**Total**: 6-7 atomic commits, each with build+v et+test+race verification before proceeding.
