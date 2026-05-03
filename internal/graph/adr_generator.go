package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/utils"
)

// ADRGenerator gerencia a criação física de Architectural Decision Records
type ADRGenerator struct {
	basePath string
}

func NewADRGenerator() *ADRGenerator {
	return &ADRGenerator{
		basePath: "docs/architecture/adr",
	}
}

// ADRData contém todas as informações necessárias para gerar um registro de decisão
type ADRData struct {
	TaskID              string
	Title               string
	Context             string
	Decision            string
	Consequences        string
	VerificationCommand string
	Status              string // Ex: PROPOSED, ACCEPTED, DRAFT
}

// Generate cria um novo arquivo de ADR baseado nos dados fornecidos
func (g *ADRGenerator) Generate(data ADRData) (string, error) {
	slug := utils.Slugify(data.Title)
	// Limita o slug para não estourar o nome do arquivo
	if len(slug) > 50 {
		slug = slug[:50]
	}

	filename := fmt.Sprintf("ADR-%s-%s.md", data.TaskID, slug)
	fullPath := filepath.Join(g.basePath, filename)

	// Template Smart ADR com Frontmatter Blindado
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
Este ADR é um contrato determinístico. Para ser validado, o comando abaixo deve passar:
`+"```bash"+`
%s
`+"```"+`

## Referências
- Task ID: [%s]
`, safeTaskID, safeTitle, now, safeStatus, data.TaskID, data.Title, data.Context, data.Decision, data.Consequences, data.VerificationCommand, data.TaskID)

	// Garante que o diretório existe
	if err := os.MkdirAll(g.basePath, 0755); err != nil {
		return "", fmt.Errorf("adr: failed to create directory: %w", err)
	}

	// Escreve o arquivo
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("adr: failed to write file %s: %w", filename, err)
	}

	return fullPath, nil
}
