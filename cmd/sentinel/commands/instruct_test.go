package commands

import "testing"

func TestIsVagueIntent(t *testing.T) {
	tests := []struct {
		name     string
		intent   string
		expected bool
	}{
		{"Short intent (1 word)", "short", true},
		{"Short intent (2 words)", "make fast", true},
		{"Non-vague intent (3 words)", "very long string", false},
		{"Non-vague intent (4 words)", "very very long string", false},
		{"Vague intent containing performance (lowercase)", "improve performance", true},
		{"Vague intent containing performance (mixed case)", "Make It PeRfoRmAnCe", true},
		{"Vague intent containing performance (uppercase)", "IMPROVE PERFORMANCE", true},
		{"Non-vague intent without performance", "add a new feature", false},
		{"Empty intent", "", true},
		{"String with just spaces", "   ", false}, // "   " has 3 spaces, strings.Count("   ", " ") == 3, so not vague unless it contains "performance"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVagueIntent(tt.intent)
			if result != tt.expected {
				t.Errorf("isVagueIntent(%q) = %v, expected %v", tt.intent, result, tt.expected)
			}
		})
	}
}
