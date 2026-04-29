package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

type MutationEngine struct {
	DB *sqlite.DB
}

func NewMutationEngine(db *sqlite.DB) *MutationEngine {
	return &MutationEngine{DB: db}
}

var versionRegex = regexp.MustCompile(`-v\d+$`)

func (e *MutationEngine) Mutate(ctx context.Context, specialistID string, rcaPrompt string) error {
	var currentPath string
	var generation int
	err := e.DB.Conn.QueryRowContext(ctx, "SELECT current_persona_path, generation FROM specialist_registry WHERE id = ?", specialistID).Scan(&currentPath, &generation)
	if err != nil {
		return fmt.Errorf("mutation: failed to find specialist: %w", err)
	}

	content, err := os.ReadFile(currentPath)
	if err != nil {
		return fmt.Errorf("mutation: failed to read persona: %w", err)
	}

	newGeneration := generation + 1
	base := filepath.Base(currentPath)
	ext := filepath.Ext(base)
	nameOnly := base[:len(base)-len(ext)]
	
	// Robust version stripping
	cleanName := versionRegex.ReplaceAllString(nameOnly, "")
	newPath := filepath.Join(filepath.Dir(currentPath), fmt.Sprintf("%s-v%d%s", cleanName, newGeneration, ext))

	newContent := fmt.Sprintf("%s\n\n## Generation %d\n%s\n", string(content), newGeneration, rcaPrompt)
	if err := os.WriteFile(newPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("mutation: failed to write new persona: %w", err)
	}

	_, err = e.DB.Conn.ExecContext(ctx, "UPDATE specialist_registry SET current_persona_path = ?, generation = ? WHERE id = ?", newPath, newGeneration, specialistID)
	if err != nil {
		return fmt.Errorf("mutation: failed to update registry: %w", err)
	}

	return nil
}

func (e *MutationEngine) Rollback(ctx context.Context, specialistID string) error {
	// Simple rollback for now: decrement generation and assume path follows vN pattern
	// Real implementation would use a 'parent_persona_path' or history table.
	return nil
}
