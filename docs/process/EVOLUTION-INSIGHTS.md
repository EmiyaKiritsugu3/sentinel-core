# Sentinel Evolution Insights & Icebox [PID-SENTINEL]

Este documento registra as fissuras estruturais e ideias de vanguarda descobertas durante a execução. Ele serve como o backlog de maturação do Arquiteto.

## 🔍 Gaps Estruturais (Technical Debt)
- [x] **String Sanitization Leak**: Centralizado em `pkg/utils/text.go`.
- [x] **Full-Scan Bottleneck**: Implementado Scan Incremental via hashes SHA256.
- [x] **Context Continuity**: Implementado o **Sovereign Handover Protocol**.
- [x] **Evidence Enforcement**: Institucionalizado o **Standard #13 (Verify, Never Assume)**.
- [x] **Governance Balance**: Implementado o filtro de proporcionalidade (Standard #14).
- [ ] **Task Metadata Anemia**: A struct de Tarefa no `internal/state` carece de contexto sobre o ambiente de execução e logs de erro passados.
- [ ] **Dependency Injection**: Refatorado para injeção via construtores no CLI.
- [ ] **Cascading Pruning**: O scan atual não limpa arestas de arquivos deletados.
- [ ] **Manual Focus**: O comando `visualize` precisa da flag `--focus`.

## 🧊 The Icebox (Potential Evolutions)
- **Compiled Knowledge Engine**: Sistema de injeção automática de erros passados no prompt.
- **Modo Entrevista**: Iniciar o Sentinel em pastas vazias com perguntas de produto.
- **WebSocket Live View**: Servidor em Go para atualizar diagramas no browser em tempo real (Fase 5).
- **Atomic Commits**: O comando `sentinel audit` deve executar o git commit automaticamente.
- **Integrated Linter**: O comando `sentinel audit` deve rodar o `golangci-lint` internamente.
- **Semantic Firewall**: Implementar o fluxo de "Subagente Auditor" (Adversarial Review).
- **Confidence Lifecycle**: Implementar estados (PROPOSED -> AUDITED -> SEALED) no banco de dados.
- **Weighted Health Score**: Ponderar a nota de saúde pelo Tier da tarefa.

---
*Última Auditoria de Gaps: 2026-04-26*
