# internal/math

Sovereign Math Engine for divergence detection, trust calibration, and cognitive entropy analysis.

## Overview

The math package implements the formulas that power Sentinel's governance gates. All functions are pure, side-effect-free computations that operate on agent metrics collected during execution.

## Key Functions

### `CalculateDelta(probHallucination, bugWeight, latencyMs, apiCost float64) float64`
The Net Gain Equation: `Δ = (Ph * Wb) - (Lo + Ca)`. Penalizes high latency (`WeightLatency = 0.001` per ms) and API cost (`WeightToken = 2.0` per dollar). Used to evaluate whether an agent's work was net-positive.

### `CalculateTrustScore(successes, total int) float64`
Bayesian trust calibration with Laplace smoothing (`alpha = 1.0, beta = 1.0`). Returns a value in (0, 1) representing the agent's reliability based on task outcomes.

### `TrustToDynamicLambda(trustScore float64) float64`
Maps trust score to a dynamic lambda multiplier. Range: 0.5 (untrusted) to 1.5 (fully trusted). Low-trust agents face stricter entropy gates.

### `CalculateLambda(actionTokens, thoughtTokens int) float64`
Cognitive averaging metric: `actionTokens / thoughtTokens`. Used by Gate A to detect over-reasoning. Guards against division by zero.

### `CalculateDivergence(lambdaCurrent, lambdaPrevious float64) float64`
Lyapunov proxy for reasoning stability. Computes the relative rate of change between consecutive lambda values: `|current - previous| / previous`. A divergence count ≥ 2 consecutive steps triggers a structural pivot (Gate A.5). Handles zero-denominator by clamping to `1e-9`.

## Usage

```go
import "github.com/EmiyaKiritsugu3/sentinel-core/internal/math"

lambda := math.CalculateLambda(actionTokens, thoughtTokens)
trust := math.CalculateTrustScore(successes, total)
dynamicLimit := *agent.MaxLambda * math.TrustToDynamicLambda(trust)

if lambda > dynamicLimit {
    // Intervene: agent is over-reasoning
}
```
