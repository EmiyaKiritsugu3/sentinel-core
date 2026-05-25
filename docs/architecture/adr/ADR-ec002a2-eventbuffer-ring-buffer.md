---
task_id: "ec002a2"
title: "EventBuffer Ring Buffer Design"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a2: EventBuffer Ring Buffer Design

## Contexto

Durante a execução de um agente Sentinel, centenas de eventos (`SessionEvent`) são gerados: decisões, erros, padrões detectados, mudanças de arquivo e comandos executados. Esses eventos precisam ser coletados em memória com performance O(1) para posterior geração de debrief e persistência em SQLite. O buffer deve ser thread-safe para suportar múltiplos emissores concorrentes (engine, tools, dispatcher).

## Decisão

Implementamos `EventBuffer` como um **ring buffer circular** com capacidade padrão de **1000 eventos**, proteção de concorrência via `sync.RWMutex`, e operações de `Record` (O(1)) e `Snapshot` (O(n) com ordenação cronológica).

A estrutura mantém um slice pré-alocado `events []SessionEvent` de tamanho fixo, um ponteiro `head` para a próxima posição de escrita, e um contador `size` para o número atual de elementos. Quando o buffer atinge a capacidade máxima, o evento mais antigo é sobrescrito silenciosamente — comportamento aceitável para um buffer de telemetria onde eventos muito antigos perdem relevância.

O método `Record` adquire `Lock()` (write lock), registra timestamp automático se ausente, clona o evento (deep copy de `Tags` via `cloneEvent`), e avança `head` circularmente. O método `Snapshot` adquire `RLock()` (read lock), reconstrói a ordem cronológica percorrendo do elemento mais antigo (`(head-size+max)%max`) até o mais recente, e aplica `sort.Slice` estável por `Timestamp` para garantir ordenação determinística mesmo com timestamps iguais.

Métodos de filtro (`ByDomain`, `ByType`, `Patterns`, `Decisions`, `Errors`) usam `RLock()` e predicado funcional, retornando slices novos sem referências ao buffer interno.

## Consequências

- **Positivo**: Operação de escrita O(1) sem alocação de memória após inicialização, ideal para hot path do engine.
- **Positivo**: Thread-safe por design — múltiplas goroutines podem gravar e ler concorrentemente.
- **Negativo**: Buffer circular descarta eventos antigos sem notificação. Para auditoria completa, o consumidor deve drenar o buffer periodicamente via `DebriefService.Save`.
- **Negativo**: `Snapshot` aloca slice completo e realiza sort — O(n log n). Para 1000 eventos, overhead é insignificante (<1ms).

## Referências

- Task ID: [ec002a2]
- Implementação: `internal/knowledge/buffer.go`
