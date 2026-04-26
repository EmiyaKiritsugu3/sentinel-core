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
			fmt.Println("🔍 Sentinel: Scanning project AST...")
			scanner := graph.NewGoScanner(db)
			err := scanner.ScanProject(".")
			if err != nil {
				return fmt.Errorf("scan: failed: %w", err)
			}
			fmt.Println("✅ Scan complete. Graph database updated.")
			return nil
		},
	}
}
