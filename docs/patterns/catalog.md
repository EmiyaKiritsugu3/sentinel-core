# Pattern Catalog — Sentinel Core

Taxonomy of workflow failures and successes mined from project history.
Each pattern = one concrete bug or design mistake that actually happened here.

---

## Security

### SEC-001: Path Traversal via User-Supplied File Paths

**What happened:** `handleGetCode` and `handleGetADR` accepted raw `os.ReadFile(filePath)` with user-supplied `?path=` query param. Any path on the filesystem could be read.

**Detection:** API handler that passes query params directly to file I/O. `os.Open`, `os.ReadFile`, `ioutil.ReadFile` with untrusted input.

**Fix:**
```go
// Reject absolute paths and traversal
path := filepath.Clean(rawPath)
if filepath.IsAbs(path) || strings.HasPrefix(path, "..") {
    http.Error(w, "invalid path", http.StatusBadRequest)
    return
}
// For directory-restricted reads, add containment check:
absPath, _ := filepath.Abs(filepath.Join(adrDir, filename))
if !strings.HasPrefix(absPath, adrDir+string(filepath.Separator)) {
    // Path escaped containment
}
```

**Real example:** PR #24 code review finding. Fixed in `0f5a69a`.

**Rule:** Every user-supplied file path goes through `filepath.Clean` + traversal rejection BEFORE any file I/O. Directory-restricted reads also need containment check.

---

### SEC-002: Path Normalization Bypass (Double-Slash)

**What happened:** `handleGetADR` extracted filename from URL path via `strings.TrimPrefix`. A request to `/api/adr//etc/passwd` produced filename `etc/passwd`, which `filepath.Join` with ADR directory resolved to absolute path, bypassing the directory containment check.

**Detection:** URL path parsing that doesn't strip leading slashes before building filesystem paths. `strings.TrimPrefix` with single `/` on paths containing `//`.

**Fix:**
```go
filename := r.URL.Path[len("/api/adr/"):]
// Reject paths with leading slash (double-slash attack)
if strings.HasPrefix(filename, "/") {
    http.Error(w, "invalid filename", http.StatusBadRequest)
    return
}
```

**Real example:** PR #24 code review finding. Fixed in `0f5a69a`.

**Rule:** After extracting filename from URL: strip ALL leading slashes (not just trim prefix), reject if absolute, THEN build filesystem path. Never trust URL path parsing alone.

---

## Safety

### SAF-001: Slice Bounds Arithmetic Overflow

**What happened:** `handleGetCode` computed `start = len(allLines) + 1` to return "empty" when `start > len(allLines)`, but then sliced `allLines[start:end]` — panicking when `start = len + 1`.

**Detection:** Arithmetic on slice lengths used as bounds. Adding/subtracting from `len()` without clamping to `[0, len]`.

**Fix:**
```go
// Clamp, never exceed
if start > len(allLines) {
    start = len(allLines)
}
if end > len(allLines) {
    end = len(allLines)
}
if start >= end {
    // Return empty
    return
}
```

**Real example:** PR #24 code review finding. Fixed in `0f5a69a`.

**Rule:** Slice bounds arithmetic: always clamp result to `[0, len(slice)]`. Never use `len + N` as a bound. Return early for empty file/slice before any index arithmetic.

---

### SAF-002: Null Guards on External Data Sources

**What happened:** `useGraphFilter` accessed `node.data()` without null check, causing crashes when cytoscape returned undefined data. `FilterToolbar` dereferenced `node.data('package')` on nodes without that field. `EventLog` received WebSocket messages with unexpected shape.

**Detection:** Accessing `.data()` or similar external data without optional chaining or null guard. TypeScript `as` casts on external data.

**Fix:**
```typescript
// Before:
const d = node.data();
packages.add(d.package as string);

// After:
const d = node.data();
if (d && d.package) {
    packages.add(d.package);
}
```

**Real example:** `ead5399` fix: null guards on node.data(), Invalid Date defense, deduplicate extractPackages.

**Rule:** External data (cytoscape, WebSocket, API response) gets null/type guard at boundary. Never trust shape — validate before use.

---

## Concurrency

### CON-001: t.Parallel() with Process-Global State

**What happened:** Mock engine tests used `t.Parallel()`, but test setup changed the process-global working directory. Multiple parallel tests raced on `os.Chdir`, causing spurious failures.

**Detection:** `t.Parallel()` in tests that call `os.Chdir`, modify global variables, or use shared in-process state.

**Fix:**
```go
// Remove t.Parallel() from tests that touch process-global state.
// Or: use test-local directories, never chdir at process level.
```

**Real example:** `ee147cf` fix: remove t.Parallel() from mock engine tests to solve process-global directory chdir race.

**Rule:** `t.Parallel()` only with local state (temp dirs, isolated fixtures). Never with `os.Chdir`, global vars, or shared in-process resources.

---

## Web / Frontend

### WEB-001: Hardcoded Protocol in URLs

**What happened:** Polling and WebSocket connections used hardcoded `http://` protocol. When served over HTTPS (production), mixed content errors blocked connections. When served over `file://` (local), connections failed silently.

