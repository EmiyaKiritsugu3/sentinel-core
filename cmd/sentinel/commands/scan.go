package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the project code to update the graph database",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔍 Sentinel: Scanning project AST...")
		scanner := graph.NewGoScanner(DBInstance)
		err := scanner.ScanProject(".")
		if err != nil {
			return fmt.Errorf("scan: failed: %w", err)
		}
		fmt.Println("✅ Scan complete. Graph database updated.")
		return nil
	},
}
