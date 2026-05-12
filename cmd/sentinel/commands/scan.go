package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewScanCmd)
}

func NewScanCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the project code to update the graph database",
	}

	if err := sqlite.ValidateDB(db, "scan-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
			var err error
			// Auto-Migração: garante que o banco esteja pronto
			err = graph.Migrate(db)
			if err != nil {
				return fmt.Errorf("scan: migration failed: %w", err)
			}

			fmt.Println("🔍 Sentinel: Scanning project AST...")

			// Inicializa o Engine Multi-Linguagem
			engine, err := graph.NewEngine(db)
			if err != nil {
				return fmt.Errorf("scan: failed to create engine: %w", err)
			}

			// Registra os Scanners
			engine.RegisterScanner(graph.NewGoScanner())
			engine.RegisterScanner(graph.NewTreeSitterScanner())

			err = engine.ScanProject(".")
			if err != nil {
				return fmt.Errorf("scan: failed: %w", err)
			}

		fmt.Println("✅ Scan complete. Graph database updated.")
		return nil
	}

	return cmd
}
