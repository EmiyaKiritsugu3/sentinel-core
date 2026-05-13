package audit

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestNewRunner_NilDB(t *testing.T) {
	t.Parallel()
	r, err := NewRunner(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if r != nil {
		t.Error("expected nil Runner for nil db")
	}
}
