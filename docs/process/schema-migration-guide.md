# Schema Migration Guide [PID-SENTINEL]

How Sentinel's SQLite schema evolves safely across versions.

## Migrate() Function

Entry point: `graph.Migrate(ctx context.Context, db *sqlite.DB) error` in `internal/graph/schema.go`.

Execution flow:

1. `sqlite.ValidateDB(db, "migrate")` — nil guard
2. `db.Conn.BeginTx(ctx, nil)` — atomic transaction
3. `tx.ExecContext(ctx, schema)` — applies the base DDL with `IF NOT EXISTS`
4. Column migrations loop — guarded by `columnExistsInTx`
5. Column renames — e.g., `specialist_id` → `agent_name`
6. Specialist seeding — `INSERT OR IGNORE` for idempotent defaults
7. `tx.Commit()` — all-or-nothing

On any error, `defer tx.Rollback()` ensures partial state is discarded.

## IF NOT EXISTS

All CREATE statements use `IF NOT EXISTS`:

```sql
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    ...
);
CREATE VIRTUAL TABLE IF NOT EXISTS patterns_fts USING fts5(...);
CREATE TRIGGER IF NOT EXISTS patterns_ai AFTER INSERT ON patterns BEGIN ... END;
CREATE INDEX IF NOT EXISTS idx_session_events_session_id ON session_events(session_id);
```

This makes `Migrate()` idempotent — running it multiple times on the same database produces no errors. Verified by `TestMigrate_Idempotent`.

## columnExistsInTx

Column existence is checked via `PRAGMA table_info()` within the active transaction:

```go
func columnExistsInTx(ctx context.Context, tx *sql.Tx, table, column string) (bool, error) {
    query, ok := pragmaTableInfo[table]
    if !ok {
        return false, fmt.Errorf("migrate: unknown table %q", table)
    }
    rows, err := tx.QueryContext(ctx, query)
    // scan cid, name, type, notnull, dflt_value, pk
    // return true if name == column
}
```

This avoids string-matching on SQLite error messages (which is fragile across driver versions). It also enables safe ALTER TABLE — columns are only added if they don't already exist.

## pragmaTableInfo Map

Hardcoded lookup prevents dynamic SQL construction (SonarCloud rule S2077):

```go
var pragmaTableInfo = map[string]string{
    "tasks":       "PRAGMA table_info(tasks)",
    "agent_trust": "PRAGMA table_info(agent_trust)",
    // ...all known tables...
}
```

Unknown table names return an error immediately, preventing SQL injection vectors.

## Column Migration Pattern

Adding new columns to existing tables:

```go
for _, m := range []colMig{
    {"tasks", "latency_ms", "ALTER TABLE tasks ADD COLUMN latency_ms REAL DEFAULT 0;"},
    {"tasks", "tokens_used", "ALTER TABLE tasks ADD COLUMN tokens_used INTEGER DEFAULT 0;"},
} {
    exists, _ := columnExistsInTx(ctx, tx, m.table, m.column)
    if !exists {
        tx.ExecContext(ctx, m.sql)
    }
}
```

## Column Rename Pattern

Renaming columns while preserving data:

```go
oldExists, _ := columnExistsInTx(ctx, tx, "agent_trust", "specialist_id")
if oldExists {
    tx.ExecContext(ctx, "ALTER TABLE agent_trust RENAME COLUMN specialist_id TO agent_name;")
}
```

SQLite's `ALTER TABLE ... RENAME COLUMN` preserves existing data. The column-existence guard makes the migration idempotent.

## Adding a New Table

1. Add `CREATE TABLE IF NOT EXISTS ...` to the `schema` constant
2. Add the table to the `pragmaTableInfo` map
3. Update `TestMigrate` table list to verify creation
4. Run `TestMigrate_Idempotent` to confirm repeatability
5. Add alter-viewer and committed-tx test cases if the table has migration-sensitive columns
