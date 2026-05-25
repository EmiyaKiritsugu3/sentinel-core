# Dependency Management [PID-SENTINEL]

How Sentinel manages Go module dependencies.

## go.mod Structure

Sentinel uses Go modules (`go.mod`) with Go 1.26+. Direct dependencies:

```
github.com/go-playground/validator/v10  — struct validation
github.com/google/generative-ai-go      — Gemini AI client
github.com/google/shlex                 — safe shell argument parsing
github.com/google/uuid                  — task/entity ID generation
github.com/gorilla/websocket            — LiveView WebSocket transport
github.com/smacker/go-tree-sitter      — TS/TSX AST parsing
github.com/spf13/cobra                  — CLI framework
github.com/stretchr/testify             — test assertions
golang.org/x/sync                       — errgroup for concurrent ops
google.golang.org/api                   — Google API client
gopkg.in/yaml.v3                        — YAML config parsing
modernc.org/sqlite                      — CGO-free SQLite driver
```

## CGO-Free Requirement

Sentinel's SQLite driver is `modernc.org/sqlite`, NOT `mattn/go-sqlite3`. This is intentional:
- Zero CGO dependency — pure Go compilation
- Static binary without libsqlite3 linking
- Cross-compilation without C toolchain
- No platform-specific shared library issues

Never add `mattn/go-sqlite3` or any CGO-requiring package. The CGO-free guarantee is a hard architectural constraint.

## Upgrade Strategy

**Weekly audit:**

```bash
go list -u -m all
```

**Selective upgrades for security patches:**

```bash
go get -u golang.org/x/net@latest
go mod tidy
go test -race ./...
```

**Full upgrade (major releases only):**

```bash
go get -u ./...
go mod tidy
go test -race ./...
golangci-lint run ./...
```

Never upgrade dependencies in the same commit as feature work. Dependency bumps are standalone commits with the message `chore(deps): upgrade <package> to vX.Y.Z`.

## Adding New Dependencies

Before adding any external module:

1. Check if the Go standard library can solve the problem (prefer stdlib)
2. Verify the dependency is actively maintained (recent commits, responsive issues)
3. Confirm CGO-free compatibility (no C bindings)
4. Check license compatibility (Apache 2.0 compatible preferred)
5. Run `go mod tidy` to clean up indirect dependencies
6. Document the rationale in the commit message

Rule of thumb: Sentinel aims for minimal dependency surface. If a dependency provides less than 100 lines of value, prefer implementing it directly.

## vendoring

Sentinel does not use vendoring. `go.sum` provides checksum verification. CI fetches dependencies fresh via `go mod download`.

## Indirect Dependency Review

Before releases, review indirect dependencies for security advisories:

```bash
go list -m -json all | grep -E '"Path"|"Version"'
```

The Semgrep supply chain scan (available via CodeRabbit integration) catches known vulnerabilities in the dependency tree automatically on PRs.

## Lockfile Integrity

`go.sum` is committed to the repository. Never delete or manually edit `go.sum`. After any `go get` or `go mod tidy`, commit the updated `go.sum` alongside `go.mod`.
