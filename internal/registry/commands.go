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

// Register adiciona uma factory ao registry global de forma thread-safe.
func Register(factory CommandFactory) {
	mu.Lock()
	defer mu.Unlock()
	factories = append(factories, factory)
}

// GetCommands retorna uma cópia defensiva das factories registradas.
func GetCommands() []CommandFactory {
	mu.Lock()
	defer mu.Unlock()
	result := make([]CommandFactory, len(factories))
	copy(result, factories)
	return result
}

// ResetForTesting limpa o registry global. Apenas para uso em testes.
func ResetForTesting() {
	mu.Lock()
	defer mu.Unlock()
	factories = nil
}
