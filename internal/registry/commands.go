package registry

import (
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

// CommandFactory is a function that creates a cobra.Command with a DB dependency.
type CommandFactory func(*sqlite.DB) *cobra.Command

var factories []CommandFactory

// Register adds a new command factory to the global registry.
func Register(factory CommandFactory) {
	factories = append(factories, factory)
}

// GetCommands returns all registered command factories.
func GetCommands() []CommandFactory {
	return factories
}
