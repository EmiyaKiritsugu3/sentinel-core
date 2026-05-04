---
task_id: ad9933bf
date: 2026-04-28
status: PROPOSED
author: Sentinel Auto-ADR
---

# ADR-ad9933bf: Refatorar camada de persistencia

## Contexto

Esta decisão foi capturada proativamente pelo Sentinel via comando 'instruct'.
Intenção original: Refatorar camada de persistencia

## Decisão

Migraremos a camada de persistência de arquivos JSON/Markdown isolados para um sistema de Ledger centralizado baseado em SQLite (`.sentinel/graph.db`). Esta mudança é fundamental para suportar fluxos de trabalho agenticos complexos, permitindo consultas relacionais eficientes sobre o grafo de dependências, sub-tarefas e histórico de auditoria, garantindo atomicidade e integridade referencial entre diferentes módulos do Sentinel.

## Consequências

- **Positivo**: Centralização do estado, suporte a transações ACID e maior velocidade em buscas complexas de impacto arquitetural.
- **Negativo**: Introduz a necessidade de gerenciar migrações de esquema SQL e aumenta a pegada de infraestrutura local do projeto.

## Referências

- Task ID: [ad9933bf]
