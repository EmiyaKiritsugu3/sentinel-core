# Performance Benchmarks [PID-SENTINEL]

Benchmarking patterns and performance analysis for Sentinel.

## Running Benchmarks

```bash
go test -bench=. ./internal/math/ -benchmem -count=5
```

`-benchmem` reports allocations per operation. `-count=5` runs each benchmark 5 times for stable results.

Current benchmark suite in `internal/math/bench_test.go` covers:
- `CalculateDelta` (HighGain, LowGain)
- `CalculateTrustScore` (ZeroHistory, PerfectRecord, LargeDataset)
- `TrustToDynamicLambda`
- `CalculateDivergence` (Stable, HighDivergence, ZeroPrevious)
- `CalculateLambda` (Normal, ZeroThought, LargeTokens)

## SQLite Query Performance

Key patterns for performant SQLite access:

**Use transactions for writes:**

```go
tx, _ := db.Conn.BeginTx(ctx, nil)
defer func() { _ = tx.Rollback() }()
// multiple writes...
tx.Commit()
```

A single transaction wrapping N writes is orders of magnitude faster than N auto-committed writes. This is critical for schema migrations and batch inserts.

**WAL mode enables concurrent reads:**

Sentinel initializes with `PRAGMA journal_mode=WAL`, allowing multiple concurrent readers while a writer holds the lock. Connection pool is configured at `SetMaxOpenConns(8)`.

**Use prepared statements for repeated queries:**

For hot paths (e.g., `SelectBest` in registry), consider preparing statements once and reusing them.

**FTS5 for text search:**

The `patterns_fts` virtual table enables full-text search over patterns. Query with:

```sql
SELECT * FROM patterns_fts WHERE patterns_fts MATCH 'cognitive';
```

## Ring Buffer Performance

The `EventBuffer` uses a fixed-size ring buffer (`internal/knowledge/buffer.go`):

- **O(1)** Record — lock, write to head slot, advance head
- **O(n)** Snapshot — copies all events for immutability
- **O(n)** filter — linear scan with predicate

Buffer defaults to 1000 capacity. For high-throughput scenarios, increase capacity at creation:

```go
buf := knowledge.NewEventBuffer(10000)
```

Concurrent access uses `sync.RWMutex`: Record acquires write lock, Snapshot/filter acquire read lock. Benchmark with:

```bash
go test -bench=. ./internal/knowledge/ -benchmem
```

## Profiling

CPU profile:

```bash
go test -bench=. -cpuprofile=cpu.prof ./internal/math/
go tool pprof cpu.prof
```

Memory profile:

```bash
go test -bench=. -memprofile=mem.prof ./internal/math/
go tool pprof mem.prof
```

## Key Performance Rules

1. `modernc.org/sqlite` is pure-Go — no CGO overhead, no shared library linking
2. `SetMaxOpenConns(8)` prevents connection thrashing
3. WAL mode avoids writer-blocking-readers contention
4. Ring buffer avoids slice growth allocations under load
5. `strings.Builder` preferred over `+=` for loop concatenation
6. `sync.Pool` for CGO parser instances (Tree-sitter) — avoids repeated C memory allocation
