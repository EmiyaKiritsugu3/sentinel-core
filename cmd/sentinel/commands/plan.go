package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/intake"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewPlanCmd)
}

func NewPlanCmd(db *sqlite.DB) *cobra.Command {
	var planTier string
	var flagRefine bool
	var flagNoSuggest bool

	cmd := &cobra.Command{
		Use:   "plan [goal] [verification_command]",
		Short: "Create a new architectural plan and task",
		Args:  cobra.ExactArgs(2),
	}

	if err := sqlite.ValidateDB(db, "plan-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
			description := args[0]
			verifyCmd := args[1]

			// Flag conflict: --refine takes precedence
			if flagRefine && flagNoSuggest {
				flagNoSuggest = false
			}

			if !flagNoSuggest {
				disambiguator := intake.NewDisambiguator(db)
				vague, suggestions := disambiguator.Analyze(description)

				if vague && len(suggestions) > 0 {
					if flagRefine {
						// Interactive mode: show options, prompt user
						fmt.Println("[SUGGEST] Task description may be vague. Did you mean:")
						for i, s := range suggestions {
							fmt.Printf("  [%d] %s  (%s)\n", i+1, s.NodeName, s.FilePath)
						}
						fmt.Printf("  [0] Keep original: %q\n", description)
						fmt.Print("Choice [0]: ")

						scanner := bufio.NewScanner(os.Stdin)
						if scanner.Scan() {
							line := strings.TrimSpace(scanner.Text())
							if idx := parseChoice(line, len(suggestions)); idx > 0 {
								// Preserve action (first word) if possible
								action := ""
								if parts := strings.Fields(description); len(parts) > 0 {
									action = parts[0]
								}
								description = fmt.Sprintf("%s (focus: %s in %s)",
									action, suggestions[idx-1].NodeName, suggestions[idx-1].FilePath)
							}						}
					} else {
						// Default mode: print suggestion, save original
						fmt.Printf("[SUGGEST] did you mean: %s in %s?\n",
							suggestions[0].NodeName, suggestions[0].FilePath)
					}
				}
			}

			mgr, err := state.NewManager(db)
			if err != nil {
				return fmt.Errorf("plan: failed to create manager: %w", err)
			}
			id, err := mgr.CreateTask(description, planTier, verifyCmd)
			if err != nil {
				return fmt.Errorf("plan: failed to create task: %w", err)
			}

			fmt.Printf("✅ PLAN FORGED [ID: %s]: %s\n", id, description)
			fmt.Printf("Tier: %s | Verification Gate: %s\n", planTier, verifyCmd)
		return nil
	}
	cmd.Flags().StringVar(&planTier, "tier", "T2", "Task tier (T1, T2, T3)")
	cmd.Flags().BoolVarP(&flagRefine, "refine", "r", false, "interactive disambiguation before saving")
	cmd.Flags().BoolVar(&flagNoSuggest, "no-suggest", false, "skip suggestion output (for scripts and CI)")
	return cmd
}

func parseChoice(s string, max int) int {
	if s == "" || s == "0" {
		return 0
	}
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	if n < 1 || n > max {
		return 0
	}
	return n
}
