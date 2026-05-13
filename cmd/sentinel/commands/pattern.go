package commands

import (
	"context"
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

// NewPatternCmd creates a cobra command that provides subcommands for
// capturing, listing, searching, retrieving, and backfilling architectural
// and cognitive patterns.
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

func patternAddCmd(db *sqlite.DB) *cobra.Command {
	var title, desc, category, source, sourceRef, tags, impact string
	var force bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Capture a new pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(cmd.Context(), db); err != nil {
				return fmt.Errorf("pattern add: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			if !force {
				tagSlice := strings.Split(tags, ",")
				similar, err := store.FindSimilar(cmd.Context(), title, tagSlice)
				if err != nil {
					return fmt.Errorf("pattern add: dedup check failed: %w", err)
				}
				if len(similar) > 0 {
					fmt.Printf("[SENTINEL] Similar pattern found: %q (ID: %s)\n", similar[0].Title, similar[0].ID)
					fmt.Println("[SENTINEL] Use --force to create anyway.")
					return nil
				}
			}

			if source == "" {
				source = patterns.SourceManual
			}
			if impact == "" {
				impact = patterns.ImpactMedium
			}

			id, err := store.Create(cmd.Context(), &patterns.Pattern{
				Title:       title,
				Description: desc,
				Category:    category,
				Source:      source,
				SourceRef:   sourceRef,
				Tags:        tags,
				Impact:      impact,
			})
			if err != nil {
				return fmt.Errorf("pattern add: %w", err)
			}

			fmt.Printf("PATTERN CAPTURED [ID: %s]: %s\n", id, title)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Pattern title (required)")
	cmd.Flags().StringVar(&desc, "desc", "", "Pattern description (required)")
	cmd.Flags().StringVar(&category, "category", "", "Category: anti-pattern, cognitive-pattern, structural-principle, routing-principle (required)")
	cmd.Flags().StringVar(&source, "source", "", "Source: cognitive-dna, evolution-insights, sentinel-log, manual, epiphany (default: manual)")
	cmd.Flags().StringVar(&sourceRef, "source-ref", "", "Reference to original source location")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVar(&impact, "impact", "", "Impact: high, medium, low (default: medium)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip dedup check")

	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("desc")
	_ = cmd.MarkFlagRequired("category")

	return cmd
}

func patternListCmd(db *sqlite.DB) *cobra.Command {
	var category, source, impact string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List captured patterns",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(cmd.Context(), db); err != nil {
				return fmt.Errorf("pattern list: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			result, err := store.List(cmd.Context(), patterns.ListFilters{
				Category: category,
				Source:   source,
				Impact:   impact,
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

	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().StringVar(&source, "source", "", "Filter by source")
	cmd.Flags().StringVar(&impact, "impact", "", "Filter by impact")

	return cmd
}

func patternSearchCmd(db *sqlite.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Full-text search across patterns",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(cmd.Context(), db); err != nil {
				return fmt.Errorf("pattern search: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			result, err := store.Search(cmd.Context(), args[0])
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
			if err := graph.Migrate(cmd.Context(), db); err != nil {
				return fmt.Errorf("pattern get: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			p, err := store.Get(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("pattern get: pattern not found: %s", args[0])
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

func runBackfillCognitiveDNA(ctx context.Context, store *patterns.PatternStore, baseDir string) {
	result, err := store.BackfillFromCognitiveDNA(ctx, baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: cognitive-dna backfill: %v\n", err)
		return
	}
	fmt.Printf("Cognitive-DNA: %d extracted, %d inserted, %d skipped\n",
		result.Extracted, result.Inserted, result.Skipped)
}

func runBackfillEvolutionInsights(ctx context.Context, store *patterns.PatternStore, baseDir string) {
	result, err := store.BackfillFromEvolutionInsights(ctx, baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: evolution-insights backfill: %v\n", err)
		return
	}
	fmt.Printf("Evolution-Insights: %d extracted, %d inserted, %d skipped\n",
		result.Extracted, result.Inserted, result.Skipped)
}

func runBackfillSentinelLog(ctx context.Context, store *patterns.PatternStore, baseDir string) {
	result, err := store.BackfillFromSentinelLog(ctx, baseDir, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: sentinel-log backfill: %v\n", err)
		return
	}
	fmt.Printf("[DRY-RUN] %d candidates extracted from sentinel-log:\n", result.Extracted)
	for i, c := range result.Candidates {
		fmt.Printf(" %d. %q (%s)\n", i+1, c.Title, c.SourceRef)
	}
	fmt.Println("Use 'sentinel pattern add' to capture selected patterns.")
}

func patternBackfillCmd(db *sqlite.DB) *cobra.Command {
	var source string
	var all bool

	cmd := &cobra.Command{
		Use:   "backfill",
		Short: "Extract and insert patterns from existing documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := graph.Migrate(cmd.Context(), db); err != nil {
				return fmt.Errorf("pattern backfill: migration failed: %w", err)
			}
			store, err := patterns.NewPatternStore(db)
			if err != nil {
				return err
			}

			baseDir := "."

			if all || source == "cognitive-dna" {
				runBackfillCognitiveDNA(cmd.Context(), store, baseDir)
			}
			if all || source == "evolution-insights" {
				runBackfillEvolutionInsights(cmd.Context(), store, baseDir)
			}
			if source == "sentinel-log" {
				runBackfillSentinelLog(cmd.Context(), store, baseDir)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Source: cognitive-dna, evolution-insights, sentinel-log")
	cmd.Flags().BoolVar(&all, "all", false, "Backfill from cognitive-dna + evolution-insights (excludes sentinel-log)")

	return cmd
}
