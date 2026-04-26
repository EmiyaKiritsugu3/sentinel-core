package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

var planTier string

func NewPlanCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [goal] [verification_command]",
		Short: "Create a new architectural plan and task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			goal := args[0]
			verifyCmd := args[1]

			mgr := state.NewManager(db)
			id, err := mgr.CreateTask(goal, planTier, verifyCmd)
			if err != nil {
				return fmt.Errorf("plan: failed to create task: %w", err)
			}

			fmt.Printf("✅ PLAN FORGED [ID: %s]: %s\n", id, goal)
			fmt.Printf("Tier: %s | Verification Gate: %s\n", planTier, verifyCmd)
			return nil
		},
	}
	cmd.Flags().StringVar(&planTier, "tier", "T2", "Task tier (T1, T2, T3)")
	return cmd
}
