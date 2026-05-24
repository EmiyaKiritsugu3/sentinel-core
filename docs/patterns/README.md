# Patterns — Workflow Taxonomy

Systematic catalog of development patterns mined from Sentinel Core project history.
Failure patterns → prevention rules. Success patterns → repeatable practices.

## Structure

```
docs/patterns/
  catalog.md     # The pattern catalog (this is the source of truth)
  template.md    # Template for adding new patterns
```

## Categories

| Prefix | Domain | Focus |
|--------|--------|-------|
| SEC | Security | Path traversal, injection, auth bypass |
| SAF | Safety | Panics, null derefs, bounds errors |
| CON | Concurrency | Race conditions, deadlocks, parallel state |
| WEB | Web/Frontend | Protocol, polling, UX, rendering |
| QUA | Quality | Linting, docs, code hygiene |
| ARC | Architecture | Rebase, dependency, structure |

## How to Add a Pattern

1. Did something break or surprise you?
2. Categorize it by domain (SEC/SAF/CON/WEB/QUA/ARC)
3. Fill in the template from `template.md`
4. Add to `catalog.md` under the appropriate category
5. Commit with message: `docs(patterns): add [PREFIX-NNN] pattern name`

## Usage

- **Pre-commit:** Skim relevant categories before starting work
- **Code review:** Check if any existing pattern applies to the changes
- **Post-mortem:** After any bug — immediately add to catalog

## Current Count

11 patterns seeded from project history (May 2026).
