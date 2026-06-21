package commands

import "testing"

func TestIsVagueIntent(t *testing.T) {
	tests := []struct {
		intent string
		want   bool
	}{
		{"fix bug", true},
		{"improve performance", true},
		{"Improve Performance", true},
		{"IMPROVE PERFORMANCE", true},
		{"Improve PeRfOrMaNcE", true},
		{"Refactor database connection pool", false},
	}

	for _, tt := range tests {
		t.Run(tt.intent, func(t *testing.T) {
			if got := isVagueIntent(tt.intent); got != tt.want {
				t.Errorf("isVagueIntent(%q) = %v, want %v", tt.intent, got, tt.want)
			}
		})
	}
}
