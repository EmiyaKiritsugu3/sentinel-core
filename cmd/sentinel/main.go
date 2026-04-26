package main

import (
	"fmt"
	"log"
	"os"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/audit"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

var (
	version = "5.0.0-alpha"
)

func main() {
	// Inicializa o "Cérebro" (SQLite)
	db, err := sqlite.Init()
	if err != nil {
		log.Fatalf("❌ Failed to initialize sentinel brain: %v", err)
	}
	defer db.Close()

	// Garante que o esquema está atualizado
	if err := graph.Migrate(db); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}

	var rootCmd = &cobra.Command{
		Use:   "sentinel",
		Short: "Sentinel Core: Governance & Context Engine for AI-Native Development",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Check the current governance status",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🛡️  Sovereign Council Status: ACTIVE")
			fmt.Println("Database: .sentinel/graph.db (Online)")
			
			mgr := state.NewManager(db)
			task, err := mgr.GetActiveTask()
			if err == nil {
				fmt.Printf("\n🔥 ACTIVE TASK: [%s] %s\n", task.ID, task.Description)
				fmt.Printf("Tier: %s | Status: %s\n", task.Tier, task.Status)
			} else {
				fmt.Println("\n✅ System Idle. No active tasks.")
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "plan [goal] [verification_command]",
		Short: "Create a new architectural plan and task",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			goal := args[0]
			verifyCmd := args[1]
			
			mgr := state.NewManager(db)
			id, err := mgr.CreateTask(goal, "T2", verifyCmd)
			if err != nil {
				log.Fatalf("❌ Failed to create task: %v", err)
			}
			
			fmt.Printf("✅ PLAN FORGED [ID: %s]: %s\n", id, goal)
			fmt.Printf("Verification Gate: %s\n", verifyCmd)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "start [task_id]",
		Short: "Start a specific task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mgr := state.NewManager(db)
			if err := mgr.StartTask(args[0]); err != nil {
				log.Fatalf("❌ Failed to start task: %v", err)
			}
			fmt.Printf("🚀 Task [%s] is now IN_PROGRESS.\n", args[0])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "audit",
		Short: "Run the verification gate for the active task",
		Run: func(cmd *cobra.Command, args []string) {
			mgr := state.NewManager(db)
			task, err := mgr.GetActiveTask()
			if err != nil {
				log.Fatal("❌ No active task found to audit. Run 'sentinel start <id>' first.")
			}

			// Busca o comando de verificação
			_, verifyCmd, err := mgr.GetTaskByID(task.ID)
			if err != nil {
				log.Fatalf("❌ Task record corrupted: %v", err)
			}

			runner := audit.NewRunner(db)
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
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "scan",
		Short: "Scan the project code to update the graph database",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔍 Sentinel: Scanning project AST...")
			scanner := graph.NewGoScanner(db)
			err := scanner.ScanProject(".")
			if err != nil {
				log.Fatalf("❌ Scan failed: %v", err)
			}
			fmt.Println("✅ Scan complete. Graph database updated.")
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "visualize",
		Short: "Generate architecture diagrams from the graph database",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🎨 Sentinel: Generating architectural maps...")
			viz := graph.NewVisualizer(db)
			
			// Gera o Mapa Mestre
			err := viz.GenerateMasterDiagram()
			if err != nil {
				log.Fatalf("❌ Visualization failed: %v", err)
			}
			
			fmt.Println("✅ MASTER-GRAPH.md generated in docs/architecture/")
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
