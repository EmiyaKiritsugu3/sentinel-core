---
task_id: "ec002ac"
title: "SQLite Migration Strategy"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002ac: SQLite Migration Strategy

## Contexto

O Sentinel utiliza SQLite como banco de dados embarcado para o grafo de dependências, ledger de tarefas e registro de especialistas. O schema evolui com o tempo — novas colunas são adicionadas a tabelas existentes, colunas são renomeadas, novas tabelas são criadas. Migrações devem ser seguras para execução repetida (idempotentes) e não devem quebrar bancos de dados existentes de usuários.

## Decisão

Adotamos uma estratégia de migração com três princípios:

**1. CREATE TABLE IF NOT EXISTS** — Toda definição de tabela no schema base usa `CREATE TABLE IF NOT EXISTS`. Isso garante que a primeira execução cria as tabelas e execuções subsequentes são no-ops. Triggers e índices seguem o mesmo padrão.

**2. columnExistsInTx + PRAGMA table_info** — Migrações incrementais (ex.: adicionar coluna `latency_ms` à tabela `tasks`) usam `columnExistsInTx(ctx, tx, table, column)` que consulta `PRAGMA table_info(table)` para verificar existência da coluna antes de executar `ALTER TABLE ADD COLUMN`. O mapeamento de tabelas permitidas é hardcoded no map `pragmaTableInfo` — prevenindo SQL injection (SonarCloud S2077) ao rejeitar nomes de tabela arbitrários.

**3. Zero breaking changes** — Migrações nunca removem colunas ou tabelas. Renomeações (ex.: `specialist_id → agent_name` em `agent_trust`) são feitas com verificação prévia da existência da coluna antiga. Se a coluna antiga não existe, a migração é skipped.

**4. Transacionalidade** — Toda migração executa dentro de uma transação SQL (`BeginTx`/`Commit`/`Rollback`). Falha em qualquer etapa faz rollback completo, garantindo atomicidade.

**5. KISS seeding** — Dados iniciais (especialistas padrão) usam `INSERT OR IGNORE` para idempotência.

## Consequências

- **Positivo**: Migrações 100% idempotentes — executar `sentinel scan` múltiplas vezes nunca corrompe o schema.
- **Positivo**: Segurança contra SQL injection via whitelist de tabelas no `pragmaTableInfo`.
- **Positivo**: Atomicidade transacional garante que o banco nunca fique em estado intermediário inválido.
- **Negativo**: `PRAGMA table_info` requer lock de leitura no banco SQLite durante a migração. Para bancos muito grandes (>1GB), a latência pode ser perceptível. Irrelevante para o caso de uso atual.

## Referências

- Task ID: [ec002ac]
- Implementação: `internal/graph/schema.go:199-276`
