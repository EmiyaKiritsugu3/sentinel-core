package commands

import (
	"errors"
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/audit"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewAuditCmd)
}

func NewAuditCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Run the verification gate for the active task",
	}

	if err := sqlite.ValidateDB(db, "audit-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
			mgr, err := state.NewManager(db)
			if err != nil {
				return fmt.Errorf("audit: failed to create manager: %w", err)
			}
			task, err := mgr.GetActiveTask()
			if err != nil {
				return fmt.Errorf("audit: no active task found. Run 'sentinel start <id>' first: %w", err)
			}

			// 1. Sovereign Gate: Validação de Padrões
			fmt.Println("🛡️  Sentinel: Running Sovereign Validator...")
			validator, err := reflect.NewValidator(db)
			if err != nil {
				return fmt.Errorf("audit: failed to create validator: %w", err)
			}
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

			runner, err := audit.NewRunner(db)
			if err != nil {
				return fmt.Errorf("audit: failed to create runner: %w", err)
			}
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
	}

	return cmd
}
