package commands

import (
	"fmt"
	"log"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(visualizeCmd)
}

var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Generate architecture diagrams from the graph database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🎨 Sentinel: Generating architectural maps...")
		viz := graph.NewVisualizer(DBInstance)

		err := viz.GenerateMasterDiagram()
		if err != nil {
			log.Fatalf("❌ Visualization failed: %v", err)
		}

		fmt.Println("✅ MASTER-GRAPH.md generated in docs/architecture/")
	},
}
