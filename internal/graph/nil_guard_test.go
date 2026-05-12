package graph

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestNewEngine_NilDB(t *testing.T) {
	e, err := NewEngine(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if e != nil {
		t.Error("expected nil Engine for nil db")
	}
}

func TestNewVisualizer_NilDB(t *testing.T) {
	v, err := NewVisualizer(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if v != nil {
		t.Error("expected nil Visualizer for nil db")
	}
}
