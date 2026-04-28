package agents

import (
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
)

type mockAuthProvider struct {
	key string
}

func (m *mockAuthProvider) GetAPIKey() (string, error) {
	return m.key, nil
}

func TestNewEngine(t *testing.T) {
	registry := NewRegistry()
	auth := &mockAuthProvider{key: "fake-key"}
	factory := bridge.NewFactory(nil)
	validator := reflect.NewValidator(nil)

	engine, err := NewEngine(registry, auth, factory, validator)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if engine.genaiClient == nil {
		t.Fatal("genaiClient should not be nil")
	}
}
