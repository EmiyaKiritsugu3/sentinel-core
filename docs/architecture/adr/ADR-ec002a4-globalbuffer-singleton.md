---
task_id: "ec002a4"
title: "GlobalBuffer Singleton Pattern"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a4: GlobalBuffer Singleton Pattern

## Contexto

Múltiplos subsistemas do Sentinel — engine, ferramentas, dispatcher, classificador — precisam registrar eventos de sessão em um buffer centralizado. A alternativa de injeção de dependência (passar `*EventBuffer` via construtor para cada componente) criaria acoplamento excessivo e poluição de assinaturas, especialmente em camadas profundas como `processResponse`, `executePhase` e `checkGateA`.

## Decisão

Adotamos o padrão **singleton com variável global exportada**: `var GlobalBuffer = NewEventBuffer(1000)` no pacote `knowledge`. Todos os subsistemas referenciam diretamente `knowledge.GlobalBuffer.Record(...)`. O buffer é inicializado estaticamente no `init` do pacote (via declaração `var`), garantindo disponibilidade antes de qualquer chamada.

Trade-offs avaliados vs DI:
- **Simplicidade**: Singleton elimina a necessidade de passar `*EventBuffer` por 7 camadas de construtores (`Engine → processResponse → checkGateA → ...`).
- **Testabilidade**: Testes isolados criam `NewEventBuffer(N)` local, sem dependência do global. O singleton é testado indiretamente via testes de integração (`TestDebriefService_SaveToGraph_WithRealDB`).
- **Acoplamento**: O singleton introduz estado global mutável, mas o buffer é essencialmente um coletor de telemetria — não afeta lógica de negócio e opera em modo append-only com leitores concorrentes.

## Consequências

- **Positivo**: API minimalista — qualquer componente pode registrar eventos com uma linha (`knowledge.GlobalBuffer.Record(...)`).
- **Positivo**: Inicialização garantida na ordem correta via inicialização de pacote Go.
- **Negativo**: Estado global mutável dificulta testes paralelos que dependem do buffer compartilhado. Mitigação: `TestMain` pode resetar o global, e testes usam buffers locais sempre que possível.
- **Negativo**: Dificulta substituição por implementação alternativa (ex.: buffer com persistência imediata). Para esse caso, seria necessário refatorar para interface com DI.

## Referências

- Task ID: [ec002a4]
- Implementação: `internal/knowledge/buffer.go:136`
