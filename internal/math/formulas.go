package math

import gomath "math"

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

// CalculateTrustScore computes the Bayesian-updated trust for an agent.
// successes = number of tasks completed without hallucination/error
// total = total tasks executed
// Returns a value in (0, 1).
func CalculateTrustScore(successes, total int) float64 {
	const alpha, beta = 1.0, 1.0 // Laplace prior
	return (float64(successes) + alpha) / (float64(total) + alpha + beta)
}

// TrustToDynamicLambda maps a TrustScore to a lambda multiplier.
// Low trust = stricter gate (multiplier < 1.0); high trust = looser gate.
// Range: 0.5 (trust=0) → 1.5 (trust=1)
func TrustToDynamicLambda(trustScore float64) float64 {
	return 0.5 + trustScore
}

// CalculateDivergence computes the relative rate of change between two lambda values.
// Returns 0 if both are zero. Used as a Lyapunov proxy for reasoning stability.
func CalculateDivergence(lambdaCurrent, lambdaPrevious float64) float64 {
	denominator := lambdaPrevious
	if denominator < 1e-9 {
		denominator = 1e-9
	}
	return gomath.Abs(lambdaCurrent-lambdaPrevious) / denominator
}

// CalculateLambda computes the cognitive averaging metric (Action Tokens / Thought Tokens).
func CalculateLambda(actionTokens, thoughtTokens int) float64 {
	if thoughtTokens == 0 {
		return float64(actionTokens) // Prevent division by zero
	}
	return float64(actionTokens) / float64(thoughtTokens)
}
