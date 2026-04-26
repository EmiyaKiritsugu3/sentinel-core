package commands

import (
	"fmt"
	"log"
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Inicializa o banco de dados antes de qualquer comando
		db, err := sqlite.Init()
		if err != nil {
			log.Fatalf("❌ Failed to initialize sentinel brain: %v", err)
		}
		DBInstance = db

		// Garante migrações
		if err := graph.Migrate(db); err != nil {
			log.Fatalf("❌ Failed to migrate database: %v", err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if DBInstance != nil {
			DBInstance.Close()
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
