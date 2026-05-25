---
task_id: "ec002a5"
title: "RWMutex Concurrency Model in EventBuffer"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a5: RWMutex Concurrency Model in EventBuffer

## Contexto

`EventBuffer` serve múltiplas goroutines simultaneamente: o engine grava eventos via `Record()` durante a execução do agente, enquanto o `DebriefService` lê o buffer inteiro via `Snapshot()` e filtros (`ByDomain`, `ByType`). O padrão de acesso é read-heavy — snapshots são tirados apenas ao final da sessão ou sob demanda, enquanto gravações ocorrem continuamente.

## Decisão

Utilizamos `sync.RWMutex` (não `sync.Mutex`) para maximizar throughput de leitura. A distinção Lock vs RLock segue o princípio:

- **Lock (write)**: `Record()` adquire write lock exclusivo, pois modifica `head`, `size` e o conteúdo do slice. A seção crítica é mínima: timestamp default, clone do evento, avanço circular do ponteiro.
- **RLock (read)**: `Snapshot()`, `Len()`, `filter()` e todos os métodos de consulta (`ByDomain`, `ByType`, `Patterns`, `Decisions`, `Errors`) adquirem read lock compartilhado. Múltiplas leituras concorrentes são permitidas sem bloqueio mútuo.

Snapshot consistency: `Snapshot()` captura o estado do buffer sob RLock, calcula o índice inicial `(head-size+max)%max`, e itera `size` elementos construindo um slice independente. O `cloneEvent` garante que dados internos (slice `Tags`) não sejam compartilhados com o chamador. O `sort.Slice` final por `Timestamp` ocorre sobre o slice clonado, sem locks adicionais.

## Consequências

- **Positivo**: Leituras concorrentes sem contenção — múltiplas chamadas a `ByDomain`, `Patterns`, etc. executam em paralelo.
- **Positivo**: Snapshots atomicamente consistentes — o RLock garante que `head` e `size` não mudem durante a cópia.
- **Negativo**: RWMutex tem overhead ligeiramente maior que Mutex simples (~40ns vs ~20ns por operação). Irrelevante para a carga de trabalho (<1000 eventos/sessão).
- **Negativo**: Em cenário de write-heavy extremo (>10k writes/s), RWMutex pode sofrer de writer starvation. Não aplicável ao caso de uso atual.

## Referências

- Task ID: [ec002a5]
- Implementação: `internal/knowledge/buffer.go:46-53`
