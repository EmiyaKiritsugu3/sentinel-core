# Sentinel Evolution Insights & Icebox [PID-SENTINEL]

Este documento registra as fissuras estruturais e ideias de vanguarda descobertas durante a execução. Ele serve como o backlog de maturação do Arquiteto.

## 🔍 Gaps Estruturais (Technical Debt & Traumas)

- [x] **CLI Input Blocking (Trauma)**: Bug nas versões v0.40.x da Gemini CLI em Linux. Enter falha em prompts interativos. **Solução**: Usar `Ctrl + J`.
- [x] **String Sanitization Leak**: Centralizado em `pkg/utils/text.go`.
- [x] **Full-Scan Bottleneck**: Implementado Scan Incremental via hashes SHA256.
- [x] **Context Continuity**: Implementado o **Sovereign Handover Protocol**.
- [ ] **Task Metadata Anemia**: A struct de Tarefa no `internal/state` carece de contexto sobre o ambiente de execução e logs de erro passados.
- [ ] **Cascading Pruning**: O scan atual não limpa arestas de arquivos deletados.

## 💡 Cognitive Patterns (Heuristics)

- **Threshold-Sensitive Testing**: Quando testar classificadores baseados em pesos (ex: `VaguenessScore`), descrições curtas e ambíguas podem falhar no threshold se não incluírem sinais múltiplos (ex: Verb + Pronoun). Testes devem ser desenhados para atingir picos de sinal claros (> 0.60) para evitar flakiness em mudanças de pesos.

## 🧊 The Icebox (Potential Evolutions)

- **Protocolo Bonsai (KISS Optimization)**: Sistema de poda automática de complexidade e redundância (Backlog de vanguarda).
- **SOLID Governance Module**: Validador semântico de princípios SOLID via análise de grafo AST.
- **Compiled Knowledge Engine**: Sistema de injeção automática de erros passados no prompt.
- **WebSocket Live View**: Servidor em Go para atualizar diagramas no browser em tempo real (Fase 5).
- **Semantic Firewall**: Implementar o fluxo de "Subagente Auditor" (Adversarial Review).

---
*Última Auditoria de Gaps: 2026-04-29*
