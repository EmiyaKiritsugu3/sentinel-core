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
