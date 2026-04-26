package commands

import (
	"fmt"
	"os"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

var (
	DBInstance *sqlite.DB
	Version    = "5.0.0-alpha"
)

var RootCmd = &cobra.Command{
	Use:   "sentinel",
	Short: "Sentinel Core: Governance & Context Engine for AI-Native Development",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Inicializa o banco de dados antes de qualquer comando
		db, err := sqlite.Init()
		if err != nil {
			return fmt.Errorf("root: failed to initialize sentinel brain: %w", err)
		}
		DBInstance = db

		// Garante migrações
		if err := graph.Migrate(db); err != nil {
			return fmt.Errorf("root: failed to migrate database: %w", err)
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if DBInstance != nil {
			DBInstance.Close()
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// Se o DBInstance ainda estiver aberto por um erro precoce, fechamos best-effort
		if DBInstance != nil {
			_ = DBInstance.Close()
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