**Detection:** String literals containing `http://` in frontend code that constructs API URLs.

**Fix:**
```typescript
// Before:
const baseUrl = `http://${host}`;

// After:
const baseUrl = `${window.location.protocol}//${host}`;
```

**Real example:** `ada9a03` fix: use location.protocol, remove double immediate poll.

**Rule:** URL construction uses `window.location.protocol`, never hardcoded `http://` or `https://`.

---

### WEB-002: setInterval for Polling Causes Overlap

**What happened:** `setInterval(fetch, 2000)` queued new polls even when previous hadn't completed. Under network latency, polls stacked up, causing stale state and redundant requests.

**Detection:** `setInterval` with async callback that performs network requests.

**Fix:**
```typescript
// Before: setInterval(fetchStatus, 2000);

// After: recursive setTimeout — schedule next only after current completes
function poll() {
    fetchStatus().finally(() => setTimeout(poll, 2000));
}
poll();
```

**Real example:** `46267ed` fix: recursive setTimeout, loading UX, cy.off guard.

**Rule:** Polling uses `setTimeout` recursively, not `setInterval`. Schedule next poll AFTER current completes.

---

### WEB-003: Scroll Direction UX Before User Testing

**What happened:** EventLog scrolled upward (newest at bottom), requiring manual scroll to see new events. Expected behavior: newest at top, auto-scrolls. Fixed after user reported it, not before.

**Detection:** Scroll container with reversed `flex-direction` or `order` that contradicts user expectation of "newest first."

**Fix:** Test scroll behavior with real event flow before shipping. Default to newest-at-top unless explicitly designed otherwise.

**Real example:** EventLog scroll direction fix in Sprint 2. Discovered during user testing, should have been caught earlier.

**Rule:** Scroll containers: newest content at top by default. Verify with simulated event flow before shipping.

---

## Quality

### QUA-001: Document Before Merge, Not After

**What happened:** Multiple PRs added docstrings in follow-up commits AFTER merge, not during feature development. Required extra PRs just for documentation.

**Detection:** Merge commits followed shortly by "docs: add docstrings" commits.

**Fix:** Add exported symbol documentation during feature development. Treat `golangci-lint` (revive:exported) as blocking — docstrings are part of the feature, not an afterthought.

**Real example:** `46267ed` and `ca24f19` both added docstrings post-merge for Sprint 1.

**Rule:** Documentation is part of the feature PR. No merge without complete exported symbol docs.

---

### QUA-002: Linter Cleanup Before Feature Work

**What happened:** Phase 2.11 spent significant time cleaning up 137 linter issues accumulated over previous phases. Delayed feature delivery.

**Detection:** `golangci-lint run` returning non-zero on main branch.

**Fix:** Run linter as pre-commit hook or CI gate. Never let debt accumulate.

**Real example:** Phase 2.11: "Linter Cleanup & Quality Firewall" — 137 issues → 0.

**Rule:** Linter runs on every commit. Debt is compound interest — clean as you go.

---

### QUA-003: Pre-Implementation Audit — Never Ship Without Auditing

**What happened:** Session Debrief plan audit (2026-05-24) found 3 critical issues before a single line was written: singleton in wrong package (test isolation broken), missing nil DB check (panic path), shared state in tests (flaky). Catching these in the plan saved hours of debugging.

**Detection:** Starting implementation without validating: dependencies, security surface, codebase consistency, edge case coverage, test isolation.

**Fix:** Run this checklist before implementing ANY plan:
```markdown
1. Dependencies — new packages? already in go.mod? version compatible?
2. Security — path traversal? injection? input validation bypass?
3. Consistency — follows codebase DI pattern? nil guards? error wrapping?
4. Edge cases — empty buffer? concurrent access? graceful degradation?
5. Tests — isolated (no shared singletons)? cover main flow? cover errors?
6. Types — signatures match across files? imports correct?
```

**Real example:** 2026-05-24 Session Debrief audit. Found `eventBuffer` singleton in `commands/` (should be in `knowledge/`), missing `ValidateDB` check, flaky singleton test. All fixed in plan before implementation.

**Rule:** Audit every implementation plan with this checklist. 5 minutes of audit saves hours of debugging. If you skip audit, you're betting against the pattern catalog.

---

## Architecture

### ARC-001: Rebase Feature Branch Before Merging

**What happened:** PR #24 had merge conflicts with main because main advanced while the branch was open. Required local rebase, conflict resolution, and force-push.

**Detection:** GitHub merge button greyed out with "This branch has conflicts that must be resolved."

**Fix:**
```bash
git checkout main && git pull
git checkout feature-branch
git rebase main
# resolve conflicts
git push --force-with-lease
```

**Real example:** PR #24 `feature/sprint-3-interactive-c4` conflicted with main. Resolved via rebase.

**Rule:** Before merging: rebase on latest main, resolve conflicts, force-push. Then merge. Prevents messy merge commits.

---

*Catalog seeded: 2026-05-24. Last updated: 2026-05-24.*
*12 patterns from project history. Categories: Security (2), Safety (2), Concurrency (1), Web/Frontend (3), Quality (3), Architecture (1).*
