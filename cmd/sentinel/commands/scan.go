package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func NewScanCmd(db *sqlite.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan the project code to update the graph database",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			// Auto-Migração: garante que o banco esteja pronto
			err = graph.Migrate(db)
			if err != nil {
				return fmt.Errorf("scan: migration failed: %w", err)
			}

			fmt.Println("🔍 Sentinel: Scanning project AST...")
			
			// Inicializa o Engine Multi-Linguagem
			engine := graph.NewEngine(db)
			
			// Registra os Scanners
			engine.RegisterScanner(graph.NewGoScanner())
			engine.RegisterScanner(graph.NewTreeSitterScanner())
			
			err = engine.ScanProject(".")
			if err != nil {
				return fmt.Errorf("scan: failed: %w", err)
			}
			
			fmt.Println("✅ Scan complete. Graph database updated.")
			return nil
		},
	}
}
