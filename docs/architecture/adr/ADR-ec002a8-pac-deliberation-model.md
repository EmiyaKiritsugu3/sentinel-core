---
task_id: "ec002a8"
title: "PAC Deliberation Decision Model"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a8: PAC Deliberation Decision Model

## Contexto

Agentes de IA operando em modo autônomo podem divergir para estratégias ineficientes, over-engineering ou loops de falha. É necessário um mecanismo de deliberação multi-ângulo que avalie a trajetória do agente e decida entre prosseguir, simplificar, pivotar ou escalar. A decisão deve ser determinística e baseada em métricas objetivas (tokens, passos, falhas, divergência), não em heurísticas de LLM.

## Decisão

Implementamos o modelo **PAC (Proceed/Abort/Change)** com três ângulos de deliberação e semântica de **worst-case wins**:

- **Ângulo A — Minimalist (YAGNI)**: Verifica over-engineering. Dispara `Simplify` se `thoughtTokens/actionTokens > 2.0` (pensando demais, fazendo de menos) ou se >70% do budget de passos foi consumido.
- **Ângulo B — Structuralist (Plan Pivot)**: Verifica se a abordagem técnica atual é fundamentalmente errada. Dispara `Pivot` se `DivergenceCount >= 2` (duas divergências consecutivas de Lyapunov) ou `FailureCount >= 2`.
- **Ângulo C — Auditor (Compliance)**: Verifica restrições de recursos e escalabilidade. Dispara `Escalate` se o modelo já é Pro e ainda está falhando, ou se >90% do budget foi consumido.

A decisão final (`PACResult.Final`) usa `pacWorstCase(a, b, c)` que retorna a recomendação mais severa: `Escalate > Pivot > Simplify > Proceed`. Cada ângulo retorna um valor do enum `PACRecommendation`, e a prioridade é dada pelo valor numérico do iota.

A deliberação é executada via `runPACDeliberation()` no engine, registrando o resultado como `SessionEvent` com tags `["pac", "deliberation"]` e aplicando a ação correspondente (escalar modelo, mudar estratégia, ou prosseguir).

## Consequências

- **Positivo**: Decisão determinística baseada em métricas objetivas — mesmo input produz mesmo output, auditável e testável.
- **Positivo**: Semântica worst-case-wins é conservadora e segura — o sistema prefere escalar ou pivotar a continuar em rota de falha.
- **Negativo**: Thresholds fixos (2.0, 70%, 90%) são calibrados heuristicamente. Diferentes domínios podem exigir thresholds distintos — possível evolução para thresholds por agente no `agent_trust`.

## Referências

- Task ID: [ec002a8]
- Implementação: `internal/agents/engine.go:555-710`
