package agents

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func TestNewMutationEngine_NilDB(t *testing.T) {
	e, err := NewMutationEngine(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if e != nil {
		t.Error("expected nil MutationEngine for nil db")
	}
}

func TestNewRegistryManager_NilDB(t *testing.T) {
	m, err := NewRegistryManager(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if m != nil {
		t.Error("expected nil RegistryManager for nil db")
	}
}

func TestNewDispatcher_NilRegistry(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	shield := &GitShield{}
	d, err := NewDispatcher(nil, shield, db)
	if err == nil {
		t.Fatal("expected error for nil registry, got nil")
	}
	if d != nil {
		t.Error("expected nil Dispatcher for nil registry")
	}
}

func TestNewDispatcher_NilShield(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	regMgr, err := NewRegistryManager(db)
	if err != nil {
		t.Fatalf("NewRegistryManager() error: %v", err)
	}

	d, err := NewDispatcher(regMgr, nil, db)
	if err == nil {
		t.Fatal("expected error for nil shield, got nil")
	}
	if d != nil {
		t.Error("expected nil Dispatcher for nil shield")
	}
}

func TestNewDispatcher_NilDB(t *testing.T) {
	regMgr := &RegistryManager{}
	shield := &GitShield{}

	d, err := NewDispatcher(regMgr, shield, nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Errorf("expected ErrNilDB, got: %v", err)
	}
	if d != nil {
		t.Error("expected nil Dispatcher for nil db")
	}
}

func TestNewEngine_NilRegistry(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	auth := &mockAuthProvider{key: "fake-key"}
	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("reflect.NewValidator() error: %v", err)
	}

	e, err := NewEngine(nil, auth, validator, db)
	if err == nil {
		t.Fatal("expected error for nil registry, got nil")
	}
	if e != nil {
		t.Error("expected nil Engine for nil registry")
	}
}

func TestNewEngine_NilAuth(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	validator, err := reflect.NewValidator(db)
	if err != nil {
		t.Fatalf("reflect.NewValidator() error: %v", err)
	}

	e, err := NewEngine(NewRegistry(), nil, validator, db)
	if err == nil {
		t.Fatal("expected error for nil auth, got nil")
	}
	if e != nil {
		t.Error("expected nil Engine for nil auth")
	}
}

func TestNewEngine_NilValidator(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	auth := &mockAuthProvider{key: "fake-key"}

	e, err := NewEngine(NewRegistry(), auth, nil, db)
	if err == nil {
		t.Fatal("expected error for nil validator, got nil")
	}
	if e != nil {
		t.Error("expected nil Engine for nil validator")
	}
}

func TestNewEngine_NilDB(t *testing.T) {
	auth := &mockAuthProvider{key: "fake-key"}
	validator := &reflect.Validator{}

	e, err := NewEngine(NewRegistry(), auth, validator, nil)
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
