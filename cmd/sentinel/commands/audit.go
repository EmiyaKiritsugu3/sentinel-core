package commands

import (
	"fmt"
	"log"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/audit"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(auditCmd)
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run the verification gate for the active task",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := state.NewManager(DBInstance)
		task, err := mgr.GetActiveTask()
		if err != nil {
			log.Fatal("❌ No active task found to audit. Run 'sentinel start <id>' first.")
		}

		_, verifyCmd, err := mgr.GetTaskByID(task.ID)
		if err != nil {
			log.Fatalf("❌ Task record corrupted: %v", err)
		}

		runner := audit.NewRunner(DBInstance)
		success, err := runner.ExecuteAudit(task.ID, verifyCmd)
		if err != nil {
			log.Fatalf("❌ Audit execution error: %v", err)
		}

		if success {
			mgr.UpdateStatus(task.ID, "DONE")
			fmt.Println("🏆 Task marked as DONE. Commit authorized.")
		} else {
			mgr.UpdateStatus(task.ID, "FAILED")
			fmt.Println("🛑 Task marked as FAILED. Fix the code and try again.")
		}
	},
}
