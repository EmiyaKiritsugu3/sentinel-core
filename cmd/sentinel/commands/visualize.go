package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(visualizeCmd)
}

var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Generate architecture diagrams from the graph database",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🎨 Sentinel: Generating architectural maps...")
		viz := graph.NewVisualizer(DBInstance)

		err := viz.GenerateMasterDiagram()
		if err != nil {
			return fmt.Errorf("visualize: failed: %w", err)
		}

		fmt.Println("✅ MASTER-GRAPH.md generated in docs/architecture/")
		return nil
	},
}
