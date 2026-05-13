package math_test

import (
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/math"
)

// --- CalculateDelta benchmarks ---

func BenchmarkCalculateDelta_HighGain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateDelta(0.9, 10, 50, 0.01)
	}
}

func BenchmarkCalculateDelta_LowGain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateDelta(0.1, 1, 500, 0.50)
	}
}

// --- CalculateTrustScore benchmarks ---

func BenchmarkCalculateTrustScore_ZeroHistory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateTrustScore(0, 0)
	}
}

func BenchmarkCalculateTrustScore_PerfectRecord(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateTrustScore(100, 100)
	}
}

func BenchmarkCalculateTrustScore_LargeDataset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateTrustScore(10000, 12000)
	}
}

// --- TrustToDynamicLambda benchmarks ---

func BenchmarkTrustToDynamicLambda(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.TrustToDynamicLambda(0.75)
	}
}

// --- CalculateDivergence benchmarks ---

func BenchmarkCalculateDivergence_Stable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateDivergence(2.1, 2.0)
	}
}

func BenchmarkCalculateDivergence_HighDivergence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateDivergence(4.0, 1.0)
	}
}

func BenchmarkCalculateDivergence_ZeroPrevious(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateDivergence(1.0, 0.0)
	}
}

// --- CalculateLambda benchmarks ---

func BenchmarkCalculateLambda_Normal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateLambda(100, 50)
	}
}

func BenchmarkCalculateLambda_ZeroThought(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateLambda(100, 0)
	}
}

func BenchmarkCalculateLambda_LargeTokens(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.CalculateLambda(50000, 25000)
	}
}
