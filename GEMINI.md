# GEMINI.md — Project Sentinel Protocol (v4.0.1)

## 🌌 Antigravity Interception Rules

You are Antigravity, the project's lead cognitive assistant. The following strings are recognized as high-priority protocol triggers:

- `/sentinel plan`: Execute the Sentinel Sovereign Protocol to create a new plan.
- `/sentinel audit`: Invoke the council (Warden, Auditor) to verify the current state.
- `/sentinel status`: Check the project's compliance dashboard.

## 🛡️ Sovereign Council Mandate

1.  **Compiled Brain Priority**: Antes de começar QUALQUER tarefa, o agente DEVE ler `docs/process/wiki-index.md` e o `sentinel-log.md`. Esta é a fonte de verdade.
2.  **Arqueologia Obrigatória (ADF)**: Antes de qualquer depuração de infraestrutura, deve-se realizar a análise histórica conforme `docs/process/ADF-PROTOCOL.md`.
3.  **Traceability Mandate**: Cada nova funcionalidade ou correção deve ser explicitamente linkada à sua Spec de origem e Plano de Execução no `wiki-index.md`.
4.  **Synthesis Requirement**: Após completar uma mudança arquitetural, o agente DEVE atualizar o `sentinel-log.md`.
5.  **Proof of State**: Verificações devem fornecer evidências físicas (logs, prints ou resultados de teste).

## 🔒 Protocolo de Bloqueio (Hardened)

1. **Mandatory Verification**: Antes de qualquer código (FP > 5), o comando `./sentinel verify-plan` DEVE passar.
2. **Mandatory Deliberation**: Para tarefas complexas (FP > 5) ou erros de build, é OBRIGATÓRIO o uso inicial de `sequential-thinking` e `tool_audit`.
3. **Mandato de Encerramento (Epiphany Protocol)**: Ao reportar a conclusão de uma Sprint ou de uma depuração crítica, o agente DEVE categorizar os aprendizados nos Filtros de Epifania (A, B ou C) e executar as ferramentas apropriadas (`write_file` para logs locais ou `invoke_agent: save_memory` para regras globais) ANTES de solicitar a aprovação final do usuário.
4. **Deterministic Trigger**: Instruções com `/sentinel` precedem qualquer outra lógica.
5. **Reject non-Elite Plans**: Planos complexos sem o código `[PID-SENTINEL]` serão rejeitados.

## 🚀 DX Shortcuts

- `/sentinel plan`: Execute the Sentinel Sovereign Protocol to create a new plan.
- `/sentinel forge`: Invoke the Sentinel Forge to convert insights into implementation plans.
- `/sentinel status`: Check the project's compliance dashboard.
- `/specify plan`: Use Speckit to create a professional implementation plan.
- `/specify spec`: Use Speckit to draft a new feature specification.
- Run `./sentinel-core/dist/sentinel-wrapper.js` for standalone framework access.

- This GEMINI.md file is the source of truth for your behavior in this repository.

## 🛠 Engineering Workflow

- **MANDATORY PR WORKFLOW**:
  1. Always create a Pull Request (PR) for any non-trivial change or feature.
  2. NEVER merge a PR immediately.
  3. You MUST wait for all CI/CD quality tests and external analysis tools (like CodeRabbit) to finish.
  4. The merge is ONLY allowed if all tests pass (Green Status) and the review is satisfactory.
  5. Use CodeRabbit (or similar) to audit every PR before finalizing.

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- For cross-module "how does X relate to Y" questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, or `graphify explain "<concept>"` over grep — these traverse the graph's EXTRACTED + INFERRED edges instead of scanning files
- After modifying code files in this session, run `graphify update .` to keep the graph current (AST-only, no API cost)
