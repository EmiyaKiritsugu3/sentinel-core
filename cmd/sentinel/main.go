package main

import (
	"fmt"
	"log"
	"os"

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
		Long: `The Sentinel is a high-performance governance wrapper that ensures architectural 
rigor and provides deterministic context loading for AI agents using AST analysis.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("🛡️ Sentinel Core v%s\n", version)
			fmt.Println("Status: Monitoring for compliance. Brain initialized.")
		},
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
		Use:   "plan",
		Short: "Create a new architectural plan and task",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				log.Fatal("❌ Goal description required: sentinel plan \"Your Goal\"")
			}
			
			goal := args[0]
			mgr := state.NewManager(db)
			id, err := mgr.CreateTask(goal, "T2") // Default T2 para o MVP
			if err != nil {
				log.Fatalf("❌ Failed to create task: %v", err)
			}
			
			fmt.Printf("✅ PLAN FORGED [ID: %s]: %s\n", id, goal)
			fmt.Println("Snapshot diagram generated in docs/architecture/tasks/")
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
