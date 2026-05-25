# Debugging Guide [PID-SENTINEL]

Practical techniques for diagnosing issues in Sentinel at runtime.

## SQLite Direct Inspection

Query the graph database directly with `sqlite3`:

```bash
sqlite3 .sentinel/graph.db "SELECT id, status, description FROM tasks ORDER BY created_at DESC LIMIT 10;"
sqlite3 .sentinel/graph.db "SELECT agent_name, successes, total, trust_score FROM agent_trust;"
sqlite3 .sentinel/graph.db "SELECT * FROM pragma_table_info('tasks');"
```

For JSON columns (capabilities, tags), pipe through `jq`:

```bash
sqlite3 .sentinel/graph.db "SELECT id, capabilities FROM specialist_registry;" | jq
```

Check migration state by verifying expected tables exist:

```bash
sqlite3 .sentinel/graph.db ".tables"
sqlite3 .sentinel/graph.db "PRAGMA table_info(agent_trust);"
```

## Schema Drift Detection

If `sentinel` commands fail with SQL errors, verify schema alignment:

```bash
go test -run TestMigrate ./internal/graph/ -v
go test -run TestMigrate_Idempotent ./internal/graph/ -v
```

These confirm Migrate() runs clean on fresh databases and twice without errors.

## EventBuffer Inspection

The global EventBuffer (`knowledge.GlobalBuffer`) collects session events. Inspect via:

```go
snap := knowledge.GlobalBuffer.Snapshot()
for _, e := range snap {
    log.Printf("[%s] %s: %s", e.Type, e.Domain, e.Summary)
}
```

Filtered access:

```go
knowledge.GlobalBuffer.Errors()
knowledge.GlobalBuffer.Decisions()
knowledge.GlobalBuffer.ByDomain("auth")
```

## Sentinel State Debugging

Check task state and verify Hard Gates:

```bash
sentinel status
sentinel audit
```

For CLI command debugging, set verbose output:

```bash
sentinel --verbose plan "test task" "go test ./internal/..."
```

## Common SQLite Debugging Commands

| Problem | Command |
|---------|---------|
| Verify WAL mode | `PRAGMA journal_mode;` |
| Check foreign keys | `PRAGMA foreign_keys;` |
| Busy timeout status | `PRAGMA busy_timeout;` |
| List indexes | `SELECT * FROM sqlite_master WHERE type='index';` |
| Check FTS | `SELECT * FROM patterns_fts WHERE patterns_fts MATCH 'cognitive';` |

## Log Inspection

Build failures and CI logs are the primary forensic source. Key signals:
- `migrate:` prefix — schema migration errors (column rename, ALTER TABLE failures)
- `state:` prefix — task manager operations (create, status update)
- `registry:` prefix — specialist selection failures
- `sqlite:` prefix — database connection/pragma errors
