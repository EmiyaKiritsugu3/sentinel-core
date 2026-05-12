package bridge

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestNewGeminiClassifier_NilClient(t *testing.T) {
	gc, err := NewGeminiClassifier(nil)
	if err == nil {
		t.Fatal("expected error for nil client, got nil")
	}
	if gc != nil {
		t.Error("expected nil GeminiClassifier for nil client")
	}
}

func TestNewFactory_NilDB(t *testing.T) {
	f, err := NewFactory(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if f != nil {
		t.Error("expected nil Factory for nil db")
	}
}
