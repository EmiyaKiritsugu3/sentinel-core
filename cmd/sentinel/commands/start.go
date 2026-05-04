package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/agents"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
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
	return &cobra.Command{
		Use:   "start [task_id]",
		Short: "Start the cognitive loop for a specific task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			mgr := state.NewManager(db)

			if err := mgr.StartTask(taskID); err != nil {
				return fmt.Errorf("start: failed to update task status: %w", err)
			}
			fmt.Printf("🚀 Sentinel: Task [%s] is now IN_PROGRESS.\n", taskID)

			auth := &agents.SovereignAuthProvider{}
			factory := bridge.NewFactory(db)
			validator := reflect.NewValidator(db)
			registry := agents.NewRegistry()
			agents.RegisterCoreTools(registry, db)

			// Dispatcher initialization
			gitShield := agents.NewGitShield(".", validator)
			regMgr := agents.NewRegistryManager(db)
			dispatcher := agents.NewDispatcher(regMgr, gitShield, db)

			// Reconcile events from sub-agents before proceeding
			if err := dispatcher.ReconcileEvents(cmd.Context()); err != nil {
				return fmt.Errorf("start: event reconciliation failed: %w", err)
			}

			engine, err := agents.NewEngine(registry, auth, factory, validator, db)
			if err != nil {
				_ = mgr.UpdateStatus(taskID, "PENDING")
				fmt.Printf("⚠️  Sentinel: Cognitive engine offline (%v). Task reset to PENDING.\n", err)
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
		},
	}
}
