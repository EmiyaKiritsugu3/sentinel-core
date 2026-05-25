# Concurrency Guide [PID-SENTINEL]

Thread-safety patterns and concurrency rules for Sentinel.

## RWMutex Pattern

All shared mutable state uses `sync.RWMutex`:

```go
type RegistryManager struct {
    mu sync.RWMutex
    db *sqlite.DB
}

func (m *RegistryManager) GetTool(name string) (Tool, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    // read...
}
```

`RLock` for read-heavy paths (snapshots, lookups). `Lock` for write paths (registration, mutation). This is the standard throughout the codebase — see `internal/agents/registry.go`, `internal/knowledge/buffer.go`, and `internal/liveview/server.go`.

## Race Detection

The `-race` flag is mandatory:

```bash
go test -race ./...
```

All tests, CI, and pre-merge checks must pass under the race detector with zero warnings. The detector catches:
- Unsynchronized map access
- Shared slice writes across goroutines
- Missing mutex in ring buffer writes
- Channel misuse patterns

A passing `go test` without `-race` is NOT sufficient evidence of thread safety.

## Global State

Sentinel has limited global state:

- `knowledge.GlobalBuffer` — process-wide singleton ring buffer (thread-safe via internal RWMutex)
- `sqlite.DB` — injected via dependency injection, not global

Do NOT add new package-level mutable globals without:
1. Wrapping in `sync.RWMutex` or `sync.Mutex`
2. Documenting the access pattern
3. Adding concurrent test coverage (`go test -race`)

## Ring Buffer Patterns

The `EventBuffer` ring buffer demonstrates concurrent-safe design:

```go
type EventBuffer struct {
    mu     sync.RWMutex
    events []SessionEvent
    max    int
    head   int
    size   int
}
```

- **Write path (Record)**: acquires `Lock`, writes to `events[head]`, advances head mod max
- **Read path (Snapshot, filter)**: acquires `RLock`, copies data before returning
- **Read path (Len)**: acquires `RLock`, returns `size`

All read methods return copies of data, not references to internal slices. This prevents external mutation of buffer state.

## Channel Communication

Prefer channels over shared memory for goroutine coordination:

```go
// LiveView server: non-blocking broadcast
select {
case s.broadcast <- event:
default:
    // drop event if client is slow — don't block the producer
}
```

The non-blocking `select` with `default` prevents slow consumers from stalling producers. This pattern is used in LiveView WebSocket broadcasting.

## Concurrency Testing

Test concurrent access explicitly:

```go
func TestConcurrentRecordAndRead(t *testing.T) {
    b := NewEventBuffer(500)
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                b.Record(SessionEvent{...})
            }
        }(i)
    }
    // ...readers...
    wg.Wait()
}
```

For CGO resources (Tree-sitter parsers), use `sync.Pool` instead of raw sharing:

```go
var parserPool = sync.Pool{
    New: func() interface{} { return sitter.NewParser() },
}
```

CGO-managed memory is NOT thread-safe. `sync.Pool` provides goroutine-local parser instances, eliminating contention on C-level state.

## Concurrency Checklist

- [ ] Mutex covers all shared state access (read AND write)
- [ ] `go test -race ./...` passes with zero warnings
- [ ] Ring buffers return copies, not internal references
- [ ] Channels use non-blocking sends with `default` for broadcast
- [ ] CGO objects accessed from single goroutine (via `sync.Pool`)
- [ ] `t.Parallel()` used only with isolated state (per-test databases)
