package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/context"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewContextCmd)
}

// NewContextCmd creates the sentinel context command, which queries the
// graphify knowledge graph and injects relevant context into AGENTS.md.
func NewContextCmd(db *sqlite.DB) *cobra.Command {
	_ = db // accepted for registry compatibility, not used by this command

	var limit int
	var budget int
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "context [query]",
		Short: "Inject graphify knowledge into AGENTS.md for AI context",
		Long: `Queries the graphify knowledge graph for the given topic and injects
relevant documents and concepts into AGENTS.md as a "Sentinel Context"
section. Any AI assistant that reads AGENTS.md will start with this context.

Requires: graphify installed and graphify-out/graph.json present.
Run "/graphify ." first if the graph has not been built.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			svc := context.NewContextService()

			result, err := svc.Query(query, budget)
			if err != nil {
				return err
			}

			content := context.Format(result, query, limit)

			if dryRun {
				fmt.Println(content)
				fmt.Println("\n[DRY RUN] Would inject into AGENTS.md")
				return nil
			}

			if err := context.Inject("AGENTS.md", content); err != nil {
				return fmt.Errorf("context: %w", err)
			}

			fmt.Printf("Context injected: %d docs, %d concepts → AGENTS.md\n",
				min(limit, len(result.Documents)), min(limit, len(result.Concepts)))
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 5, "Max documents/concepts to inject")
	cmd.Flags().IntVar(&budget, "budget", 2000, "Token budget for graphify query")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without modifying AGENTS.md")

	return cmd
}
