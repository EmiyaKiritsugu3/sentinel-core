package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/report"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewReportCmd)
}

// NewReportCmd creates a cobra command that displays a colorful compliance
// dashboard with governance KPIs and exports the report to Markdown.
func NewReportCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Show a colorful compliance dashboard and export to Markdown",
	}

	if err := sqlite.ValidateDB(db, "report-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("\n📊 SENTINEL COMPLIANCE REPORT")
		fmt.Println("======================================")

		agg, err := report.NewAggregator(db)
		if err != nil {
			return fmt.Errorf("report: failed to create aggregator: %w", err)
		}
		stats, err := agg.FetchStats(cmd.Context())
		if err != nil {
			return fmt.Errorf("report: failed to fetch report data: %w", err)
		}

		// Cores ANSI
		colorReset := "\033[0m"
		colorGreen := "\033[32m"
		colorRed := "\033[31m"
		colorCyan := "\033[36m"
		colorYellow := "\033[33m"

		// KPI Section
		fmt.Printf("Health Score: ")
		scoreColor := colorGreen
		if stats.SuccessRate < 80 {
			scoreColor = colorYellow
		}
		if stats.SuccessRate < 50 {
			scoreColor = colorRed
		}
		fmt.Printf("%s%.2f%%%s\n", scoreColor, stats.SuccessRate, colorReset)

		fmt.Println("\n--- ARCHITECTURE ---")
		fmt.Printf("Total Nodes: %s%d%s\n", colorCyan, stats.TotalNodes, colorReset)
		fmt.Printf("Files:       %d\n", stats.TotalFiles)
		fmt.Printf("Functions:   %d\n", stats.TotalFunctions)
		fmt.Printf("Structs:     %d\n", stats.TotalStructs)

		fmt.Println("\n--- GOVERNANCE ---")
		fmt.Printf("Completed: %s%d%s\n", colorGreen, stats.CompletedTasks, colorReset)
		fmt.Printf("Failed:    %s%d%s\n", colorRed, stats.FailedTasks, colorReset)
		fmt.Printf("Total:     %d\n", stats.TotalTasks)

		fmt.Println("\n--- INTENT INVENTORY ---")
		if len(stats.Tasks) == 0 {
			fmt.Println("No intents captured yet.")
		} else {
			for _, t := range stats.Tasks {
				adrStatus := fmt.Sprintf("%sN/A%s", colorRed, colorReset)
				if t.ADRPath != "" {
					adrStatus = fmt.Sprintf("%s[LINKED]%s", colorGreen, colorReset)
				}
				fmt.Printf("[%s] %s | %-10s | %s %s\n", t.Tier, t.ID, t.Status, t.Description, adrStatus)
			}
		}

		// Export to MD
		err = agg.GenerateMarkdown(stats)
		if err != nil {
			fmt.Printf("\n⚠️  Markdown export failed: %v\n", err)
		} else {
			fmt.Printf("\n✅ Dashboard exported to: docs/process/COMPLIANCE-DASHBOARD.md\n")
		}
		fmt.Println("======================================")
		return nil
	}

	return cmd
}
