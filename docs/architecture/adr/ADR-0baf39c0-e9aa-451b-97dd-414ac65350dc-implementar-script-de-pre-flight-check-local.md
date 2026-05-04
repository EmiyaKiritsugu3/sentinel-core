---
task_id: "0baf39c0-e9aa-451b-97dd-414ac65350dc"
title: "Implementar script de Pre-Flight Check local"
date: "2026-05-03"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-0baf39c0-e9aa-451b-97dd-414ac65350dc: Implementar script de Pre-Flight Check local

## Contexto

Capturado via comando 'instruct'.
Intenção: Implementar script de Pre-Flight Check local

## Decisão

Criar e integrar um script shell `scripts/audit-local.sh` que execute uma matriz de conformidade antes dos commits, garantindo verificação contínua e impedindo que código que não compila ou que viola o lint chegue à branch principal.

## Consequências

## Protocolo de Verificação

Este ADR é um contrato determinístico. Para ser validado, o comando abaixo deve passar:

```bash
./scripts/audit-local.sh
```

## Referências

- Task ID: [0baf39c0-e9aa-451b-97dd-414ac65350dc]
