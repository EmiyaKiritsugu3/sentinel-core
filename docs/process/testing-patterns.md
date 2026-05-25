# Testing Patterns [PID-SENTINEL]

Conventions and patterns for writing tests in Sentinel.

## Database Isolation

Every test that touches SQLite must use isolated databases via `testutil.SetupTestDB`:

```go
func TestMyFeature(t *testing.T) {
    t.Parallel()
    db := testutil.SetupTestDB(t)
    defer func() { _ = db.Close() }()
    // ...
}
```

`SetupTestDB` creates a temporary database via `t.TempDir()`, ensuring no cross-test contamination. Each test gets its own SQLite file, independent WAL journal, and isolated pragma state.

Never hardcode `graph.db` or `.sentinel` paths in tests — use `testutil.SetupTestDB` exclusively.

## Race Detection

All tests must pass under the race detector:

```bash
go test -race ./...
```

This is enforced in CI. The `-race` flag catches data races in shared state (`sync.RWMutex`, channels, global variables). In CI, the same command gates merges — zero warnings tolerated.

## Parallel Execution

Prefer `t.Parallel()` in tests that use isolated databases:

```go
func TestMigration_XYZ(t *testing.T) {
    t.Parallel()
    // Each parallel test owns its DB via SetupTestDB
}
```

Tests that share global state (e.g., `GlobalBuffer`) should NOT use `t.Parallel()` unless the shared state is read-only during the test.

## Nil DB Testing

Every component that accepts `*sqlite.DB` must have a nil-DB test:

```go
func TestMigrate_NilDB(t *testing.T) {
    t.Parallel()
    err := Migrate(context.Background(), nil)
    if err == nil {
        t.Fatal("expected error for nil db")
    }
}
```

This validates that `sqlite.ValidateDB` is called at every entry point. Pattern: pass `nil`, assert non-nil error.

## Closed DB Testing

Test behavior when a DB connection is closed before use:

```go
func TestMigrate_ClosedDB(t *testing.T) {
    t.Parallel()
    db := testutil.SetupTestDB(t)
    _ = db.Conn.Close()
    err := Migrate(context.Background(), db)
    if err == nil {
        t.Fatal("expected error for closed db connection")
    }
}
```

## Global State Isolation

`knowledge.GlobalBuffer` is a process-wide singleton. Tests that depend on its state should either:
1. Reset it before the test, or
2. Create a local `NewEventBuffer(N)` instance

Global buffer tests belong in `internal/knowledge/buffer_test.go` and should not rely on ordering with other test packages.

## Benchmarks

Benchmark files follow `*_test.go` convention in their respective packages. See `internal/math/bench_test.go` for examples using `testing.B`:

```bash
go test -bench=. ./internal/math/ -benchmem
```

## Error Path Coverage

Every test suite must cover: happy path, nil/zero input, invalid input, and error propagation. See `TestMigrate_AlterView` and `TestColumnExistsInTx_CommittedTx` for examples of exercising error branches that require specific setup.
