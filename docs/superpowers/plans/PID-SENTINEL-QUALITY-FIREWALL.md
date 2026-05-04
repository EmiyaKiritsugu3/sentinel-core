# Implementation Plan - Sovereign Quality Firewall [PID-SENTINEL-QUALITY-FIREWALL]

## 🎯 Goal

Implement the Foundational Quality Layer using native Go tools and Markdownlint to ensure codebase health and documentation standards, adhering to the "Native-First" pivot mandated in `sentinel-log.md`.

- **FPA Estimate**: 12 (Critical infrastructure change + multi-file impact)
- **Status**: COMPLETED 🛡️
- **Council Ratification**: Required (Warden/Auditor roles activated)

## 📋 Steps

### Phase 1: Step 1.1 - Native Go Quality Gate

- [x] **Task 1.1.1**: Update `docs/architecture/adr/ADR-d0555ca9-...`
  - Status: `ACCEPTED`
  - Decision: Use `go fmt ./... && go vet ./...` instead of `golangci-lint`.
  - Verification Command: `bash -c "if [ -z \"\$(gofmt -l .)\" ]; then go vet ./...; else gofmt -l .; exit 1; fi"`
- [x] **Task 1.1.2**: Execute formatting and linting.
  - Run `go fmt ./...`
  - Run `go vet ./...`
- [x] **Task 1.1.3**: Audit via Sentinel.
  - `sentinel start d0555ca9-2139-4b67-8261-5a64afd44e24`
  - `sentinel audit`

### Phase 2: Step 1.2 - Markdown Standards

- [x] **Task 1.2.1**: Update `docs/architecture/adr/ADR-306b0ea4-...`
  - Status: `ACCEPTED`
  - Decision: Use `markdownlint-cli2` for documentation consistency.
  - Verification Command: `npx --yes markdownlint-cli2`
- [x] **Task 1.2.2**: Resolve markdown violations.
  - Audit all `.md` files.
  - Fix common issues (spacing, headings, line length).
- [x] **Task 1.2.3**: Audit via Sentinel.
  - `sentinel start 306b0ea4-2d45-445c-b098-8efca7a98745`
  - `sentinel audit`

### Phase 3: Consolidation

- [x] **Task 1.3.1**: Verify `scripts/audit-local.sh` alignment.
- [x] **Task 1.3.2**: Update `sentinel-log.md` with session epiphany.

## 🛡️ Verification (Hard Gates)

- [x] `go vet ./...` returns 0.
- [x] `gofmt -l .` returns empty string.
- [x] `npx markdownlint-cli2` returns 0.
- [x] `sentinel audit` passes for both tasks.

## ⚠️ Gotchas

- `go vet` might complain about shadow variables or printf formats in older code.
- `markdownlint` can be noisy on legacy docs; prioritizing `.markdownlint-cli2.yaml` configuration.
- `sentinel audit` requires an active task in the SQLite DB.
