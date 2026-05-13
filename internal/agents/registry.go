package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// RegistryManager handles persistent specialist selection and lifecycle.
type RegistryManager struct {
	db *sqlite.DB
}

// NewRegistryManager initializes a manager with a database handle.
func NewRegistryManager(db *sqlite.DB) (*RegistryManager, error) {
	if err := sqlite.ValidateDB(db, "registry-manager"); err != nil {
		return nil, err
	}
	return &RegistryManager{db: db}, nil
}

// SelectBest finds the specialist with the highest reliability score that matches ALL requested capabilities.
func (m *RegistryManager) SelectBest(ctx context.Context, caps []string) (*Specialist, error) {
	// Standard #05: Error governance - wrapping errors with context
	rows, err := m.db.Conn.QueryContext(ctx, "SELECT id, name, base_persona, current_persona_path, reliability_score, capabilities FROM specialist_registry ORDER BY reliability_score DESC")
	if err != nil {
		return nil, fmt.Errorf("registry: failed to query specialists: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var s Specialist
		var capsJSON string
		if err := rows.Scan(&s.ID, &s.Name, &s.BasePersona, &s.CurrentPersonaPath, &s.ReliabilityScore, &capsJSON); err != nil {
			return nil, fmt.Errorf("registry: failed to scan specialist: %w", err)
		}

		specialistCaps, err := unmarshalCapabilities(capsJSON)
		if err != nil {
			continue
		}
		s.Capabilities = specialistCaps

		if m.matchesAll(s.Capabilities, caps) {
			return &s, nil
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("registry: row iteration error: %w", err)
	}

	return nil, fmt.Errorf("registry: no specialist found with capabilities: %s", strings.Join(caps, ", "))
}

func (m *RegistryManager) matchesAll(specialistCaps []string, requestedCaps []string) bool {
	capMap := make(map[string]bool)
	for _, c := range specialistCaps {
		capMap[c] = true
	}

	for _, req := range requestedCaps {
		if !capMap[req] {
			return false
		}
	}
	return true
}
