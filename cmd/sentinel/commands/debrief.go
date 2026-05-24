package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/knowledge"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewDebriefCmd)
}

// NewDebriefCmd creates the sentinel debrief command, which generates a
// session debrief from captured events and persists it to ~/knowledge/.
func NewDebriefCmd(db *sqlite.DB) *cobra.Command {
	var auto, dryRun bool
	var editor bool
	var outputPath string

	cmd := &cobra.Command{
		Use:   "debrief",
		Short: "Generate session debrief from captured events",
		Long: `Debrief captures decisions, errors, patterns, and file changes
from the current sentinel session and saves them to ~/knowledge/sessions/.

Events are collected automatically via the EventBuffer. Run this command
at the end of your session to persist captured knowledge.`,
	}

	cmd.Flags().BoolVar(&auto, "auto", false, "Skip prompts, save all captured events")
	cmd.Flags().BoolVar(&editor, "editor", false, "Open in $EDITOR instead of interactive prompts")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be saved, don't persist")
	cmd.Flags().StringVar(&outputPath, "output", "", "Override output path")

	if err := sqlite.ValidateDB(db, "debrief-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := graph.Migrate(cmd.Context(), db); err != nil {
			return fmt.Errorf("debrief: migration failed: %w", err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("debrief: cannot find home directory: %w", err)
		}
		baseDir := fmt.Sprintf("%s/knowledge", homeDir)

		svc := knowledge.NewDebriefService(knowledge.GlobalBuffer, db, baseDir)
		content := svc.Generate()

		if dryRun {
			fmt.Println(content)
			fmt.Printf("\n[DRY RUN] Would save to %s/knowledge/sessions/\n", homeDir)
			return nil
		}

		if auto {
			id, path, err := svc.Save(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Printf("Saved: %s (session %s, %d events)\n", path, id, knowledge.GlobalBuffer.Len())
			return nil
		}

		if editor {
			return openInEditor(content, svc, cmd.Context())
		}

		return interactiveDebrief(content, svc, cmd.Context())
	}

	return cmd
}

func interactiveDebrief(content string, svc *knowledge.DebriefService, ctx context.Context) error {
	fmt.Println(content)
	fmt.Print("\nSave this debrief? [Y/n]: ")
	var answer string
	fmt.Scanln(&answer)
	if answer == "" || answer == "y" || answer == "Y" {
		id, path, err := svc.Save(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Saved: %s (session %s)\n", path, id)
	} else {
		fmt.Println("Debrief discarded.")
	}
	return nil
}

func openInEditor(content string, svc *knowledge.DebriefService, ctx context.Context) error {
	tmpFile, err := os.CreateTemp("", "sentinel-debrief-*.md")
	if err != nil {
		return fmt.Errorf("debrief: create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("debrief: write temp file: %w", err)
	}
	tmpFile.Close()

	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vi"
	}
	parts := strings.Fields(editorCmd)
	c := exec.Command(parts[0], append(parts[1:], tmpFile.Name())...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("debrief: editor failed: %w", err)
	}

	edited, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("debrief: read edited file: %w", err)
	}

	fmt.Println(string(edited))
	fmt.Print("\nSave this debrief? [Y/n]: ")
	var answer string
	fmt.Scanln(&answer)
	if answer == "" || answer == "y" || answer == "Y" {
		id, path, err := svc.Save(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Saved: %s (session %s)\n", path, id)
	} else {
		fmt.Println("Debrief discarded.")
	}
	return nil
}
