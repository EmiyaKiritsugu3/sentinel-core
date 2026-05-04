# Implementation Plan - Root Hub Decoupling [PID-SENTINEL-HUB-DECOUPLING-V2]

## 🎯 Goal

Decouple the `NewRootCmd` (CLI Dispatcher) from explicit sub-command instantiation. This implements the **"Open-Closed Principle"** for Sentinel commands, allowing for the addition of new features (like future Linkers or Graph analyzers) without modifying the root hierarchy.

- **FPA Estimate**: 5 (Architectural refactor, low risk, high maintainability gain)
- **Status**: PROPOSED
- **Council Ratification**: Required (Auditor role activated)

## 📋 Steps

### Phase 1: Command Registry Infrastructure

- [ ] **Task 1.1**: Create `internal/commands/registry.go`.
  - Define `CommandFactory` type: `func(*sqlite.DB) *cobra.Command`.
  - Implement a global slice and a `Register(CommandFactory)` function.
  - Implement `GetCommands(db *sqlite.DB)` to return the instantiated commands.

### Phase 2: Command Discovery Migration

- [ ] **Task 2.1**: Refactor each command in `cmd/sentinel/commands/`.
  - Files: `audit.go`, `plan.go`, `scan.go`, `start.go`, `status.go`, `visualize.go`, `report.go`, `instruct.go`.
  - Add an `init()` block to each that calls `registry.Register(New...Cmd)`.
- [ ] **Task 2.2**: Decouple `cmd/sentinel/commands/root.go`.
  - Remove all explicit `New...Cmd` calls.
  - Replace with a loop: `for _, cmd := range registry.GetCommands(db) { RootCmd.AddCommand(cmd) }`.

### Phase 3: Sovereign Verification

- [ ] **Task 3.1**: Run `sentinel status` to verify DB connectivity through the new registry.
- [ ] **Task 3.2**: Run `sentinel scan` to verify AST functionality through the new registry.
- [ ] **Task 3.3**: Execute `./scripts/audit-local.sh` (Pre-Flight 2.1).

## 🛡️ Verification (Hard Gates)

- `sentinel --help` displays all commands correctly.
- No circular dependencies introduced during registration.
- Pre-Flight 2.1 returns `ALL GATES OPEN`.

## ⚠️ Gotchas

- **Registration Order**: Sub-commands will now be registered in the order they are initialized by the Go runtime. This shouldn't affect functionality but might change the order in `--help`.
- **Package Imports**: To trigger `init()`, the packages must be imported (even as blank imports) in `root.go`. We will use a central `loader.go` to keep `root.go` clean.
