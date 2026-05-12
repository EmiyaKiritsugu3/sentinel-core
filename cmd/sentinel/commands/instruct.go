package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewInstructCmd)
}

func NewInstructCmd(db *sqlite.DB) *cobra.Command {
	var message string
	var quick bool

	cmd := &cobra.Command{
		Use:   "instruct",
		Short: "Interview mode to capture user intent and generate tasks",
	}

	if err := sqlite.ValidateDB(db, "instruct-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
			var intent string

			// 1. Prioridade: Flag explícita ou Stdin
			if message != "" {
				intent = message
			} else {
				stat, err := os.Stdin.Stat()
				if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						intent = strings.TrimSpace(scanner.Text())
					}
				} else {
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
				return nil
			}

			// 2. Sovereign Triage & Data-Driven Scout
			fmt.Printf("\n🔍 Sentinel: Analisando '%s'...\n", intent)

			// Busca por 'God Objects' ou pontos de impacto se a intenção for vaga
			isVague := isVagueIntent(intent)
			var evidence string
			if isVague && !quick {
				evidence = performDiagnostic(cmd.Context(), db)
				if evidence != "" {
					fmt.Println("\n⚠️  EVIDÊNCIA ENCONTRADA (Sessão de Diagnóstico):")
					fmt.Println(evidence)
					fmt.Println("\nIsto parece ser o ponto de partida ideal para evitar 'guessing'.")
				}
			}

			// 3. Sovereign Menu
			var choice string
			if quick {
				choice = "q"
			} else {
				fmt.Println("\nComo deseja preencher os detalhes técnicos do ADR?")
				fmt.Println("[m] Manual (Entrevista Socrática)")
				fmt.Println("[a] IA Now (Sugestão via Gemini)")
				fmt.Println("[s] Sentinel (Delegar ao Agente durante a execução)")
				fmt.Println("[q] Quick (Usar placeholders)")
				fmt.Print("\nEscolha> ")
				fmt.Scanln(&choice)
			}

			adrData := graph.ADRData{
				Title:  intent,
				Status: "PROPOSED",
			}

			switch choice {
			case "m":
				adrData = runSocraticInterview(intent, evidence)
			case "a":
				fmt.Println("✨ Chamando AI Bridge para expansão... (Simulado na Fase 4.1)")
				adrData.Context = "Expandido via IA baseado em: " + intent + "\n" + evidence
				adrData.Decision = "Padrão recomendado pela IA para este cenário."
				adrData.VerificationCommand = "go test ./..."
			case "s":
				adrData.Status = "DRAFT"
				adrData.Context = "Aguardando refinamento pelo Sentinel Agent."
				adrData.VerificationCommand = "# O Agente definirá o comando de prova"
			default: // q
				adrData.Context = "Capturado via comando 'instruct'.\nIntenção: " + intent
				adrData.Decision = "[Descreva a abordagem técnica]"
				adrData.VerificationCommand = "go build ./..."
			}

			// 4. Persistência
			if adrData.Status != "DRAFT" && strings.TrimSpace(adrData.VerificationCommand) == "" {
				return fmt.Errorf("instruct: verification command is required for non-draft ADRs")
			}

			manager, err := state.NewManager(db)
			if err != nil {
				return fmt.Errorf("instruct: failed to create manager: %w", err)
			}
			id, err := manager.CreateTask(intent, "T1", adrData.VerificationCommand)
			if err != nil {
				return fmt.Errorf("instruct: failed to create task: %w", err)
			}

			adrData.TaskID = id
			gen := graph.NewADRGenerator()
			adrPath, err := gen.Generate(adrData)
			if err != nil {
				fmt.Printf("\n⚠️  ADR Generation failed: %v\n", err)
			} else {
				fmt.Printf("\n📄 ADR Gerado: %s\n", adrPath)
				fmt.Printf("✅ Task [%s] criada com Protocolo de Verificação: %s\n", id, adrData.VerificationCommand)
			}

		return nil
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "User intent message")
	cmd.Flags().BoolVarP(&quick, "quick", "q", false, "Skip interview and use defaults")
	return cmd
}

func isVagueIntent(intent string) bool {
	return len(strings.Split(intent, " ")) < 3 || strings.Contains(strings.ToLower(intent), "performance")
}

func performDiagnostic(ctx context.Context, db *sqlite.DB) string {
	// Query real no banco para encontrar os 3 arquivos mais complexos (God Objects)
	query := `
		SELECT n.file_path, COUNT(e.from_node_id) as degree
		FROM nodes n
		JOIN edges e ON n.id = e.from_node_id
		WHERE n.type = 'file'
		GROUP BY n.file_path
		ORDER BY degree DESC
		LIMIT 3
	`
	rows, err := db.Conn.QueryContext(ctx, query)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var evidence strings.Builder
	evidence.WriteString("Baseado no Grafo de Dependências, os arquivos com maior carga de complexidade são:\n")
	for rows.Next() {
		var path string
		var degree int
		if err := rows.Scan(&path, &degree); err == nil {
			fmt.Fprintf(&evidence, "- %s (%d conexões)\n", path, degree)
		}
	}

	if err := rows.Err(); err != nil {
		return ""
	}

	return evidence.String()
}

func runSocraticInterview(intent string, evidence string) graph.ADRData {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n--- ENTREVISTA ADR ---")

	fmt.Println("\n1. Qual o CONTEXTO ou problema técnico? (Enter para aceitar sugestão)")
	if evidence != "" {
		fmt.Printf("Sugestão baseada em dados: %s\n", evidence)
	}
	fmt.Print("> ")
	context, _ := reader.ReadString('\n')
	if strings.TrimSpace(context) == "" {
		context = evidence
	}

	fmt.Println("\n2. Qual a DECISÃO técnica tomada?")
	fmt.Print("> ")
	decision, _ := reader.ReadString('\n')

	fmt.Println("\n3. Qual o COMANDO DE VERIFICAÇÃO (ex: go test)?")
	fmt.Print("> ")
	verify, _ := reader.ReadString('\n')

	return graph.ADRData{
		Title:               intent,
		Context:             strings.TrimSpace(context),
		Decision:            strings.TrimSpace(decision),
		VerificationCommand: strings.TrimSpace(verify),
		Status:              "ACCEPTED",
	}
}
