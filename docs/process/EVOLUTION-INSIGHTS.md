# Sentinel Evolution Insights & Icebox [PID-SENTINEL]

Este documento registra as fissuras estruturais e ideias de vanguarda descobertas durante a execução. Ele serve como o backlog de maturação do Arquiteto.

## 🔍 Gaps Estruturais (Technical Debt)
- [x] **String Sanitization Leak**: Centralizado em `pkg/utils/text.go`.
- [x] **Full-Scan Bottleneck**: Implementado Scan Incremental via hashes SHA256.
- [ ] **Task Metadata Anemia**: A struct de Tarefa no `internal/state` carece de contexto sobre o ambiente de execução e logs de erro passados.
- [ ] **Subagent Interface (Bridge)**: A Prompt Factory ainda não expõe uma forma de "receber de volta" os metadados do subagente para auditoria.

## 🧊 The Icebox (Potential Evolutions)
- **Modo Entrevista**: Iniciar o Sentinel em pastas vazias com perguntas de produto para o usuário leigo.
- **WebSocket Live View**: Servidor em Go para atualizar diagramas no browser em tempo real (Fase 5).
- **Atomic Commits**: O comando `sentinel audit` deve executar o git commit automaticamente após o sucesso.
- **Integrated Linter**: O comando `sentinel audit` deve rodar o `golangci-lint` (ou similar) internamente para automatizar as mesmas checagens de vanguarda que o CodeRabbit realiza.

---
*Última Auditoria de Gaps: 2026-04-26*
