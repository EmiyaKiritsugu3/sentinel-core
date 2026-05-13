package math_test

import (
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/math"
)

func TestCalculateDelta(t *testing.T) {
	t.Parallel()
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

func TestCalculateTrustScore(t *testing.T) {
	t.Parallel()
	if got := math.CalculateTrustScore(0, 0); got != 0.5 {
		t.Errorf("zero history: want 0.5, got %f", got)
	}
	if got := math.CalculateTrustScore(100, 100); got <= 0.9 {
		t.Errorf("perfect record: want > 0.9, got %f", got)
	}
	if got := math.CalculateTrustScore(0, 100); got >= 0.05 {
		t.Errorf("zero successes: want < 0.05, got %f", got)
	}
}

func TestTrustToDynamicLambda(t *testing.T) {
	t.Parallel()
	if got := math.TrustToDynamicLambda(0.0); got != 0.5 {
		t.Errorf("trust=0: want 0.5, got %f", got)
	}
	if got := math.TrustToDynamicLambda(1.0); got != 1.5 {
		t.Errorf("trust=1: want 1.5, got %f", got)
	}
}

func TestCalculateDivergence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		current  float64
		previous float64
		wantHigh bool // true = expect divergence > 1.0
	}{
		{"stable", 2.1, 2.0, false},         // 5% change = stable
		{"high_divergence", 4.0, 1.0, true}, // 300% change = diverging
		{"zero_previous", 1.0, 0.0, true},   // uses 1e-9 denominator
		{"both_zero", 0.0, 0.0, false},      // 0/1e-9 = 0
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := math.CalculateDivergence(tt.current, tt.previous)
			if tt.wantHigh && got <= 1.0 {
				t.Errorf("%s: expected divergence > 1.0, got %.4f", tt.name, got)
			}
			if !tt.wantHigh && got > 1.0 {
				t.Errorf("%s: expected divergence <= 1.0, got %.4f", tt.name, got)
			}
		})
	}
}

func TestCalculateLambda(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		action   int
		thought  int
		expected float64
	}{
		{"normal", 100, 50, 2.0},
		{"lazy_thought", 500, 10, 50.0},
		{"zero_thought", 100, 0, 100.0}, // Fallback behavior
		{"zero_action", 0, 50, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := math.CalculateLambda(tt.action, tt.thought)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
