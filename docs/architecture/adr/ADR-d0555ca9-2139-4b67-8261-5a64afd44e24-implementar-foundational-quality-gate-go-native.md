---
task_id: "d0555ca9-2139-4b67-8261-5a64afd44e24"
title: "Implementar Foundational Quality Gate (Go Native)"
date: "2026-05-03"
status: "ACCEPTED"
author: "Sentinel Auto-ADR"
---

# ADR-d0555ca9-2139-4b67-8261-5a64afd44e24: Implementar Foundational Quality Gate (Go Native)

## Contexto

Devido a conflitos de ambiente entre o `golangci-lint` v1.55 e o compilador Go 1.26 (erros de depuração nodwarf5 e dependências de ferramentas), a estratégia de linting foi pivotada para uma abordagem "Native-First". Isso garante que o Gate de Qualidade seja resiliente e independente de binários externos complexos.

## Decisão

Utilizaremos o `go fmt` para garantir a padronização do código e o `go vet` para análise estática básica (shadowing, printf, etc.).

## Consequências

- Positivo: Gate de qualidade extremamente rápido e sem dependências externas.
- Positivo: Alinhamento total com a versão do compilador instalada.
- Negativo: Perda de algumas regras avançadas (linter-specific) que o `golangci-lint` proveria.

## Protocolo de Verificação

Este ADR é um contrato determinístico. Para ser validado, o comando abaixo deve passar:

```bash
bash -c "if [ -z \"\$(gofmt -l .)\" ]; then go vet ./...; else gofmt -l .; exit 1; fi"
```

## Referências

- Task ID: [d0555ca9-2139-4b67-8261-5a64afd44e24]
- Standard: STD-11 (Native-First Governance)
