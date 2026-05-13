package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewVisualizeCmd)
}

// NewVisualizeCmd creates a cobra command that generates architecture
// diagrams (master graph and C4 container) from the graph database.
func NewVisualizeCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visualize",
		Short: "Generate architecture diagrams from the graph database",
	}

	if err := sqlite.ValidateDB(db, "visualize-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("🎨 Sentinel: Generating architectural maps...")
		viz, err := graph.NewVisualizer(db)
		if err != nil {
			return fmt.Errorf("visualize: failed to create visualizer: %w", err)
		}

		err = viz.GenerateMasterDiagram(cmd.Context())
		if err != nil {
			return fmt.Errorf("visualize: master graph failed: %w", err)
		}

		err = viz.GenerateC4ContainerDiagram(cmd.Context())
		if err != nil {
			return fmt.Errorf("visualize: C4 container diagram failed: %w", err)
		}

		fmt.Println("✅ MASTER-GRAPH.md and C4-CONTAINER-GRAPH.md generated in docs/architecture/")
		return nil
	}

	return cmd
}
