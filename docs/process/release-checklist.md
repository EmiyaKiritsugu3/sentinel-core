# Release Checklist [PID-SENTINEL]

Pre-release verification steps for Sentinel.

## 1. Test Suite

All tests must pass under race detection:

```bash
go test -race ./...
```

Verify coverage:

```bash
go test -coverprofile=coverage.out -race ./...
go tool cover -func=coverage.out | grep total
```

Coverage must remain at or above the SonarCloud minimum (50%). Target: 80%+ for releases.

## 2. Lint

Zero linter violations:

```bash
golangci-lint run ./...
```

Configured linters in `.golangci.yml`:
- `gocyclo` (max complexity 15)
- `revive/exported` (all exported symbols documented)
- `noctx` (no bare DB calls without context)
- `errcheck` (no discarded errors)
- `gosec` (security rules)
- `govet`, `staticcheck`, `ineffassign`, `unused`, `misspell`, `bodyclose`, `unconvert`, `errname`

## 3. SonarCloud Quality Gate

SonarCloud scan runs automatically on PRs and `main` branch. Before release, verify:
- 0 open issues (S3776, S1192, S8209, etc.)
- Coverage >= minimum threshold (50%)
- No new code smells or bugs on main branch

Configuration: `sonar-project.properties`. Coverage report: `coverage.out`.

## 4. CodeRabbit Review

Every release-candidate PR must pass CodeRabbit automated review. Address:
- Error handling gaps (missing nil checks, unhandled errors)
- Edge case coverage
- Security concerns (path traversal, injection)
- Code quality (complexity, duplication)

The `.github/copilot-instructions.md` configures review focus areas.

## 5. Changelog

Update `docs/process/sentinel-log.md` with milestone entry:

```markdown
## [YYYY-MM-DD] Milestone: Feature Name [PID-SENTINEL-XXXX]

**Status**: COMPLETED
**Impact**: HIGH/MEDIUM/LOW

### Analysis (Epiphanies)
1. Finding 1
2. Finding 2

### Metrics
| Check | Before | After |
|---|---|---|
| ... | ... | ... |

### Key Learning
"..."
```

## 6. Build Verification

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/sentinel ./cmd/sentinel/
./dist/sentinel --version
ldd dist/sentinel  # must output "not a dynamic executable"
```

## 7. Final Verification

```bash
go vet ./...
go test -count=1 -race ./...
```

The `-count=1` flag disables test caching, ensuring a clean run. If any test is flaky, it must be fixed before release — never ship tests that pass only on retry.

## 8. Git Tag

After merge to `main`:

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z: <short description>"
git push origin vX.Y.Z
```

Semantic versioning: major for breaking changes, minor for features, patch for fixes.
