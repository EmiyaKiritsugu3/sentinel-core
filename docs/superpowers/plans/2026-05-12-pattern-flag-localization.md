# Ch6: Pattern Flag Localization

**Branch:** `refactor/pattern-flag-localization` (from `main` post PR#9 merge)
**Date:** 2026-05-12
**Status:** Pending

## Problem

`cmd/sentinel/commands/pattern.go` declares 13 package-level flag variables across 3 `var` blocks. These vars are shared mutable state — any test that calls `NewPatternCmd(db)` multiple times risks flag value leakage between runs. The test file compensates with 3 `reset*Flags()` helpers that manually zero out the vars before each test case.

This is the standard Cobra anti-pattern. The recommended approach is to declare flag vars as **local variables inside the command constructor**, where they are captured by the `RunE` closure. Each call to the constructor produces a fresh command with fresh flag vars — no shared state, no reset helpers needed.

## Scope

### Affected files
- `cmd/sentinel/commands/pattern.go` — 3 var blocks (13 vars) + 3 subcommand constructors
- `cmd/sentinel/commands/pattern_test.go` — 3 reset helpers + all their call sites

### Out of scope
- `patternSearchCmd` — no flag vars (positional args only)
- `patternGetCmd` — no flag vars (positional args only)
- `NewPatternCmd` parent — no flags of its own
- `runBackfillCognitiveDNA`, `runBackfillEvolutionInsights`, `runBackfillSentinelLog` — helper functions, no flag interaction
- Any other package

## Current State

```go
// pattern.go — 3 package-level var blocks

var (                              // Block 1: add subcommand (8 vars)
    addTitle    string
    addDesc     string
    addCategory string
    addSource   string
    addSourceRef string
    addTags     string
    addImpact   string
    addForce    bool
)

var (                              // Block 2: list subcommand (3 vars)
    listCategory string
    listSource   string
    listImpact   string
)

var (                              // Block 3: backfill subcommand (2 vars)
    backfillSource string
    backfillAll    bool
)
```

```go
// pattern_test.go — 3 reset helpers (will be deleted)

func resetAddFlags()      { addTitle = ""; addDesc = ""; ... addForce = false }
func resetListFlags()     { listCategory = ""; listSource = ""; listImpact = "" }
func resetBackfillFlags() { backfillSource = ""; backfillAll = false }
```

## Target State

Each flag var becomes a local inside its subcommand constructor. The `RunE` closure captures it by reference — standard Go closure semantics.

```go
func patternAddCmd(db *sqlite.DB) *cobra.Command {
    var title, desc, category, source, sourceRef, tags, impact string
    var force bool

    cmd := &cobra.Command{
        Use: "add",
        RunE: func(cmd *cobra.Command, args []string) error {
            // references title, desc, etc. via closure
        },
    }

    cmd.Flags().StringVar(&title, "title", "", "Pattern title (required)")
    // ...
    return cmd
}
```

No package-level flag vars. No reset helpers. Each `NewPatternCmd(db)` call produces an isolated command tree.

## Implementation Steps

### Step 1: Refactor `patternAddCmd` (8 vars → local)

**Changes in `pattern.go`:**
- Delete `var ( addTitle ... addForce )` block (lines 39-48)
- Inside `patternAddCmd`, declare: `var title, desc, category, source, sourceRef, tags, impact string` + `var force bool`
- Replace all `addTitle` → `title`, `addDesc` → `desc`, `addCategory` → `category`, `addSource` → `source`, `addSourceRef` → `sourceRef`, `addTags` → `tags`, `addImpact` → `impact`, `addForce` → `force`
- Update `cmd.Flags().StringVar(&addTitle, ...)` → `cmd.Flags().StringVar(&title, ...)`
- Update `cmd.MarkFlagRequired("title")` — no change needed (flag name stays the same)

**Changes in `pattern_test.go`:**
- Remove all `resetAddFlags()` calls (6 occurrences: lines 98, 119, 123, 140, 144, 160, 200, 238)
- The tests that create multiple `NewPatternCmd(db)` calls will work because each call creates fresh local vars

### Step 2: Refactor `patternListCmd` (3 vars → local)

**Changes in `pattern.go`:**
- Delete `var ( listCategory ... listImpact )` block (lines 117-121)
- Inside `patternListCmd`, declare: `var category, source, impact string`
- Replace `listCategory` → `category`, `listSource` → `source`, `listImpact` → `impact`
- Update `cmd.Flags().StringVar(&listCategory, ...)` → `cmd.Flags().StringVar(&category, ...)`

**Changes in `pattern_test.go`:**
- Remove all `resetListFlags()` calls (2 occurrences: lines 165, 184)

### Step 3: Refactor `patternBackfillCmd` (2 vars → local)

**Changes in `pattern.go`:**
- Delete `var ( backfillSource ... backfillAll )` block (lines 242-245)
- Inside `patternBackfillCmd`, declare: `var source string` + `var all bool`
- Replace `backfillSource` → `source`, `backfillAll` → `all`
- Update `cmd.Flags().StringVar(&backfillSource, ...)` → `cmd.Flags().StringVar(&source, ...)`

**Changes in `pattern_test.go`:**
- Remove all `resetBackfillFlags()` calls (2 occurrences: lines 272, 287)

### Step 4: Delete reset helpers from test file

**Changes in `pattern_test.go`:**
- Delete `resetAddFlags()` function (lines 73-82)
- Delete `resetListFlags()` function (lines 84-88)
- Delete `resetBackfillFlags()` function (lines 90-93)

These functions reference package-level vars that no longer exist — compilation will fail if they remain.

### Step 5: Verify

```bash
go vet ./cmd/sentinel/commands/
go test ./cmd/sentinel/commands/ -count=1 -timeout 30s
go test ./... -count=1 -timeout 60s
```

## Verification Criteria

- [ ] `go vet` passes on `cmd/sentinel/commands/`
- [ ] All tests in `cmd/sentinel/commands/` pass
- [ ] No package-level flag vars remain in `pattern.go`
- [ ] No `reset*Flags()` helpers remain in `pattern_test.go`
- [ ] No `reset*Flags()` calls remain in test bodies
- [ ] Flag names (`--title`, `--desc`, etc.) unchanged — CLI interface is backward-compatible

## Risks

| Risk | Mitigation |
|------|-----------|
| Closure captures stale value | Go closures capture by reference — the `&title` in `StringVar` and the `title` in `RunE` both reference the same local var. No stale value risk. |
| Test isolation breaks | Each `NewPatternCmd(db)` creates a fresh command tree with fresh local vars. No shared state. Better isolation than before. |
| Flag name collision in list + add | `addCategory` and `listCategory` become `category` in different function scopes. No collision — they're in different constructors. |
| `listSource` shadows `addSource` | Same as above — different scopes. No issue. |

## Commit Strategy

Single commit — all 3 subcommands + test cleanup in one atomic change. The refactor is tightly coupled: removing the package-level vars breaks the reset helpers, and removing the reset helpers without localizing the vars would break flag binding. One commit keeps the tree compilable at every step if done atomically.

**Message:** `refactor(commands): localize pattern flag vars to subcommand scope`
