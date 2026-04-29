package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// MutationEngine handles the evolution of specialist personas based on RCA feedback.
type MutationEngine struct {
	DB *sqlite.DB
}

// Mutate takes a specialistID and an RCA prompt, simulates an evolution, 
// saves a new versioned persona file, and updates the registry.
func (e *MutationEngine) Mutate(ctx context.Context, specialistID string, rcaPrompt string) error {
	// 1. Get current specialist info from DB
	var currentPath string
	var generation int
	query := `SELECT current_persona_path, generation FROM specialist_registry WHERE id = ?`
	err := e.DB.QueryRow(query, specialistID).Scan(&currentPath, &generation)
	if err != nil {
		return fmt.Errorf("failed to fetch specialist %s: %w", specialistID, err)
	}

	// 2. Read the current persona file
	content, err := os.ReadFile(currentPath)
	if err != nil {
		return fmt.Errorf("failed to read current persona file %s: %w", currentPath, err)
	}

	// 3. Simulate mutation by appending the RCA prompt
	newGeneration := generation + 1
	mutationHeader := fmt.Sprintf("\n\n## Evolution Generation %d\n\n### RCA Insights\n%s\n", newGeneration, rcaPrompt)
	newContent := append(content, []byte(mutationHeader)...)

	// 4. Save as a new versioned file
	ext := filepath.Ext(currentPath)
	base := strings.TrimSuffix(currentPath, ext)
	// Remove previous version suffix if exists (e.g., specialist-v1 -> specialist)
	if strings.Contains(base, "-v") {
		parts := strings.Split(base, "-v")
		base = strings.Join(parts[:len(parts)-1], "-v")
	}
	newPath := fmt.Sprintf("%s-v%d%s", base, newGeneration, ext)

	err = os.WriteFile(newPath, newContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write mutated persona to %s: %w", newPath, err)
	}

	// 5. Update registry in DB
	updateQuery := `
		UPDATE specialist_registry 
		SET current_persona_path = ?, 
		    parent_persona_path = ?, 
		    generation = ? 
		WHERE id = ?`
	_, err = e.DB.Exec(updateQuery, newPath, currentPath, newGeneration, specialistID)
	if err != nil {
		return fmt.Errorf("failed to update registry for %s: %w", specialistID, err)
	}

	return nil
}

// Rollback reverts the specialist to its parent persona version.
func (e *MutationEngine) Rollback(specialistID string) error {
	// 1. Get parent path from DB
	var parentPath string
	var currentGeneration int
	query := `SELECT parent_persona_path, generation FROM specialist_registry WHERE id = ?`
	err := e.DB.QueryRow(query, specialistID).Scan(&parentPath, &currentGeneration)
	if err != nil {
		return fmt.Errorf("failed to fetch parent info for %s: %w", specialistID, err)
	}

	if parentPath == "" {
		return fmt.Errorf("no parent persona found for %s to rollback", specialistID)
	}

	// 2. Update registry to use parent path and decrement generation
	// Note: In a production system, we'd probably want to find the grandparent for the next parent_persona_path
	// but for this implementation, we satisfy the draft logic.
	newGeneration := currentGeneration - 1
	if newGeneration < 0 {
		newGeneration = 0
	}

	updateQuery := `
		UPDATE specialist_registry 
		SET current_persona_path = ?, 
		    parent_persona_path = '', -- Resetting parent for simplicity in this version
		    generation = ? 
		WHERE id = ?`
	_, err = e.DB.Exec(updateQuery, parentPath, newGeneration, specialistID)
	if err != nil {
		return fmt.Errorf("failed to rollback registry for %s: %w", specialistID, err)
	}

	return nil
}
