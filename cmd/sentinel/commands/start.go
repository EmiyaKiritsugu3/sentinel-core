package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start [task_id]",
	Short: "Start a specific task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := state.NewManager(DBInstance)
		if err := mgr.StartTask(args[0]); err != nil {
			return fmt.Errorf("start: failed to start task %s: %w", args[0], err)
		}
		fmt.Printf("🚀 Task [%s] is now IN_PROGRESS.\n", args[0])
		return nil
	},
}
