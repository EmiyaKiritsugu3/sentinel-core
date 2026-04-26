# Sentinel Evolution Insights & Icebox [PID-SENTINEL]

Este documento registra as fissuras estruturais e ideias de vanguarda descobertas durante a execução. Ele serve como o backlog de maturação do Arquiteto.

## 🔍 Gaps Estruturais (Technical Debt)
- [x] **String Sanitization Leak**: Centralizado em `pkg/utils/text.go`.
- [x] **Full-Scan Bottleneck**: Implementado Scan Incremental via hashes SHA256.
- [x] **Context Continuity**: Implementado o **Sovereign Handover Protocol** para trocas de sessão.
- [x] **Evidence Enforcement**: Institucionalizado o **Standard #13 (Verify, Never Assume)**.
- [ ] **Task Metadata Anemia**: A struct de Tarefa no `internal/state` carece de contexto sobre o ambiente de execução e logs de erro passados.
- [ ] **Dependency Injection**: Comandos CLI utilizam a global `DBInstance`. Refatorar para injeção de dependência via construtores para viabilizar testes unitários.
- [ ] **Cascading Pruning**: O scan atual não limpa arestas de arquivos que foram deletados (apenas modificados).

## 🧊 The Icebox (Potential Evolutions)
- **Modo Entrevista**: Iniciar o Sentinel em pastas vazias com perguntas de produto para o usuário leigo.
- **WebSocket Live View**: Servidor em Go para atualizar diagramas no browser em tempo real (Fase 5).
- **Atomic Commits**: O comando `sentinel audit` deve executar o git commit automaticamente após o sucesso.
- **Integrated Linter**: O comando `sentinel audit` deve rodar o `golangci-lint` internamente para automatizar as checagens do CodeRabbit.
- **Semantic Firewall**: Implementar o fluxo de "Subagente Auditor" (Adversarial Review) para validar novas epifanias antes de selá-las.
- **Confidence Lifecycle**: Implementar estados (PROPOSED -> AUDITED -> SEALED) no banco de dados para cada Engineering Standard.
- **Weighted Health Score**: Ponderar a nota de saúde do projeto pelo Tier da tarefa (falhar em T3 impacta mais que em T1).

---
*Última Auditoria de Gaps: 2026-04-26*
