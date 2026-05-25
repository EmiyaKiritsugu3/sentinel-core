# pkg/sqlite

SQLite database wrapper with WAL-mode initialization, connection pooling, and nil-guard validation.

## Overview

The sqlite package provides a thin wrapper around `database/sql` for the CGo-free `modernc.org/sqlite` driver. It configures performance pragmas, connection pooling, and exposes a `ValidateDB` sentinel used by every other package.

## Key Types

### `DB`
```go
type DB struct {
    Conn *sql.DB
}
```
Wraps `*sql.DB` for method attachment. All packages access the database through this type.

### `Init() (*DB, error)`
Initializes the database at the default path `.sentinel/graph.db`. Creates parent directory if missing (`0750`).

### `InitAtPath(dbPath string) (*DB, error)`
Same as `Init` but with a custom path.

Configuration applied:
- `PRAGMA journal_mode=WAL` — write-ahead logging for concurrent readers
- `PRAGMA foreign_keys=ON` — referential integrity enforcement
- `PRAGMA busy_timeout=5000` — 5-second busy wait before SQLITE_BUSY
- `PRAGMA synchronous=NORMAL` — balance between durability and performance
- `SetMaxOpenConns(8)`, `SetMaxIdleConns(8)` — connection pooling for WAL mode
- `db.PingContext(ctx)` — connection health check before returning

### `Close() error`
Closes the underlying database connection.

## Validation

### `ErrNilDB`
```go
var ErrNilDB = errors.New("nil db")
```
Sentinel error for nil database detection. All packages check `errors.Is(err, sqlite.ErrNilDB)` for graceful nil handling.

### `ValidateDB(db *DB, caller string) error`
Returns `"<caller>: nil db"` if `db == nil || db.Conn == nil`. Used by every constructor and method that requires a database handle:

```go
if err := sqlite.ValidateDB(db, "engine"); err != nil {
    return nil, err
}
```

This pattern ensures systematic nil-guard hardening across all 14 internal packages.

## Dependencies

- `modernc.org/sqlite` — CGo-free SQLite driver (imported for side effects)
- `database/sql` — standard library

## Usage

```go
db, err := sqlite.Init()
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Validate before use
if err := sqlite.ValidateDB(db, "my-component"); err != nil {
    return err
}

// Use directly
rows, err := db.Conn.QueryContext(ctx, "SELECT ...", args...)
```
