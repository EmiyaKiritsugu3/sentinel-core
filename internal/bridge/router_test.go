// internal/bridge/router_test.go
package bridge_test

import (
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
)

func TestStrategyFor_KnownIntents_HaveNodeLimit(t *testing.T) {
	t.Parallel()
	for _, intent := range []bridge.Intent{
		bridge.IntentDiagnose,
		bridge.IntentImplement,
		bridge.IntentRefactor,
		bridge.IntentReview,
	} {
		s := bridge.StrategyFor(intent)
		if s.NodeLimit == 0 {
			t.Errorf("intent %s: NodeLimit must be > 0", intent)
		}
	}
}

func TestStrategyFor_Unknown_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	s := bridge.StrategyFor(bridge.IntentUnknown)
	if s.NodeLimit != 0 || s.HighCoupling || s.IncludeTests || s.IncludeADRs {
		t.Error("IntentUnknown must return zero-value ContextStrategy")
	}
}

func TestStrategyFor_Diagnose_HasHighCoupling(t *testing.T) {
	t.Parallel()
	s := bridge.StrategyFor(bridge.IntentDiagnose)
	if !s.HighCoupling {
		t.Error("diagnose strategy must include high coupling nodes")
	}
}

func TestStrategyFor_Implement_HasTests(t *testing.T) {
	t.Parallel()
	s := bridge.StrategyFor(bridge.IntentImplement)
	if !s.IncludeTests {
		t.Error("implement strategy must include test files")
	}
}
