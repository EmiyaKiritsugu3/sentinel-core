---
task_id: "306b0ea4-2d45-445c-b098-8efca7a98745"
title: "Implementar markdownlint para padronização de documentos"
date: "2026-05-03"
status: "ACCEPTED"
author: "Sentinel Auto-ADR"
---

# ADR-306b0ea4-2d45-445c-b098-8efca7a98745: Implementar markdownlint para padronização de documentos

## Contexto

Para garantir que a documentação (ADRs, Wikies, Logs) siga um padrão rigoroso de formatação e legibilidade, é necessária a implementação de um linter de Markdown.

## Decisão

Utilizaremos o `markdownlint-cli2` configurado via `.markdownlint-cli2.yaml` na raiz do projeto. O gate será executado via `npx` para evitar a necessidade de instalação global.

## Consequências

- Positivo: Documentação consistente e profissional.
- Positivo: Erros de formatação detectados antes do commit.
- Negativo: Requer ambiente Node.js/NPM para execução do gate.

## Protocolo de Verificação

Este ADR é um contrato determinístico. Para ser validado, o comando abaixo deve passar:

```bash
npx --yes markdownlint-cli2
```

## Referências

- Task ID: [306b0ea4-2d45-445c-b098-8efca7a98745]
- Config: .markdownlint-cli2.yaml
