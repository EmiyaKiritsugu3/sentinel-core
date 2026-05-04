package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewStatusCmd)
}

func NewStatusCmd(db *sqlite.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check the current governance status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🛡️  Sovereign Council Status: ACTIVE")
			fmt.Println("Database: Online")
			fmt.Println("")

			mgr := state.NewManager(db)
			task, err := mgr.GetActiveTask()
			if err != nil {
				fmt.Println("✅ System Idle. No active tasks.")
				return nil
			}

			fmt.Println("🚀 ACTIVE TASK:")
			fmt.Printf("   ID:     %s\n", task.ID)
			fmt.Printf("   Goal:   %s\n", task.Description)
			fmt.Printf("   Tier:   %s\n", task.Tier)
			fmt.Printf("   Status: %s\n", task.Status)

			return nil
		},
	}
}
