package agents

import (
	"fmt"
	"os"
)

// AuthProvider defines the interface for retrieving API keys.
type AuthProvider interface {
	GetAPIKey() (string, error)
}

// SovereignAuthProvider is the default implementation of AuthProvider for Sentinel.
type SovereignAuthProvider struct{}

// GetAPIKey retrieves the GOOGLE_API_KEY from the environment.
// It follows Standard #05 for error wrapping.
func (p *SovereignAuthProvider) GetAPIKey() (string, error) {
	key := os.Getenv("GOOGLE_API_KEY")
	if key != "" {
		return key, nil
	}
	return "", fmt.Errorf("no GOOGLE_API_KEY found in environment")
}
