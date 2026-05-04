package commands

import (
	"fmt"
	"os"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "sentinel",
	Short: "Sentinel Core: Governance & Context Engine for AI-Native Development",
}

func NewRootCmd(db *sqlite.DB) *cobra.Command {
	// Agrega todos os subcomandos registrados dinamicamente
	for _, factory := range registry.GetCommands() {
		RootCmd.AddCommand(factory(db))
	}

	return RootCmd
}

func Execute(db *sqlite.DB) {
	if err := NewRootCmd(db).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
