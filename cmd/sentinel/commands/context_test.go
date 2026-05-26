package commands

import (
	"testing"
)

func TestNewContextCmd_Registers(t *testing.T) {
	cmd := NewContextCmd(nil)
	if cmd.Name() != "context" {
		t.Fatalf("expected Name()='context', got %q", cmd.Name())
	}
}

func TestContextCmdFlags(t *testing.T) {
	cmd := NewContextCmd(nil)
	flags := []string{"limit", "budget", "dry-run"}
	for _, f := range flags {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("missing flag: %s", f)
		}
	}
}
