// Package registry manages command registration and lifecycle.
package registry

import (
	"sync"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

// CommandFactory is a function that creates a cobra.Command with a DB dependency.
type CommandFactory func(*sqlite.DB) *cobra.Command

var (
	factories []CommandFactory
	mu        sync.Mutex
)

// Register adds a factory to the global registry in a thread-safe way.
func Register(factory CommandFactory) {
	mu.Lock()
	defer mu.Unlock()
	factories = append(factories, factory)
}

// GetCommands returns a defensive copy of registered factories.
func GetCommands() []CommandFactory {
	mu.Lock()
	defer mu.Unlock()
	result := make([]CommandFactory, len(factories))
	copy(result, factories)
	return result
}

// ResetForTesting clears the global registry. Only for test use.
func ResetForTesting() {
	mu.Lock()
	defer mu.Unlock()
	factories = nil
}
