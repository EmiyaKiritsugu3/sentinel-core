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

func TestCalculateLambda(t *testing.T) {
	tests := []struct {
		name       string
		action     int
		thought    int
		expected   float64
	}{
		{"normal", 100, 50, 2.0},
		{"lazy_thought", 500, 10, 50.0},
		{"zero_thought", 100, 0, 100.0}, // Fallback behavior
		{"zero_action", 0, 50, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := math.CalculateLambda(tt.action, tt.thought)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
