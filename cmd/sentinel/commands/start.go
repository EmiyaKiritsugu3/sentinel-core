package commands

import (
	"context"
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/agents"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/reflect"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func NewStartCmd(db *sqlite.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "start [task_id]",
		Short: "Start the cognitive loop for a specific task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			mgr := state.NewManager(db)

			// 1. Mark task as IN_PROGRESS
			if err := mgr.StartTask(taskID); err != nil {
				return fmt.Errorf("start: failed to update task status: %w", err)
			}
			fmt.Printf("🚀 Sentinel: Task [%s] is now IN_PROGRESS.\n", taskID)

			// 2. Initialize Infrastructure
			auth := &agents.SovereignAuthProvider{}
			factory := bridge.NewFactory(db)
			validator := reflect.NewValidator(db)
			
			registry := agents.NewRegistry()
			agents.RegisterCoreTools(registry, db)
			
			engine, err := agents.NewEngine(registry, auth, factory, validator)
			if err != nil {
				return fmt.Errorf("start: failed to initialize engine: %w", err)
			}
			defer engine.Close()

			// 3. Load Agent Definition
			loader := agents.NewLoader()
			agentDef, err := loader.LoadAgent("internal/agents/definitions/architect.md")
			if err != nil {
				return fmt.Errorf("start: failed to load sovereign architect: %w", err)
			}

			// 4. Create Context and Execute
			ctx := agents.NewAgentContext(context.Background(), taskID, agentDef)
			fmt.Printf("🧠 Sentinel: Invoking '%s' (Model: %s)...\n", agentDef.Name, agentDef.ModelID)
			
			if err := engine.Execute(ctx); err != nil {
				return fmt.Errorf("start: cognitive loop execution failed: %w", err)
			}

			fmt.Printf("\n✅ Sentinel: Task [%s] execution cycle completed.\n", taskID)
			return nil
		},
	}
}
