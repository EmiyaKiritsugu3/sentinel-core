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

// Generate cria um novo arquivo de ADR baseado na intenção e ID da tarefa
func (g *ADRGenerator) Generate(taskID string, intent string) (string, error) {
	slug := utils.Slugify(intent)
	// Limita o slug para não estourar o nome do arquivo
	if len(slug) > 50 {
		slug = slug[:50]
	}

	filename := fmt.Sprintf("ADR-%s-%s.md", taskID, slug)
	fullPath := filepath.Join(g.basePath, filename)

	// Template Smart ADR com Frontmatter Blindado
	now := time.Now().Format("2006-01-02")
	
	// Escapando campos para o YAML
	safeTaskID := utils.EscapeYAML(taskID)
	safeIntent := utils.EscapeYAML(intent)

	content := fmt.Sprintf(`---
task_id: "%s"
title: "%s"
date: "%s"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-%s: %s

## Contexto
Esta decisão foi capturada proativamente pelo Sentinel via comando 'instruct'.
Intenção original: %s

## Decisão
[Descreva aqui a abordagem técnica e as ferramentas escolhidas]

## Consequências
- [Ponto Positivo 1]
- [Ponto Negativo 1]

## Referências
- Task ID: [%s]
`, safeTaskID, safeIntent, now, taskID, intent, intent, taskID)

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
