# ADR Placeholder Cleanup Implementation Plan [PID-SENTINEL]

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace template placeholders in four ADR files with actual technical context and reasoning derived from the project's current state (Phase 5.8 "KISS Pivot", SQLite persistence, and autonomous task decomposition).

**Architecture:** We are documenting the transition from a stateless CLI to an autonomous engine with a persistent ledger. ADRs will reflect decisions on documentation evolution, self-auditing compliance, centralized persistence refactoring, and edge-case handling for special characters.

**Tech Stack:** Markdown, Sentinel Core (Go), SQLite (Ledger).

---

## Task 1: Update ADR-3c3075f2 (Complex Task: Documentation Evolution)

**Files:**
- Modify: `docs/architecture/adr/ADR-3c3075f2-complex-task-documentation-evolution.md`

- [ ] **Step 1: Write minimal implementation**

Replace the "Decisão" and "Consequências" sections with reasoning about the new `sentinel:decompose` tool and how multi-step tasks require persistent state tracking rather than just simple logs.

- [ ] **Step 2: Commit**

```bash
git add docs/architecture/adr/ADR-3c3075f2-complex-task-documentation-evolution.md
git commit -m "docs(adr): document evolution of complex task handling"
```

## Task 2: Update ADR-841fa0a2 (Self-Audit for Standard Compliance)

**Files:**
- Modify: `docs/architecture/adr/ADR-841fa0a2-self-audit-for-standard-compliance.md`

- [ ] **Step 1: Write minimal implementation**

Replace placeholders with details about the automated verification loop (STD-01, STD-05) and the `npm run sentinel audit` command.

- [ ] **Step 2: Commit**

```bash
git add docs/architecture/adr/ADR-841fa0a2-self-audit-for-standard-compliance.md
git commit -m "docs(adr): specify self-audit and compliance strategies"
```

## Task 3: Update ADR-ad9933bf (Refatorar camada de persistencia)

**Files:**
- Modify: `docs/architecture/adr/ADR-ad9933bf-refatorar-camada-de-persistencia.md`

- [ ] **Step 1: Write minimal implementation**

Replace placeholders with the decision to use SQLite as a centralized ledger to track sub-tasks, specialist assignments, and ensure atomicity in agentic workflows.

- [ ] **Step 2: Commit**

```bash
git add docs/architecture/adr/ADR-ad9933bf-refatorar-camada-de-persistencia.md
git commit -m "docs(adr): document persistence layer refactoring to SQLite"
```

## Task 4: Update ADR-fe2bb6f9 (Decisão Crítica --- special characters)

**Files:**
- Modify: `docs/architecture/adr/ADR-fe2bb6f9-deciso-crtica-com-caracteres-perigosos-aspas.md`

- [ ] **Step 1: Write minimal implementation**

Replace placeholders with technical reasoning on how the `instruct` command and ADR generator handle shell-sensitive characters through escaping and sanitization.

- [ ] **Step 2: Commit**

```bash
git add docs/architecture/adr/ADR-fe2bb6f9-deciso-crtica-com-caracteres-perigosos-aspas.md
git commit -m "docs(adr): address special character handling in automation"
```

## Task 5: Final Validation

- [ ] **Step 1: Verify all files**

Run a check to ensure no `[Descreva...]` or `[Ponto...]` strings remain in the `docs/architecture/adr/` directory.

Run: `grep -R --line-number '\[Descreva\.\.\.\]|\[Ponto\.\.\.\]' docs/architecture/adr/`
Expected: No matches (0 results). Any match means placeholders remain.
