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

// NewScanCmd creates a cobra command that scans the project codebase to
// update the graph database with current AST information.
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
		// Auto-Migration: ensures the database is ready
		err = graph.Migrate(cmd.Context(), db)
		if err != nil {
			return fmt.Errorf("scan: migration failed: %w", err)
		}

		fmt.Println("🔍 Sentinel: Scanning project AST...")

		// Initializes the Multi-Language Engine
		engine, err := graph.NewEngine(db)
		if err != nil {
			return fmt.Errorf("scan: failed to create engine: %w", err)
		}

		// Registra os Scanners
		engine.RegisterScanner(graph.NewGoScanner())
		engine.RegisterScanner(graph.NewTreeSitterScanner())

		err = engine.ScanProject(cmd.Context(), ".")
		if err != nil {
			return fmt.Errorf("scan: failed: %w", err)
		}

		fmt.Println("✅ Scan complete. Graph database updated.")
		return nil
	}

	return cmd
}
