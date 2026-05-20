package bridge_test

import (
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
)

func TestNewSDKClient_NilClient(t *testing.T) {
	c, err := bridge.NewSDKClient(nil)
	if !errors.Is(err, bridge.ErrNilClient) {
		t.Fatalf("expected ErrNilClient, got: %v", err)
	}
	if c != nil {
		t.Fatalf("expected nil client wrapper, got: %v", c)
	}
}
