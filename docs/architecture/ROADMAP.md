# Sentinel Sovereign Roadmap [PID-SENTINEL]

Este documento define a trajetória oficial de desenvolvimento do Sentinel Core. Nenhuma tarefa deve ser iniciada sem estar mapeada neste roadmap.

## 🏁 Milestones Alcançados
- [x] **Fase 1: The Fail-Safe Foundation**
  - Implementação de Timeouts de Auditoria.
  - Governança de Erros com Wrapping.
  - Definição da Tríade (Warden, Chief Engineer, Operator).
- [x] **Fase 2: The Context Engine**
  - Scanner de AST (Go) paralelo via Worker Pool.
  - Persistência em SQLite (CGO-free).
  - Extração cirúrgica de linhas de código real.
- [x] **Fase 2.10: Sovereign Hardening**
  - Refatoração para Injeção de Dependência (Fim das Globais).
  - Blindagem de Segurança (shlex, Foreign Keys, Transactions).
  - Implementação do Sovereign Validator (Hard Gates).

## 🚀 Próximas Frentes (O Plano Concreto)

### Fase 3: The Language Expansion (AST Evolution)
O Sentinel deve ser capaz de gerir projetos Web de vanguarda.
- [x] Abstração da Engine Multi-Linguagem (Orchestrator).
- [ ] Integração com **Tree-sitter** (C-bindings ou Pure Go).
- [ ] Scanner AST para **TypeScript/TSX**.
- [ ] Mapeamento de dependências entre componentes React.
*Critério de Sucesso: `sentinel scan` em um projeto Next.js povoa o SQLite com sucesso.*

### Fase 4: The Agentic State Machine (Proactive Governance)
Transformar o Sentinel em um guia proativo para o usuário.
- [x] Saneamento de Grafo via .gitignore (Hybrid Filter).
- [x] Modo Entrevista (Comando `instruct` blindado para CI/CD).
- [x] **Auto-ADR**: Gera o rascunho do ADR baseado na conversa inicial.
- [x] **Dashboard Visibility**: Vincula fisicamente tarefas aos ADRs no relatório.
- [ ] **Subagent Dispatcher**: Ferramenta nativa para o Chief Engineer invocar e monitorar Operadores.
*Critério de Sucesso: Criação de uma feature completa apenas via diálogo, sem intervenção manual no plano.*

### Fase 5: The Visual Sovereign (Live UI)
A visualização de arquitetura deve ser interativa.
- [ ] **Sentinel Live View**: Servidor WebSocket em Go que envia o Grafo para uma UI Web.
- [ ] **Interactive C4**: Clique no nó do diagrama para abrir o código ou ver o ADR relacionado.
*Critério de Sucesso: Visualização em tempo real no browser enquanto o código muda.*

---
*Atualizado em: 2026-04-26*
*Assinado: Sovereign Council*
