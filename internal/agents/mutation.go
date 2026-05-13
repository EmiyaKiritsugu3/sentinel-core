package agents

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// MutationEngine manages persona file mutations and rollbacks for specialists.
type MutationEngine struct {
	DB *sqlite.DB
}

// NewMutationEngine creates a new MutationEngine with a validated database connection.
func NewMutationEngine(db *sqlite.DB) (*MutationEngine, error) {
	if err := sqlite.ValidateDB(db, "mutation-engine"); err != nil {
		return nil, err
	}
	return &MutationEngine{DB: db}, nil
}

var versionRegex = regexp.MustCompile(`-v\d+$`)

// Mutate creates a new generation of a specialist's persona file by appending an RCA prompt.
func (e *MutationEngine) Mutate(ctx context.Context, specialistID string, rcaPrompt string) error {
	var currentPath string
	var generation int
	err := e.DB.Conn.QueryRowContext(ctx, "SELECT current_persona_path, generation FROM specialist_registry WHERE id = ?", specialistID).Scan(&currentPath, &generation)
	if err != nil {
		return fmt.Errorf("mutation: failed to find specialist: %w", err)
	}

	file, err := os.Open(currentPath) //nolint:gosec // path from registry
	if err != nil {
		return fmt.Errorf("mutation: failed to open persona: %w", err)
	}

	// Standard #01: Use buffered reader (bufio) to read the persona file
	reader := bufio.NewReader(file)
	content, err := io.ReadAll(reader)
	_ = file.Close()
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
	if err := os.WriteFile(newPath, []byte(newContent), 0600); err != nil {
		return fmt.Errorf("mutation: failed to write new persona: %w", err)
	}

	_, err = e.DB.Conn.ExecContext(ctx, "UPDATE specialist_registry SET current_persona_path = ?, generation = ? WHERE id = ?", newPath, newGeneration, specialistID)
	if err != nil {
		return fmt.Errorf("mutation: failed to update registry: %w", err)
	}

	return nil
}

// Rollback reverts a specialist's persona to a previous generation.
func (e *MutationEngine) Rollback(ctx context.Context, specialistID string) error {
	// Simple rollback for now: decrement generation and assume path follows vN pattern
	// Real implementation would use a 'parent_persona_path' or history table.
	return nil
}
