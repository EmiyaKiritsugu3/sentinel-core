# Error Handling Patterns [PID-SENTINEL]

Standard error handling conventions for Sentinel.

## Error Wrapping with %w

Every error propagated upward must carry context using `fmt.Errorf` with `%w`:

```go
if err != nil {
    return nil, fmt.Errorf("registry: failed to query specialists: %w", err)
}
```

The `%w` verb preserves the original error for `errors.Is()` and `errors.As()` matching. Never use `%v` for error wrapping — it discards the error chain.

**Correct:**

```go
return fmt.Errorf("migrate: could not begin transaction: %w", err)
```

**Incorrect:**

```go
return fmt.Errorf("migrate: could not begin transaction: %v", err)
return err  // lost context
return errors.New("something failed")  // no chain
```

## Scope Prefix Convention

Error messages follow the pattern `<package>: <context>: %w`:

- `sqlite:` — database connection, pragma errors
- `migrate:` — schema migration errors
- `state:` — task manager operations
- `registry:` — specialist selection/lookup
- `graph:` — graph engine operations

This enables grep-based debugging: search for `grep "migrate:"` to find all migration error sites.

## ErrNilDB Sentinel

The typed sentinel error `sqlite.ErrNilDB` replaces ad-hoc nil checks:

```go
var ErrNilDB = errors.New("nil db")

func ValidateDB(db *DB, caller string) error {
    if db == nil || db.Conn == nil {
        return fmt.Errorf("%s: %w", caller, ErrNilDB)
    }
    return nil
}
```

Callers match with:

```go
if errors.Is(err, sqlite.ErrNilDB) {
    // handle nil DB case
}
```

Every constructor with a `*sqlite.DB` parameter must call `ValidateDB`:

```go
func NewManager(db *sqlite.DB) (*Manager, error) {
    if err := sqlite.ValidateDB(db, "state-manager"); err != nil {
        return nil, err
    }
    return &Manager{db: db}, nil
}
```

## ValidateDB Pattern

The pattern is:
1. Constructor receives `*sqlite.DB`
2. First line calls `sqlite.ValidateDB(db, "<component-name>")`
3. If nil, returns wrapped `ErrNilDB`
4. Tests verify nil-rejection: `NewManager(nil)` must error

This provides systematic nil-guard coverage across all 6+ components that depend on the database.

## Error Discard Rules

Never silence errors with `_` except:

1. `defer func() { _ = rows.Close() }()` — standard Go pattern for SQL result sets
2. `defer func() { _ = tx.Rollback() }()` — rollback after commit is a no-op
3. `_ = db.Close()` in deferred main cleanup

All other error discards require explicit `//nolint:errcheck` with justification comment.

## Test Assertions for Errors

```go
func TestNewManager_NilDB(t *testing.T) {
    _, err := state.NewManager(nil)
    if err == nil {
        t.Fatal("expected error for nil db")
    }
    if !errors.Is(err, sqlite.ErrNilDB) {
        t.Errorf("expected ErrNilDB, got %v", err)
    }
}
```

Always assert that the specific sentinel error is returned, not just that any error is returned.
