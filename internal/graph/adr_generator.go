// Package graph provides AST analysis, dependency resolution, and visualization.
package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/utils"
)

// ADRGenerator manages the physical creation of Architectural Decision Records
type ADRGenerator struct {
	basePath string
}

// NewADRGenerator creates a new ADR generator instance.
func NewADRGenerator() *ADRGenerator {
	return &ADRGenerator{
		basePath: "docs/architecture/adr",
	}
}

// ADRData contains all information needed to generate a decision record
type ADRData struct {
	TaskID              string
	Title               string
	Context             string
	Decision            string
	Consequences        string
	VerificationCommand string
	Status              string // Ex: PROPOSED, ACCEPTED, DRAFT
}

// Generate creates a new ADR file based on the provided data
func (g *ADRGenerator) Generate(data ADRData) (string, error) {
	slug := utils.Slugify(data.Title)
	// Limits the slug to avoid overflowing the filename
	if len(slug) > 50 {
		slug = slug[:50]
	}

	filename := fmt.Sprintf("ADR-%s-%s.md", data.TaskID, slug)
	fullPath := filepath.Join(g.basePath, filename)

	// Smart ADR Template with Hardened Frontmatter
	now := time.Now().Format("2006-01-02")
	status := data.Status
	if status == "" {
		status = "PROPOSED"
	}
	safeStatus := utils.EscapeYAML(status)

	// Escapando campos para o YAML
	safeTaskID := utils.EscapeYAML(data.TaskID)
	safeTitle := utils.EscapeYAML(data.Title)

	content := fmt.Sprintf(`---
task_id: "%s"
title: "%s"
date: "%s"
status: "%s"
author: "Sentinel Auto-ADR"
---

# ADR-%s: %s

## Contexto
%s

## Decisão
%s

## Consequências
%s

## Protocolo de Verificação
Este ADR é um contrato determinístico. Para ser validado, o commando abaixo deve passar:
`+"```bash"+`
%s
`+"```"+`

## Referências
- Task ID: [%s]
`, safeTaskID, safeTitle, now, safeStatus, data.TaskID, data.Title, data.Context, data.Decision, data.Consequences, data.VerificationCommand, data.TaskID)

	// Ensures the directory exists
	if err := os.MkdirAll(g.basePath, 0750); err != nil {
		return "", fmt.Errorf("adr: failed to create base directory %s: %w", g.basePath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("adr: failed to write file %s: %w", filename, err)
	}

	return fullPath, nil
}
