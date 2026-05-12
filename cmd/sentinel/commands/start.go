package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/agents"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewStartCmd)
}

func NewStartCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [task_id]",
		Short: "Start the cognitive loop for a specific task",
		Args:  cobra.ExactArgs(1),
	}

	if err := sqlite.ValidateDB(db, "start-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			mgr, err := state.NewManager(db)
			if err != nil {
				return fmt.Errorf("start: failed to create manager: %w", err)
			}

			if err := mgr.StartTask(taskID); err != nil {
				return fmt.Errorf("start: failed to update task status: %w", err)
			}
			fmt.Printf("🚀 Sentinel: Task [%s] is now IN_PROGRESS.\n", taskID)

			auth := &agents.SovereignAuthProvider{}
			validator, err := reflect.NewValidator(db)
			if err != nil {
				return fmt.Errorf("start: failed to create validator: %w", err)
			}
			registry := agents.NewRegistry()
			agents.RegisterCoreTools(registry, db)

			// Dispatcher initialization
			gitShield := agents.NewGitShield(".", validator)
			regMgr, err := agents.NewRegistryManager(db)
			if err != nil {
				return fmt.Errorf("start: failed to create registry manager: %w", err)
			}
			dispatcher, err := agents.NewDispatcher(regMgr, gitShield, db)
			if err != nil {
				return fmt.Errorf("start: failed to create dispatcher: %w", err)
			}

			// Reconcile events from sub-agents before proceeding
			if err := dispatcher.ReconcileEvents(cmd.Context()); err != nil {
				return fmt.Errorf("start: event reconciliation failed: %w", err)
			}

			engine, err := agents.NewEngine(registry, auth, validator, db)
			if err != nil {
				if rollbackErr := mgr.UpdateStatus(taskID, "PENDING"); rollbackErr != nil {
					fmt.Printf("⚠️  Sentinel: Cognitive engine offline (%v). Rollback also failed: %v. Task may still be IN_PROGRESS.\n", err, rollbackErr)
				} else {
					fmt.Printf("⚠️  Sentinel: Cognitive engine offline (%v). Task reset to PENDING.\n", err)
				}
				return fmt.Errorf("start: cognitive engine failed to initialize: %w", err)
			}
			engine.Dispatcher = dispatcher // Wired for Phase 5.8
			defer engine.Close()

			loader := agents.NewLoader()
			agentDef, err := loader.LoadAgent("internal/agents/definitions/architect.md")
			if err != nil {
				return fmt.Errorf("start: failed to load sovereign architect: %w", err)
			}

			ctx := agents.NewAgentContext(cmd.Context(), taskID, agentDef)
			fmt.Printf("🧠 Sentinel: Invoking '%s' (Model: %s)...\n", agentDef.Name, agentDef.ModelID)

			if err := engine.Execute(ctx); err != nil {
				return fmt.Errorf("start: cognitive loop execution failed: %w", err)
			}

			fmt.Printf("\n✅ Sentinel: Task [%s] execution cycle completed.\n", taskID)
		return nil
	}

	return cmd
}
