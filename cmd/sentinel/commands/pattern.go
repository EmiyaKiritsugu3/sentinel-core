package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/patterns"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewPatternCmd)
}

func NewPatternCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pattern",
		Short: "Capture and query architectural and cognitive patterns",
	}

	if err := sqlite.ValidateDB(db, "pattern-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.AddCommand(patternAddCmd(db))
	cmd.AddCommand(patternListCmd(db))
	cmd.AddCommand(patternSearchCmd(db))
	cmd.AddCommand(patternGetCmd(db))
	cmd.AddCommand(patternBackfillCmd(db))

	return cmd
}

var (
	addTitle    string
	addDesc     string
	addCategory string
	addSource   string
	addSourceRef string
	addTags     string
	addImpact   string
	addForce    bool
)

func patternAddCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Capture a new pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(db); err != nil {
				return fmt.Errorf("pattern add: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			if !addForce {
				tagSlice := strings.Split(addTags, ",")
				similar, _ := store.FindSimilar(addTitle, tagSlice)
				if len(similar) > 0 {
					fmt.Printf("[SENTINEL] Similar pattern found: %q (ID: %s)\n", similar[0].Title, similar[0].ID)
					fmt.Println("[SENTINEL] Use --force to create anyway.")
					return nil
				}
			}

			if addSource == "" {
				addSource = patterns.SourceManual
			}
			if addImpact == "" {
				addImpact = patterns.ImpactMedium
			}

			id, err := store.Create(&patterns.Pattern{
				Title:       addTitle,
				Description: addDesc,
				Category:    addCategory,
				Source:      addSource,
				SourceRef:   addSourceRef,
				Tags:        addTags,
				Impact:      addImpact,
			})
			if err != nil {
				return fmt.Errorf("pattern add: %w", err)
			}

			fmt.Printf("PATTERN CAPTURED [ID: %s]: %s\n", id, addTitle)
			return nil
		},
	}

	cmd.Flags().StringVar(&addTitle, "title", "", "Pattern title (required)")
	cmd.Flags().StringVar(&addDesc, "desc", "", "Pattern description (required)")
	cmd.Flags().StringVar(&addCategory, "category", "", "Category: anti-pattern, cognitive-pattern, structural-principle, routing-principle (required)")
	cmd.Flags().StringVar(&addSource, "source", "", "Source: cognitive-dna, evolution-insights, sentinel-log, manual, epiphany (default: manual)")
	cmd.Flags().StringVar(&addSourceRef, "source-ref", "", "Reference to original source location")
	cmd.Flags().StringVar(&addTags, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVar(&addImpact, "impact", "", "Impact: high, medium, low (default: medium)")
	cmd.Flags().BoolVar(&addForce, "force", false, "Skip dedup check")

	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("desc")
	cmd.MarkFlagRequired("category")

	return cmd
}

var (
	listCategory string
	listSource   string
	listImpact   string
)

func patternListCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List captured patterns",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(db); err != nil {
				return fmt.Errorf("pattern list: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			result, err := store.List(patterns.ListFilters{
				Category: listCategory,
				Source:   listSource,
				Impact:   listImpact,
				Limit:    20,
			})
			if err != nil {
				return fmt.Errorf("pattern list: %w", err)
			}

			if len(result) == 0 {
				fmt.Println("No patterns found.")
				return nil
			}

			fmt.Printf("%-38s %-40s %-20s %-8s %-15s\n", "ID", "TITLE", "CATEGORY", "IMPACT", "SOURCE")
			for _, p := range result {
				title := p.Title
				if len(title) > 38 {
					title = title[:35] + "..."
				}
				fmt.Printf("%-38s %-40s %-20s %-8s %-15s\n",
					p.ID, title, p.Category, p.Impact, p.Source)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&listCategory, "category", "", "Filter by category")
	cmd.Flags().StringVar(&listSource, "source", "", "Filter by source")
	cmd.Flags().StringVar(&listImpact, "impact", "", "Filter by impact")

	return cmd
}

