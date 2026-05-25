# Deployment Guide [PID-SENTINEL]

Building and deploying Sentinel as a static binary.

## Building

Sentinel compiles to a single static binary with CGO disabled:

```bash
CGO_ENABLED=0 go build -o sentinel ./cmd/sentinel/
```

The resulting binary has zero runtime dependencies — no `node_modules`, no shared libraries, no external runtimes.

Verify the build is truly static:

```bash
ldd sentinel
# Expected: "not a dynamic executable"
```

## Environment Variables

| Variable | Required | Purpose |
|----------|----------|---------|
| `GOOGLE_API_KEY` | No (soft) | Gemini API access for AI features |
| `SENTINEL_HOME` | No | Override default `.sentinel` directory |

When `GOOGLE_API_KEY` is absent, Sentinel operates in local-first mode: graph operations, task management, and audits work without AI. Classification and generation features degrade gracefully to heuristics.

## Knowledge Directory

Sentinel creates `.sentinel/` in the project root on first run. This directory contains:

```
.sentinel/
  graph.db          # SQLite database (WAL mode)
  graph.db-wal      # Write-Ahead Log
  graph.db-shm      # Shared memory (WAL index)
```

The `sentinel` binary and `.sentinel/` directory should travel together — copy both to deploy.

## CI Deployment

GitHub Actions workflow (`.github/workflows/sonarcloud.yml`) handles CI. For custom deploy pipelines:

```bash
# Build
CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/sentinel ./cmd/sentinel/

# Verify
./dist/sentinel --version

# Run tests with coverage
go test -coverprofile=coverage.out -race ./...
go tool cover -func=coverage.out

# Lint
golangci-lint run ./...
```

The `-ldflags="-s -w"` strip debug symbols and DWARF tables, reducing binary size ~30%.

## Platform Support

Sentinel is tested on Linux (primary). Cross-compilation targets:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o sentinel-linux ./cmd/sentinel/
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o sentinel-darwin ./cmd/sentinel/
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o sentinel.exe ./cmd/sentinel/
```

## Startup Sequence

1. `sqlite.Init()` — creates `.sentinel/` directory, opens `graph.db`
2. Applies pragmas: WAL mode, foreign keys, busy timeout (5000ms), synchronous NORMAL
3. `graph.Migrate()` — runs schema migrations in a transaction
4. CLI command executes with injected `*sqlite.DB`

Startup failures emit to `stderr` with `os.Exit(1)`. The `defer db.Close()` in `main.go` ensures clean shutdown.
