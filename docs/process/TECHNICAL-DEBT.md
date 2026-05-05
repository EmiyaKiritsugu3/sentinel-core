# Technical Debt Log [PID-SENTINEL]

## [2026-05-03] State Machine & Orchestration (Filtro B)

- **Sub-task Depth Explosion**: O `DecomposeTool` limita a largura (5 sub-tarefas), mas o `Dispatcher` não possui um limite de profundidade (Task -> Subtask -> Sub-subtask). Risco teórico de recursão exponencial em árvores de decisão complexas.
- **Verification Command Sandbox**: Atualmente o comando de verificação do ADR roda com o mesmo privilégio do Sentinel. Falta isolamento em container ou sandbox para comandos de prova de origem externa.

## [2026-04-28] Dashboard & Discovery (Filtro A)

- **Scalability of Glob**: O uso de `filepath.Glob` para descobrir ADRs no Aggregator é $O(N)$ sobre o sistema de arquivos. Para projetos massivos, isso se tornará um gargalo de I/O.
- **Dashboard Growth**: `COMPLIANCE-DASHBOARD.md` crescerá linearmente. Falta suporte para arquivamento ou paginação.
- **Missing CLI Metadata**: O relatório CLI não exibe o `created_at`, dificultando a análise cronológica.

## [2026-05-04] WebSocket & Dependency Security (Filtro A)

- **liveview Graceful Shutdown**: `liveview.Server.Run` retorna quando o contexto é cancelado, mas não fecha as conexões `wsClient` abertas. Os goroutines `readPump`/`writePump` de clientes ativos sobrevivem até o TCP timeout. Impacto: aceitável para CLI dev tool; crítico se o servidor evoluir para produção. Fix: iterar `s.clients` no `ctx.Done()` e fechar todos os canais `send`.
- **Indirect Dependency CVEs**: `grpc` e `oauth2` são dependências indiretas (via `google/generative-ai-go`) que tiveram CVEs críticos sem update automático. Mitigação: rodar `go list -m -u all | grep available` periodicamente ou adicionar Dependabot ao workflow de CI para monitorar upgrades automáticos de dependências indiretas.

---
*Assinado: Security Auditor & Senior Architect*