func patternSearchCmd(db *sqlite.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Full-text search across patterns",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(db); err != nil {
				return fmt.Errorf("pattern search: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			result, err := store.Search(args[0])
			if err != nil {
				return fmt.Errorf("pattern search: %w", err)
			}

			if len(result) == 0 {
				fmt.Printf("No patterns found matching %q.\n", args[0])
				return nil
			}

			fmt.Printf("%-38s %-40s %-20s %-8s\n", "ID", "TITLE", "CATEGORY", "IMPACT")
			for _, p := range result {
				title := p.Title
				if len(title) > 38 {
					title = title[:35] + "..."
				}
				fmt.Printf("%-38s %-40s %-20s %-8s\n", p.ID, title, p.Category, p.Impact)
			}
			return nil
		},
	}
}

func patternGetCmd(db *sqlite.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Show full details of a pattern",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(db); err != nil {
				return fmt.Errorf("pattern get: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			p, err := store.Get(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Pattern not found: %s\n", args[0])
				os.Exit(1)
			}

			fmt.Printf("ID: %s\n", p.ID)
			fmt.Printf("Title: %s\n", p.Title)
			fmt.Printf("Description: %s\n", p.Description)
			fmt.Printf("Category: %s\n", p.Category)
			fmt.Printf("Source: %s\n", p.Source)
			fmt.Printf("Source Ref: %s\n", p.SourceRef)
			fmt.Printf("Tags: %s\n", p.Tags)
			fmt.Printf("Impact: %s\n", p.Impact)
			fmt.Printf("Created: %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", p.UpdatedAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
}

var (
	backfillSource string
	backfillAll    bool
)

func runBackfillCognitiveDNA(store *patterns.PatternStore, baseDir string) {
	result, err := store.BackfillFromCognitiveDNA(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: cognitive-dna backfill: %v\n", err)
		return
	}
	fmt.Printf("Cognitive-DNA: %d extracted, %d inserted, %d skipped\n",
		result.Extracted, result.Inserted, result.Skipped)
}

func runBackfillEvolutionInsights(store *patterns.PatternStore, baseDir string) {
	result, err := store.BackfillFromEvolutionInsights(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: evolution-insights backfill: %v\n", err)
		return
	}
	fmt.Printf("Evolution-Insights: %d extracted, %d inserted, %d skipped\n",
		result.Extracted, result.Inserted, result.Skipped)
}

func runBackfillSentinelLog(store *patterns.PatternStore, baseDir string) {
	candidates, err := store.BackfillFromSentinelLog(baseDir, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: sentinel-log backfill: %v\n", err)
		return
	}
	fmt.Printf("[DRY-RUN] %d candidates extracted from sentinel-log:\n", len(candidates))
	for i, c := range candidates {
		fmt.Printf(" %d. %q (%s)\n", i+1, c.Title, c.SourceRef)
	}
	fmt.Println("Use 'sentinel pattern add' to capture selected patterns.")
}

func patternBackfillCmd(db *sqlite.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backfill",
		Short: "Extract and insert patterns from existing documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(db); err != nil {
				return fmt.Errorf("pattern backfill: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			baseDir := "."

			if backfillAll || backfillSource == "cognitive-dna" {
				runBackfillCognitiveDNA(store, baseDir)
			}
			if backfillAll || backfillSource == "evolution-insights" {
				runBackfillEvolutionInsights(store, baseDir)
			}
			if backfillSource == "sentinel-log" {
				runBackfillSentinelLog(store, baseDir)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&backfillSource, "source", "", "Source: cognitive-dna, evolution-insights, sentinel-log")
	cmd.Flags().BoolVar(&backfillAll, "all", false, "Backfill from cognitive-dna + evolution-insights (excludes sentinel-log)")

	return cmd
}
