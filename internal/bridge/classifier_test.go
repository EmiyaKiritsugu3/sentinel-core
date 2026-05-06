package bridge_test

import (
	"context"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
)

func TestHeuristic_Diagnose(t *testing.T) {
	c := bridge.NewIntentClassifier(nil, 0.60)
	got := c.Classify(context.Background(), "t1", "fix the broken JWT validation")
	if got != bridge.IntentDiagnose {
		t.Errorf("want diagnose, got %s", got)
	}
}

func TestHeuristic_Implement(t *testing.T) {
	c := bridge.NewIntentClassifier(nil, 0.60)
	got := c.Classify(context.Background(), "t2", "add OAuth2 support to auth module")
	if got != bridge.IntentImplement {
		t.Errorf("want implement, got %s", got)
	}
}

func TestHeuristic_Ambiguous_ReturnsUnknown(t *testing.T) {
	// "fix" (diagnose) + "review" (review) = 2 categories = confidence 0.30 < 0.60
	// AI is nil → IntentUnknown
	c := bridge.NewIntentClassifier(nil, 0.60)
	got := c.Classify(context.Background(), "t3", "fix and review the auth module")
	if got != bridge.IntentUnknown {
		t.Errorf("want unknown for ambiguous+nil AI, got %s", got)
	}
}

func TestHeuristic_NoMatch_ReturnsUnknown(t *testing.T) {
	c := bridge.NewIntentClassifier(nil, 0.60)
	got := c.Classify(context.Background(), "t4", "the JWT module")
	if got != bridge.IntentUnknown {
		t.Errorf("want unknown for no match, got %s", got)
	}
}

func TestCache_ReturnsCachedOnSecondCall(t *testing.T) {
	c := bridge.NewIntentClassifier(nil, 0.60)
	first := c.Classify(context.Background(), "t5", "fix the bug")
	second := c.Classify(context.Background(), "t5", "completely different description")
	if first != second {
		t.Errorf("want cache hit (same intent), got %s vs %s", first, second)
	}
}

func TestNilClassifier_ReturnsUnknown(t *testing.T) {
	n := bridge.NewNilClassifier()
	got, err := n.Classify(context.Background(), "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bridge.IntentUnknown {
		t.Errorf("want unknown, got %s", got)
	}
}
