package commands

import (
	"fmt"
	"log"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the project code to update the graph database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Sentinel: Scanning project AST...")
		scanner := graph.NewGoScanner(DBInstance)
		err := scanner.ScanProject(".")
		if err != nil {
			log.Fatalf("❌ Scan failed: %v", err)
		}
		fmt.Println("✅ Scan complete. Graph database updated.")
	},
}
