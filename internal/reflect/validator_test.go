package reflect

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestNewValidator_NilDB(t *testing.T) {
	v, err := NewValidator(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if v != nil {
		t.Error("expected nil Validator for nil db")
	}
}
