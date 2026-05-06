// internal/intake/disambiguator_test.go
package intake_test

import (
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/intake"
)

func TestVaguenessScore_HighForShortGeneric(t *testing.T) {
	d := intake.NewDisambiguator(nil) // nil db = skip graph phase
	score := d.VaguenessScore("fix bug")
	if score <= 0.50 {
		t.Errorf("want score > 0.50 for 'fix bug', got %.2f", score)
	}
}

func TestVaguenessScore_LowForPrecise(t *testing.T) {
	d := intake.NewDisambiguator(nil)
	score := d.VaguenessScore("fix JWT validation in internal/agents/auth_provider.go")
	if score > 0.50 {
		t.Errorf("want score <= 0.50 for precise description, got %.2f", score)
	}
}

func TestVaguenessScore_LowForLongDescriptive(t *testing.T) {
	d := intake.NewDisambiguator(nil)
	score := d.VaguenessScore("refactor loadSurgicalContext to use graph-aware ranking based on edge count")
	if score > 0.50 {
		t.Errorf("want score <= 0.50 for long descriptive, got %.2f", score)
	}
}

func TestAnalyze_NotVague_NoSuggestions(t *testing.T) {
	d := intake.NewDisambiguator(nil)
	vague, suggestions := d.Analyze("fix JWT validation in internal/agents/auth_provider.go")
	if vague {
		t.Error("want not vague for precise description")
	}
	if len(suggestions) != 0 {
		t.Error("want no suggestions for non-vague description")
	}
}

func TestAnalyze_Vague_NilDB_ReturnsSuggestions(t *testing.T) {
	d := intake.NewDisambiguator(nil)
	vague, _ := d.Analyze("fix bug")
	if !vague {
		t.Error("want vague=true for 'fix bug'")
	}
}
