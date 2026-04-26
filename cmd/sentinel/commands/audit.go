package commands

import (
	"errors"
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/audit"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(auditCmd)
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run the verification gate for the active task",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := state.NewManager(DBInstance)
		task, err := mgr.GetActiveTask()
		if err != nil {
			return fmt.Errorf("audit: no active task found. Run 'sentinel start <id>' first: %w", err)
		}

		// 1. Sovereign Gate: Validação de Padrões
		fmt.Println("🛡️  Sentinel: Running Sovereign Validator...")
		validator := reflect.NewValidator(DBInstance)
		violations, err := validator.ValidateProject(".")
		if err != nil {
			return fmt.Errorf("audit: validator internal error: %w", err)
		}

		if len(violations) > 0 {
			fmt.Printf("🛑 ARCHITECTURAL VIOLATIONS DETECTED (%d):\n", len(violations))
			for _, v := range violations {
				fmt.Printf("   - [%s] %s:%d: %s\n", v.StandardID, v.FilePath, v.Line, v.Reason)
			}
			_ = mgr.UpdateStatus(task.ID, "FAILED")
			return errors.New("task rejected by Sovereign Validator. Fix the standards and try again")
		}

		// 2. Technical Gate: Build & Tests
		_, verifyCmd, err := mgr.GetTaskByID(task.ID)
		if err != nil {
			return fmt.Errorf("audit: task record corrupted: %w", err)
		}

		runner := audit.NewRunner(DBInstance)
		success, err := runner.ExecuteAudit(task.ID, verifyCmd)
		if err != nil {
			return fmt.Errorf("audit: execution error: %w", err)
		}

		if success {
			if err := mgr.UpdateStatus(task.ID, "DONE"); err != nil {
				return fmt.Errorf("audit: failed to mark task as DONE: %w", err)
			}
			fmt.Println("🏆 Task marked as DONE. Commit authorized.")
		} else {
			_ = mgr.UpdateStatus(task.ID, "FAILED")
			return errors.New("audit failed. Fix the code and try again")
		}
		return nil
	},
}
