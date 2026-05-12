package commands

import (
	"fmt"
	"os"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func NewRootCmd(db *sqlite.DB) *cobra.Command {
	root := &cobra.Command{
		Use:   "sentinel",
		Short: "Sentinel Core: Governance & Context Engine for AI-Native Development",
	}

	if err := sqlite.ValidateDB(db, "root-cmd"); err != nil {
		root.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return root
	}

	// Agrega todos os subcomandos registrados dinamicamente
	for _, factory := range registry.GetCommands() {
		root.AddCommand(factory(db))
	}

	return root
}

func Execute(db *sqlite.DB) {
	if err := NewRootCmd(db).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
