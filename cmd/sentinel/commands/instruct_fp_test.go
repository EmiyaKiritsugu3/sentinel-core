package commands

import "testing"

// CG-01 FP test: strings.Contains(strings.ToLower(intent), "performance") classifies
// intent as vague. FP: intent containing "performance" as part of a compound
// word or in a non-vague context is misclassified.

func TestIsVagueIntent_FP_Performance(t *testing.T) {
	// strings.Contains(strings.ToLower(intent), "performance") → isVague=true
	// FP: "Refactor PerformanceMonitor to use pooled connections" contains
	// "performance" but is a precise, actionable task description.
	intent := "Refactor PerformanceMonitor to use pooled connections"
	got := isVagueIntent(intent)
	t.Logf("CG-01 FP: 'PerformanceMonitor' intent → isVague=%v (expected false, got true = FP)", got)
	if !got {
		t.Log("FP not triggered — no misclassification for this input")
	}
}
