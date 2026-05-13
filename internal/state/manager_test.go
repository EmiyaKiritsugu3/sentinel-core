package state

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestNewManager_NilDB(t *testing.T) {
	t.Parallel()
	m, err := NewManager(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if m != nil {
		t.Error("expected nil Manager for nil db")
	}
}
