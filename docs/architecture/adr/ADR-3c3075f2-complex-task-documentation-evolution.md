---
task_id: "3c3075f2"
title: "Complex Task: Documentation Evolution"
date: "2026-04-30"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-3c3075f2: Complex Task: Documentation Evolution

## Contexto
Esta decisão foi capturada proativamente pelo Sentinel via comando 'instruct'.
Intenção original: Complex Task: Documentation Evolution

## Decisão
Implementaremos um sistema de orquestração sequencial baseado em um "Ledger de Estado" persistente (SQLite). A ferramenta central desta evolução é a `sentinel:decompose`, que permite ao Agente Chefe (Central Engine) quebrar metas de alto nível em sub-tarefas atômicas. 

Componentes Técnicos:
1. **Tool `sentinel:decompose`**: Expõe uma interface para o LLM definir sub-tarefas com descrições, branches Git dedicados e requisitos de capacidades (`capabilities`).
2. **Persistence Layer**: Sub-tarefas são armazenadas na tabela `sub_tasks` com suporte a `UPSERT` para garantir idempotência em re-despachos.
3. **Sequential Dispatcher**: A `Engine` detecta a intenção de decomposição e invoca o `Dispatcher`, que provisiona worktrees isolados e gerencia o ciclo de vida (Git checkout, execução, commit atômico).

Esta abordagem substitui logs estáticos por um grafo de tarefas vivo e auditável, permitindo que falhas em sub-etapas sejam recuperadas sem perda de progresso global.

## Consequências
- **Positivo**: Rastreabilidade cirúrgica de cada mudança atômica; isolamento de ambiente via Worktrees; conformidade nativa com o Padrão #13 (Atomic Persistence).
- **Negativo**: Maior overhead inicial na gestão de transações SQL e na coordenação de múltiplos diretórios de trabalho (Worktrees).
- **Rollback**: Em caso de falha crítica na decomposição, o sistema permite o reset do estado das sub-tarefas via Ledger sem afetar o repositório principal.

## Referências
- Task ID: [3c3075f2]
