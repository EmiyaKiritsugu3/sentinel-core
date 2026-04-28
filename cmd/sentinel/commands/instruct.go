package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func NewInstructCmd(db *sqlite.DB) *cobra.Command {
	var message string
	cmd := &cobra.Command{
		Use:   "instruct",
		Short: "Interview mode to capture user intent and generate tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			var intent string

			// 1. Prioridade: Flag explícita
			if message != "" {
				intent = message
			} else {
				// 2. Verifica Stdin (Pipe vs TTY)
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					// Dados via Pipe/Redirecionamento
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						intent = strings.TrimSpace(scanner.Text())
					}
				} else {
					// 3. Modo Interativo (Humano)
					fmt.Println("\n🧠 SENTINEL INTERVIEW MODE")
					fmt.Println("======================================")
					fmt.Println("O que você deseja construir hoje?")
					fmt.Print("> ")

					reader := bufio.NewReader(os.Stdin)
					line, _ := reader.ReadString('\n')
					intent = strings.TrimSpace(line)
				}
			}

			if intent == "" {
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					return fmt.Errorf("instruct: non-interactive environment detected. Provide a message via -m or piped stdin")
				}
				fmt.Println("⚠️  Nenhuma instrução capturada. Operação cancelada.")
				return nil
			}

			// Salva a intenção como uma nova tarefa T1 (Trivial/Proativa)
			manager := state.NewManager(db)
			id, err := manager.CreateTask(intent, "T1", "go build ./...")
			if err != nil {
				return fmt.Errorf("instruct: failed to create task: %w", err)
			}

			// Gera o ADR Automático (Fase 4.2)
			gen := graph.NewADRGenerator()
			adrPath, err := gen.Generate(id, intent)
			if err != nil {
				fmt.Printf("\n⚠️  ADR Generation failed: %v (Task created nonetheless)\n", err)
			} else {
				fmt.Printf("\n📄 ADR Gerado: %s\n", adrPath)
			}

			fmt.Printf("\n✅ Intenção capturada e registrada como tarefa: [%s]\n", id)
			fmt.Println("Sentinel agora está monitorando este objetivo.")
			fmt.Println("======================================")
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "User intent message")
	return cmd
}
