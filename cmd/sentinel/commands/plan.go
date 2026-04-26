package commands

import (
	"fmt"
	"log"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:   "plan [goal] [verification_command]",
	Short: "Create a new architectural plan and task",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		goal := args[0]
		verifyCmd := args[1]

		mgr := state.NewManager(DBInstance)
		id, err := mgr.CreateTask(goal, "T2", verifyCmd)
		if err != nil {
			log.Fatalf("❌ Failed to create task: %v", err)
		}

		fmt.Printf("✅ PLAN FORGED [ID: %s]: %s\n", id, goal)
		fmt.Printf("Verification Gate: %s\n", verifyCmd)
	},
}
