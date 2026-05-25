package commands

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestNewDebriefCmd_Registers(t *testing.T) {
	cmd := NewDebriefCmd(nil)
	if cmd.Use != "debrief" {
		t.Fatalf("expected Use='debrief', got %q", cmd.Use)
	}
}

func TestNewDebriefCmd_NilDB(t *testing.T) {
	cmd := NewDebriefCmd(nil)
	err := cmd.Execute()
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}

func TestDebriefFlags(t *testing.T) {
	cmd := NewDebriefCmd(nil)
	flags := []string{"auto", "editor", "dry-run", "output"}
	for _, f := range flags {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("missing flag: %s", f)
		}
	}
}
