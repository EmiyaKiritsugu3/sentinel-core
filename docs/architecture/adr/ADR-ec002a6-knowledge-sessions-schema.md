---
task_id: "ec002a6"
title: "knowledge_sessions + session_events Schema"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a6: knowledge_sessions + session_events Schema

## Contexto

O debrief de cada sessão de agente precisa ser persistido em SQLite para consultas históricas, análise de tendências e geração de relatórios. Os dados incluem metadados da sessão (timestamp, contagem de eventos por tipo, domínios) e os eventos individuais (tipo, domínio, sumário, detalhes, arquivo afetado, tags). A modelagem relacional deve suportar consultas por sessão, por domínio e por tipo de evento.

## Decisão

Modelamos com **duas tabelas** relacionadas via chave estrangeira com `ON DELETE CASCADE`:

**`knowledge_sessions`** (tabela pai):
- `id TEXT PRIMARY KEY` — identificador curto da sessão (8 caracteres do UUID)
- `markdown_path TEXT NOT NULL` — caminho do arquivo .md gerado
- `started_at`, `ended_at TIMESTAMP` — intervalo da sessão
- `event_count`, `decision_count`, `error_count`, `pattern_count INTEGER` — contadores pré-agregados para consultas rápidas (evitam `COUNT(*)` nas queries de listagem)
- `domains TEXT NOT NULL DEFAULT ''` — lista de domínios separados por vírgula (denormalização intencional para queries simples de filtro)

**`session_events`** (tabela filha):
- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `session_id TEXT NOT NULL` com `FOREIGN KEY REFERENCES knowledge_sessions(id) ON DELETE CASCADE`
- `event_type`, `domain`, `summary`, `detail`, `file_path`, `tags TEXT DEFAULT ''`
- `timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP`
- Índice `idx_session_events_session_id` para lookup por sessão

A denormalização de `domains` na tabela `knowledge_sessions` é intencional: permite filtrar sessões por domínio com uma query simples (`WHERE domains LIKE '%auth%'`) sem JOIN, ao custo de redundância controlada. Os contadores agregados (`decision_count`, `error_count`, `pattern_count`) eliminam a necessidade de subqueries para listagens de dashboard.

## Consequências

- **Positivo**: `ON DELETE CASCADE` garante que remover uma sessão limpa automaticamente todos os seus eventos.
- **Positivo**: Contadores pré-agregados permitem listagem de sessões com métricas em O(1).
- **Negativo**: Denormalização de `domains` como string CSV impede queries com índice sobre domínios individuais. Para catálogo de domínios com milhões de sessões, seria necessária tabela de junção `session_domains`.

## Referências

- Task ID: [ec002a6]
- Implementação: `internal/graph/schema.go:168-196`, `internal/knowledge/debrief.go:198-218`
