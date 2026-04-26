package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/report"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func NewReportCmd(db *sqlite.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Show a colorful compliance dashboard and export to Markdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("\n📊 SENTINEL COMPLIANCE REPORT")
			fmt.Println("======================================")

			agg := report.NewAggregator(db)
			stats, err := agg.FetchStats()
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

			// Export to MD
			err = agg.GenerateMarkdown(stats)
			if err != nil {
				fmt.Printf("\n⚠️  Markdown export failed: %v\n", err)
			} else {
				fmt.Printf("\n✅ Dashboard exported to: docs/process/COMPLIANCE-DASHBOARD.md\n")
			}
			fmt.Println("======================================\n")
			return nil
		},
	}
}
