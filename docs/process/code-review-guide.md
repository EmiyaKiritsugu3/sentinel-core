# Code Review Guide [PID-SENTINEL]

What reviewers check and how authors respond.

## QUA-003 Pre-Implementation Audit

Before writing code, run the 6-point audit (from `AGENTS.md`):

1. **Dependencies** — new? in go.mod? version match?
2. **Security** — path traversal? injection? missing validation?
3. **Consistency** — DI pattern? nil guards? error wrapping? codebase convention?
4. **Edge cases** — empty input? concurrency? graceful degradation?
5. **Tests** — isolated (no shared singletons)? cover errors? follow existing test patterns?
6. **Types** — signatures consistent across files? imports correct?

Fix critical findings before implementing. This prevents hours of debugging downstream.

## What Reviewers Check

### Architecture

- Does the change respect package boundaries? (`internal/bridge/` vs `internal/intake/`)
- Are new interfaces defined where they're consumed, not where they're implemented?
- Does resource ownership follow the Engine lifecycle? (e.g., `genai.Client` owned by Engine, injected downstream)

### Error Handling

- Every error wrapped with context: `fmt.Errorf("scope: %w", err)`
- Never discard errors with `_` unless explicitly justified with `//nolint`
- All exported methods validate dependencies via `sqlite.ValidateDB(db, "caller")`
- Nil DB returns `ErrNilDB` sentinel, enabling `errors.Is()` matching

### Concurrency

- Shared state protected by `sync.RWMutex` or `sync.Mutex`
- No data races — verified by `go test -race ./...`
- Ring buffers and global state thread-safe
- Channels used for goroutine communication, not shared memory

### Testing

- Happy path + nil input + invalid input + error propagation
- Tests isolated via `testutil.SetupTestDB(t)` (never hardcoded paths)
- Parallel-safe when using isolated databases
- Mock strategy: interface-based injection, not package-level globals

### Code Quality

- gocyclo <= 15 per function
- All exported symbols have doc comments
- Context-aware DB calls (no bare `Query`/`Exec`/`Begin`)
- English-only code, comments, and logs

## How to Respond to Review

### For each finding:

1. **Critical/Major**: Fix in the same PR. Do not defer security, race conditions, or broken error paths.
2. **Minor**: Fix or justify with a comment. Style preferences are not blocking.
3. **False Positive**: Reply with evidence — test output, linter config, or codebase precedent.

### Format:

- Apply fix → push commit → resolve thread
- If disagreeing: reply with technical reasoning, not preference
- Never resolve a thread without action or explicit reviewer acknowledgment

## Sovereign Audit Report

Every PR includes Standard #08 audit report with 5 points:

1. **The Good** — what is now solid and correct
2. **The Bad** — technical debt consciously accepted
3. **The Ugly** — risks, fragilities, or inconsistencies
4. **The Lesson** — new universal principle or pattern extracted
5. **The Next** — identified optimization or evolution path

Reviewers validate that the Audit Report accurately reflects the code change. A report claiming "no debt" for a feature introducing new dependencies is rejected.
