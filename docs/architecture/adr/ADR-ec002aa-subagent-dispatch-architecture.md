---
task_id: "ec002aa"
title: "Sub-agent Dispatch Architecture"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002aa: Sub-agent Dispatch Architecture

## Contexto

O Sentinel opera no modelo de Tríade Soberana: um Warden (Chief Engineer) decompõe tarefas em sub-tarefas e as despacha para Operators (subagentes especialistas). Cada sub-agente deve executar em isolamento completo de workspace, com seu próprio branch git, e reportar resultados de volta ao ledger central. O dispatcher é o ponto único de serialização de writes no banco.

## Decisão

`Dispatcher` atua como **Write Serializer** — único componente autorizado a escrever na tabela `sub_tasks` e criar worktrees. Estrutura:

```go
type Dispatcher struct {
    Registry *RegistryManager  // seleção de especialista
    Shield   *GitShield        // isolamento de worktree
    DB       *sqlite.DB        // ledger central
}
```

**Fluxo de dispatch** (`Dispatch`):
1. `Registry.SelectBest(ctx, caps)` — seleciona especialista com maior `reliability_score` que atenda todas as capacidades requeridas.
2. `Shield.CreateWorktree(taskID, branch)` — cria worktree git isolado em `.worktrees/sentinel-task-{slug}` com validação de path via `Validator.ValidatePath`.
3. Persiste `SubTask` no ledger via `INSERT ... ON CONFLICT DO UPDATE` com rollback do worktree em caso de falha (best-effort cleanup).
4. `SubTask` carrega `ID`, `ParentTaskID`, `SpecialistID`, `Description`, `Status`, `WorktreePath`, `BranchName`, `RequiredCapabilities` (JSON).

**Worktree isolation**: Cada sub-agente opera em um git worktree dedicado, isolado do workspace principal. O `GitShield` gerencia ciclo de vida: criação, remoção individual e cleanup massivo (`CleanupWorktrees`).

**Reconciliação**: `ReconcileEvents` lê arquivos JSON do diretório `.sentinel/events/`, atualiza status no ledger e remove arquivos processados. Usa `bufio.NewReader` para leitura eficiente (Standard #01).

## Consequências

- **Positivo**: Write Serializer elimina race conditions no ledger — apenas o Dispatcher escreve em `sub_tasks`.
- **Positivo**: Worktree isolation previne conflitos de arquivo entre sub-agentes concorrentes.
- **Positivo**: KISS — processamento sequencial de sub-tarefas pendentes (`SELECT ... WHERE status = 'PENDING'`), sem orquestrador complexo.
- **Negativo**: Modelo sequencial não escala para paralelismo real de sub-agentes. Para execução paralela, seria necessário pool de Engines com coordenação via canal de resultados.

## Referências

- Task ID: [ec002aa]
- Implementação: `internal/agents/dispatcher.go`, `internal/agents/git_shield.go`
