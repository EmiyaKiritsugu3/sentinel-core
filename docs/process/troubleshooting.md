# Troubleshooting Guide [PID-SENTINEL]

Common errors, build failures, test flakes, and database lock issues.

## Common Build Errors

### `CGO_ENABLED=0` required

```
# github.com/mattn/go-sqlite3
cgo: C compiler "gcc" not found
```

This means a CGO-dependent SQLite driver was pulled in. Sentinel uses `modernc.org/sqlite` exclusively. Check `go.mod` for accidental `mattn/go-sqlite3` and remove it.

### Undefined symbols

```
undefined: some.NewFunction
```

Run `go mod tidy` to sync indirect dependencies. If linking against a local replace directive, verify the module path in `go.mod`.

### Tree-sitter compilation

```
# github.com/smacker/go-tree-sitter
```

This package requires CGO for the Tree-sitter runtime. On platforms without a C compiler, use `CGO_ENABLED=1` for builds that include Tree-sitter. For pure-Go builds, exclude the scanner package.

## Test Failures

### Race condition failures

```
WARNING: DATA RACE
Write at 0x... by goroutine X:
  sentinel-core/internal/knowledge.(*EventBuffer).Record(...)
```

Race detector found unsynchronized access. Add `sync.RWMutex` protection or restructure goroutine ownership. Never suppress with `-race=false`.

### Test flakes

```
--- FAIL: TestConcurrentRecordAndRead (0.00s)
    buffer_test.go:213: unexpected buffer length 0
```

Flaky tests are usually timing-dependent. Replace `time.Sleep` with poll loops:

```go
deadline := time.After(5 * time.Second)
for {
    if condition {
        break
    }
    select {
    case <-deadline:
        t.Fatal("timeout waiting for condition")
    default:
        time.Sleep(10 * time.Millisecond)
    }
}
```

### Database not found

```
sqlite: could not open db: <path>
```

The `.sentinel/` directory was not created. Run `sentinel plan` or any command that initializes the DB first. In tests, always use `testutil.SetupTestDB(t)` which creates the temp directory automatically.

## Database Lock Issues

### `database is locked`

SQLite allows only one writer at a time. Causes:
- Two processes holding write transactions simultaneously
- Long-running read transaction blocking WAL checkpoint
- Missing `busy_timeout` pragma

Sentinel configures `PRAGMA busy_timeout = 5000` (5-second retry). If this is insufficient, increase the timeout or restructure the write to batch in a single transaction.

### `database table is locked`

This occurs when a transaction tries to write a table that another transaction is reading. Solution: ensure reads use `QueryRowContext` (which auto-commits) instead of long-lived read transactions. WAL mode allows concurrent readers during writes.

## Schema Errors

### `migrate: unknown table`

A column migration references a table not in `pragmaTableInfo`. Add the table name to the map in `internal/graph/schema.go`.

### `duplicate column name`

A column already exists AND the `columnExistsInTx` check failed (false negative). Verify the table name is correct and the column name matches exactly (case-sensitive).

### `no such column`

Schema is out of sync with code. Verify Migrate() ran successfully. Check with:

```bash
sqlite3 .sentinel/graph.db "PRAGMA table_info(tasks);"
```

## Runtime Errors

### `nil db` / `ErrNilDB`

A component received a nil database pointer. Trace the initialization chain — check that `sqlite.Init()` succeeded and the `*DB` was properly injected into constructors. Every constructor validates with `sqlite.ValidateDB`.

### `failed to create task`

Check that Migrate() ran before task operations. The `tasks` table may not exist if the database was created but migrations were not applied.

## CI Failures

### SonarCloud scan fails

```
ERROR: Error during SonarQube Scanner execution
```

Common causes:
- `coverage.out` not generated before scan
- SonarCloud project properties mismatch
- Scanner version outdated (use `sonarsource/sonarqube-scan-action@v7.1.0`)

### golangci-lint timeout

```
level=error msg="Timeout exceeded: try increasing it by passing --timeout option"
```

Increase timeout in `.golangci.yml`: `run.timeout: 10m`. Or run linters individually to isolate the slow one.
