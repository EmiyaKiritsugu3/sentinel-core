package commands

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the current governance status",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🛡️  Sovereign Council Status: ACTIVE")
		fmt.Println("Database: .sentinel/graph.db (Online)")

		mgr := state.NewManager(DBInstance)
		task, err := mgr.GetActiveTask()
		
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("\n✅ System Idle. No active tasks.")
				return nil
			}
			return fmt.Errorf("status: failed to query active task: %w", err)
		}

		fmt.Printf("\n🔥 ACTIVE TASK: [%s] %s\n", task.ID, task.Description)
		fmt.Printf("Tier: %s | Status: %s\n", task.Tier, task.Status)
		return nil
	},
}
