# Sovereign Math Engine (SME) — Phase 1: Foundations & Metrics

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the foundation of the Sovereign Math Engine, focusing on real-time metric collection (latency, tokens) and the implementation of the Net Gain Equation ($\Delta$).

**Architecture:** We will extend the SQLite schema to persist high-precision execution data and create an `internal/math` package to house the SME formulas. The `Engine` will be updated to emit these metrics during the ReAct loop.

**Tech Stack:** Go 1.26, SQLite, standard `math` and `time` packages.

---

## File Map

**Create:**
- `internal/math/formulas.go` — SME constants and the Delta ($\Delta$) function.
- `internal/math/formulas_test.go` — Validation of mathematical correctness.

**Modify:**
- `pkg/sqlite/db.go` — Update schema to include execution metrics.
- `internal/agents/types.go` — Add metric fields to `AgentContext` or a new `ExecutionResult` struct.
- `internal/agents/engine.go` — Capture and persist metrics during `Execute`.
- `internal/report/aggregator.go` — Calculate and display the SME score in the report.

---

### Task 1: Mathematical Foundations & Delta Formula

**Files:**
- Create: `internal/math/formulas.go`
- Create: `internal/math/formulas_test.go`

- [ ] **Step 1.1: Write failing tests for Delta equation**

```go
package math_test

import (
	"testing"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/math"
)

func TestCalculateDelta(t *testing.T) {
	// Case 1: High gain, low cost
	// Prob=0.9, Impact=10, Latency=50ms, Cost=0.01
	got := math.CalculateDelta(0.9, 10, 50, 0.01)
	if got <= 0 {
		t.Errorf("expected positive delta for high gain scenario, got %f", got)
	}

	// Case 2: Low gain, high cost (Placebo Processing)
	// Prob=0.1, Impact=1, Latency=500ms, Cost=0.50
	got = math.CalculateDelta(0.1, 1, 500, 0.50)
	if got >= 0 {
		t.Errorf("expected negative delta for placebo processing, got %f", got)
	}
}
```

- [ ] **Step 1.2: Run test — verify failure**

Run: `go test ./internal/math/...`
Expected: FAIL (package/function not found)

- [ ] **Step 1.3: Implement internal/math/formulas.go**

```go
package math

import "math"

// SME Constants for weighting
const (
	WeightLatency = 0.001 // Penalty per ms
	WeightToken   = 2.0   // Penalty per dollar (normalized)
)

// CalculateDelta implements the Net Gain Equation:
// Δ = (Ph * Wb) - (Lo + Ca)
func CalculateDelta(probHallucination, bugWeight, latencyMs, apiCost float64) float64 {
	gain := probHallucination * bugWeight
	loss := (latencyMs * WeightLatency) + (apiCost * WeightToken)
	return gain - loss
}
```

- [ ] **Step 1.4: Run test — verify PASS**

Run: `go test ./internal/math/... -v`
Expected: PASS

- [ ] **Step 1.5: Commit**

```bash
git add internal/math/
git commit -m "feat(math): implement foundational Net Gain Equation (Delta)"
```

---

### Task 2: Persistence Layer Hardening

**Files:**
- Modify: `pkg/sqlite/db.go`

- [ ] **Step 2.1: Update Schema**

Find `Init` or the table creation logic. Add the following columns to the `tasks` or a new `executions` table:
- `latency_ms` (REAL)
- `tokens_used` (INTEGER)
- `api_cost` (REAL)
- `math_delta` (REAL)

```go
// Inside Init() in pkg/sqlite/db.go, add to tasks table if not present:
// ALTER TABLE tasks ADD COLUMN latency_ms REAL DEFAULT 0;
// ... (repeat for other columns)
```

- [ ] **Step 2.2: Verify build**

Run: `go build ./pkg/sqlite/...`
Expected: Success

- [ ] **Step 2.3: Commit**

```bash
git add pkg/sqlite/db.go
git commit -m "feat(db): update schema for high-precision math metrics"
```

---

### Task 3: Engine Integration & Real-Time Collection

**Files:**
- Modify: `internal/agents/engine.go`
- Modify: `internal/agents/types.go`

- [ ] **Step 3.1: Add metrics to AgentContext**

Update `AgentContext` struct in `internal/agents/types.go`:

```go
type AgentContext struct {
    // ... existing fields
    StartTime     time.Time
    EndTime       time.Time
    TokensUsed    int
    APICost       float64
}
```

- [ ] **Step 3.2: Instrument Engine.Execute**

In `internal/agents/engine.go`, capture start and end times. At the end of `Execute`, calculate Delta and save to DB.

```go
// Near start of Execute:
ctx.StartTime = time.Now()

// Near end of Execute (before returning):
ctx.EndTime = time.Now()
latency := float64(ctx.EndTime.Sub(ctx.StartTime).Milliseconds())

// Use a placeholder for now for prob and weight (to be refined in Phase 7.3)
delta := math.CalculateDelta(0.5, 5, latency, ctx.APICost)

// Save to DB:
_, err := e.DB.Conn.Exec("UPDATE tasks SET latency_ms = ?, math_delta = ? WHERE id = ?", latency, delta, ctx.StateID)
```

- [ ] **Step 3.3: Verify with existing tests**

Run: `go test ./internal/agents/...`
Expected: PASS (ensure no regression in the ReAct loop)

- [ ] **Step 3.4: Commit**

```bash
git add internal/agents/
git commit -m "feat(engine): instrument execution with real-time math metrics"
```

---

### Task 4: Mathematical Reporting

**Files:**
- Modify: `internal/report/aggregator.go`

- [ ] **Step 4.1: Update Report Output**

Modify the aggregator to pull the new metrics and display the "Sovereign Efficiency" score.

```go
// Inside aggregator logic:
// Fetch AVG(math_delta) from tasks table.
// Display: "Sovereign Efficiency (Δ): +X.XX"
```

- [ ] **Step 4.2: Commit**

```bash
git add internal/report/
git commit -m "feat(report): include Mathematical Efficiency in the sprint summary"
```
