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
- **Cadeia de Qualidade Redundante**: Nenhuma ferramenta de qualidade (linter, tests, race detector, SonarCloud, CodeRabbit) pega todos os bugs isoladamente. A eficácia emerge da **redundância de perspectivas**: cada camada cobre o ponto cego da anterior. O `_` no `time.Parse` passou por linter (não é retorno de função), testes (silenciado), race detector (não é concorrência) — mas CodeRabbit + SonarCloud detectaram juntos pelo padrão + complexidade.
- **modernc.org/sqlite TIMESTAMP Behavior**: `DEFAULT CURRENT_TIMESTAMP` neste driver retorna RFC3339 (`"2026-05-14T07:26:16Z"`), não o formato clássico do SQLite C (`"YYYY-MM-DD HH:MM:SS"`). Não documentado oficialmente. Qualquer `time.Parse` em colunas TIMESTAMP com default deve usar `time.RFC3339`.

## 🧊 The Icebox (Potential Evolutions)

- **Protocolo Bonsai (KISS Optimization)**: Sistema de poda automática de complexidade e redundância (Backlog de vanguarda).
- **SOLID Governance Module**: Validador semântico de princípios SOLID via análise de grafo AST.
- **Compiled Knowledge Engine**: Sistema de injeção automática de erros passados no prompt.
- ~~**WebSocket Live View**: Servidor em Go para atualizar diagramas no browser em tempo real (Fase 5).~~ **IMPLEMENTED**: `internal/liveview` — Go WebSocket server + React/Cytoscape frontend.
- **Semantic Firewall**: Implementar o fluxo de "Subagente Auditor" (Adversarial Review).
- **GenaiClient Interface Extraction**: Extrair interface de `*genai.Client` para permitir mocking em testes. `Engine.Execute()` está em ~13% de cobertura. Com interface + mock, estimativa de >60% coverage no pipeline AI. Prioridade: próxima sprint para fechar gap do SonarCloud QG (78.8% → 80%+).
- **SonarCloud QG Coverage Target**: Plano free não permite customizar thresholds. `qualitygate.wait=true` removido do workflow. Monitorar via dashboard, não via CI block.
- **Chronicle Protocol Formalization**: O ato de escrever crônicas de sessão como capítulos narrativos (~/Documents/Chronicles/sentinel-core/) ainda não é um protocolo formal. Considerar documentar como PMO se o padrão se consolidar.

---

*Última Auditoria de Gaps: 2026-05-14*
