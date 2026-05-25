# Contributing Guide [PID-SENTINEL]

How to contribute code to Sentinel Core.

## Branch Naming

Branch names follow the pattern `<type>/<short-description>`:

- `feature/sprint-X-feature-name` — new feature work
- `feat/short-description` — smaller feature additions
- `docs/short-description` — documentation changes
- `refactor/short-description` — refactoring without behavior change
- `fix/short-description` — bug fixes

Use kebab-case for descriptions. Keep names under 50 characters.

## Commit Format

Conventional Commits with scope prefix:

```
type(scope): terse summary

Optional body explaining "why", not "what".
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`

Scopes: `agents`, `graph`, `sqlite`, `bridge`, `state`, `knowledge`, `liveview`, `cmd`, `audit`, `math`

Example:

```
feat(agents): add Lyapunov divergence gate A.5

Tracks per-step lambda divergence. Intervenes when consecutive
steps exceed 5x increase to prevent hallucination cascades.
```

Lines wrap at 72 characters in body. Subject line max 50 characters.

## PR Template

Every pull request must complete the template at `.github/pull_request_template.md`, including:

1. **Overview** — one-sentence impact summary
2. **References** — plan path, ADR link, roadmap milestone
3. **Technical Changes** — checklist of concrete code changes
4. **Sovereign Audit Report** — 5-point Standard #08 assessment
5. **Evidence** — command output proving the change works

All sections required. PRs missing the Audit Report or Evidence block are blocked.

## Review Process

1. Open PR against `main` from feature branch
2. **CodeRabbit** runs automated review on every PR — address all critical findings
3. **SonarCloud** quality gate must pass (0 open issues, coverage >= minimum)
4. **golangci-lint** zero issues enforced
5. Human reviewer validates architecture, error handling, and test coverage
6. Merge after all gates pass

## Pre-Merge Checklist

Before requesting review, run locally:

```bash
go test -race ./...
golangci-lint run ./...
go vet ./...
```

For new SQLite schema changes, also:

```bash
go test -run TestMigrate -v ./internal/graph/
go test -run TestMigrate_Idempotent -v ./internal/graph/
```

## Code conventions

- All code, comments, docstrings, and log messages in English
- Every exported symbol (type, func, const, method) has a doc comment: `// SymbolName verb phrase.`
- gocyclo complexity <= 15 per function
- Error wrapping with context: `fmt.Errorf("context: %w", err)`
- Context-aware DB calls: use `QueryRowContext`, `ExecContext`, `QueryContext`
- No `os.Exit` or `log.Fatalf` in library code — return errors to caller

## First-Time Setup

```bash
git clone https://github.com/EmiyaKiritsugu3/sentinel-core.git
cd sentinel-core
go mod download
go build -o sentinel ./cmd/sentinel/
go test ./...
```
