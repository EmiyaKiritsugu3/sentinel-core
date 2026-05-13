package commands

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewStatusCmd)
}

// NewStatusCmd creates a cobra command that checks and displays the current
// governance status, including any active task.
func NewStatusCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check the current governance status",
	}

	if err := sqlite.ValidateDB(db, "status-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("🛡️  Sovereign Council Status: ACTIVE")
		fmt.Println("Database: Online")
		fmt.Println("")

		mgr, err := state.NewManager(db)
		if err != nil {
			return fmt.Errorf("status: failed to create manager: %w", err)
		}
		task, err := mgr.GetActiveTask(cmd.Context())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("✅ System Idle. No active tasks.")
				return nil
			}
			return fmt.Errorf("failed to read active task: %w", err)
		}

		fmt.Println("🚀 ACTIVE TASK:")
		fmt.Printf("   ID:     %s\n", task.ID)
		fmt.Printf("   Goal:   %s\n", task.Description)
		fmt.Printf("   Tier:   %s\n", task.Tier)
		fmt.Printf("   Status: %s\n", task.Status)

		return nil
	}

	return cmd
}
