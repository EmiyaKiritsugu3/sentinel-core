# Sentinel Compliance Dashboard 📊 [PID-SENTINEL]

> [!NOTE]
> Este relatório é gerado automaticamente pelo Guardião.

## 🏁 Key Performance Indicators (KPIs)

| Métrica | Valor |
| :--- | :--- |
| **Test Coverage (SonarCloud)** | 78.8% (QG threshold: 80%) |
| **Test Coverage (go test)** | 32.3% (statements) |
| **Go Source Files** | 136 |
| **Internal Packages** | 17 (13 with tests) |
| **SME Math Coverage** | 100.0% (`internal/math`) |
| **Intake Coverage** | 90.4% (`internal/intake`) |
| **Report Coverage** | 82.5% (`internal/report`) |
| **Agents Coverage** | 59.2% (`internal/agents`) |
| **SQLite Coverage** | 68.2% (`pkg/sqlite`) |
| **SonarCloud QG Status** | ⚠️ FAILED (78.8% < 80% threshold) |
| **CI/CD Checks** | CodeQL ✅ | CodeRabbit ✅ | golangci-lint ✅ |

## 🛡️ Mathematical Sovereignty Status

| Pillar | Status | Implementation |
| :--- | :--- | :--- |
| **Pillar A: Net Gain Equation (Δ)** | ✅ COMPLETED | `internal/math/formulas.go` |
| **Pillar B: Lyapunov Stability (λ)** | ✅ COMPLETED | Gate A (`engine.go`) + Gate A.5 (`engine_helpers.go`) |
| **Pillar C: Topological Analysis** | ⏳ DEFERRED | Not in current sprint |
| **Pillar D: Bayesian Trust** | ✅ COMPLETED | `agent_trust` table + `CalculateTrustScore` + `TrustToDynamicLambda` |

## 🛡️ Task Lifecycle Status

- ✅ **Completed**: 0
- 🛑 **Failed**: 0
- 🕒 **Total Attempts**: 4

## 📝 Detailed Intent Inventory

| ID | Tier | Status | Description | Decision Record |
| :--- | :--- | :--- | :--- | :--- |
| `fe2bb6f9` | T1 | PENDING | Decisão Crítica --- com : caracteres @ perigosos "aspas" | [View ADR](../architecture/adr/ADR-fe2bb6f9-deciso-crtica-com-caracteres-perigosos-aspas.md) |
| `ad9933bf` | T1 | PENDING | Refatorar camada de persistencia | [View ADR](../architecture/adr/ADR-ad9933bf-refatorar-camada-de-persistencia.md) |
| `aca540c1` | T1 | PENDING | Implementar Auto-ADR Engine | N/A |
| `de75082b` | T1 | PENDING | Auditoria de Seguranca de vanguarda | N/A |
| `66d2618f` | T1 | PROPOSED | melhorar performance | [View ADR](../architecture/adr/ADR-66d2618f-melhorar-performance.md) |
