package commands

import (
	"fmt"
	"os"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "sentinel",
	Short: "Sentinel Core: Governance & Context Engine for AI-Native Development",
}

func NewRootCmd(db *sqlite.DB) *cobra.Command {
	// Agrega todos os subcomandos injetando a dependência do DB
	RootCmd.AddCommand(NewPlanCmd(db))
	RootCmd.AddCommand(NewStartCmd(db))
	RootCmd.AddCommand(NewAuditCmd(db))
	RootCmd.AddCommand(NewScanCmd(db))
	RootCmd.AddCommand(NewVisualizeCmd(db))
	RootCmd.AddCommand(NewReportCmd(db))
	RootCmd.AddCommand(NewInstructCmd(db))
	RootCmd.AddCommand(NewStatusCmd(db))

	return RootCmd
}

func Execute(db *sqlite.DB) {
	if err := NewRootCmd(db).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
