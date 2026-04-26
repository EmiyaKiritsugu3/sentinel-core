package commands

import (
	"fmt"
	"log"

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
	Run: func(cmd *cobra.Command, args []string) {
		mgr := state.NewManager(DBInstance)
		if err := mgr.StartTask(args[0]); err != nil {
			log.Fatalf("❌ Failed to start task: %v", err)
		}
		fmt.Printf("🚀 Task [%s] is now IN_PROGRESS.\n", args[0])
	},
}
