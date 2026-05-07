package math

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

// CalculateLambda computes the cognitive averaging metric (Action Tokens / Thought Tokens).
func CalculateLambda(actionTokens, thoughtTokens int) float64 {
	if thoughtTokens == 0 {
		return float64(actionTokens) // Prevent division by zero
	}
	return float64(actionTokens) / float64(thoughtTokens)
}
