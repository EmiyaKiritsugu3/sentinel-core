package agents

import (
	"os"
	"testing"
)

func TestSovereignAuthProvider_GetAPIKey(t *testing.T) {
	t.Parallel()
	t.Run("returns key when GOOGLE_API_KEY is set", func(t *testing.T) {
		t.Parallel()
		expectedKey := "test-api-key"
		_ = os.Setenv("GOOGLE_API_KEY", expectedKey)
		defer func() { _ = os.Unsetenv("GOOGLE_API_KEY") }()

		provider := &SovereignAuthProvider{}
		key, err := provider.GetAPIKey()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if key != expectedKey {
			t.Errorf("expected key %s, got %s", expectedKey, key)
		}
	})

	t.Run("returns error when GOOGLE_API_KEY is not set", func(t *testing.T) {
		t.Parallel()
		_ = os.Unsetenv("GOOGLE_API_KEY")

		provider := &SovereignAuthProvider{}
		_, err := provider.GetAPIKey()

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		expectedErr := "no GOOGLE_API_KEY found in environment"
		if err.Error() != expectedErr {
			t.Errorf("expected error %q, got %q", expectedErr, err.Error())
		}
	})
}
