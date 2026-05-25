# internal/scanner

> **Note:** The scanner sub-package was merged into `internal/graph/`. All file-scanning logic lives there.

The scanning subsystem in `internal/graph/` provides a multi-language code analysis pipeline with worker pool parallelism and incremental content hashing.

## Architecture (in `internal/graph/`)

### `FileScanner` Interface
```go
type FileScanner interface {
    Scan(path string) ScanResult
    SupportedExtensions() []string
}
```

### `GoScanner`
Uses `go/ast` and `go/parser` to parse `.go` files. Extracts:
- **Nodes**: file, function declarations, method receivers, struct and interface types
- **Edges**: imports, function calls, method invocations on struct receivers

Each node receives a stable ID (e.g., `file:path/to/file.go`, `func:path/to/file.go:FuncName`).

### `TreeSitterScanner`
Multi-language scanner using Tree-sitter grammars. Supports additional languages beyond Go by delegating to Tree-sitter's incremental parsing.

### Engine Integration

The `graph.Engine` registers scanners by file extension:
```go
engine.RegisterScanner(graph.NewGoScanner())
engine.RegisterScanner(graph.NewTreeSitterScanner())
```

During `ScanProject`, an 8-worker goroutine pool processes files concurrently. Each file undergoes:
1. Content hash calculation
2. Hash comparison against existing DB entry (skip if unchanged)
3. Scanner execution
4. Transaction-per-file persistence (nodes upsert, edges insert or ignore, stale symbol pruning)

The linker phase (`LinkDependencies`) runs after all files are scanned.

## Usage

```go
engine, _ := graph.NewEngine(db)
engine.RegisterScanner(graph.NewGoScanner())
engine.RegisterScanner(graph.NewTreeSitterScanner())
if err := engine.ScanProject(ctx, "."); err != nil { /* handle */ }
```
