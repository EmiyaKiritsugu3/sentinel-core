---
task_id: "ec002a9"
title: "Gate A + Gate A.5 Entropy Circuit Breakers"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a9: Gate A + Gate A.5 Entropy Circuit Breakers

## Contexto

Modelos de linguagem em modo autônomo tendem a degenerar para dois modos de falha: (1) geração excessiva de código sem planejamento (baixa razão pensamento/ação) e (2) divergência progressiva de raciocínio que leva a alucinações. O Sentinel precisa detectar e interromper ambos os modos antes que comprometam a integridade do código.

## Decisão

Implementamos dois circuit breakers complementares no loop ReAct do engine:

**Gate A — Entropy Threshold**: Calcula `lambda = ActionTokens / ThoughtTokens` cumulativo da sessão. Se `lambda > effectiveMaxLambda`, onde `effectiveMaxLambda = MaxLambda * TrustToDynamicLambda(priorTrust)`, o gate dispara. A confiança do agente (`priorTrust`) atua como modulador dinâmico: agentes confiáveis (trust→1.0) têm threshold relaxado (1.5x), agentes não confiáveis (trust→0) têm threshold restrito (0.5x). A intervenção injeta uma mensagem instruindo o agente a reavaliar sua estratégia e produzir raciocínio detalhado antes de gerar código.

**Gate A.5 — Lyapunov Divergence Detection**: Calcula `divergence = |lambda_step - lambda_prev| / max(lambda_prev, 1e-9)` entre passos consecutivos. Se `divergence > 1.0` (variação >100% entre passos), incrementa `DivergenceCount`. Se duas divergências consecutivas (`DivergenceCount >= 2`), o gate dispara com mensagem instruindo re-planejamento completo. Passos estáveis resetam o contador para 0.

Ambos os gates registram intervenções como `SessionEvent` (tipo `EventPattern`) no `GlobalBuffer`, com tags `["gate-a", "entropy", "intervention"]` ou `["gate-a5", "divergence", "intervention"]`.

## Consequências

- **Positivo**: Proteção em duas camadas — Gate A previne geração impulsiva, Gate A.5 detecta deriva de raciocínio.
- **Positivo**: Dynamic lambda via `TrustToDynamicLambda` personaliza thresholds por agente sem configuração manual.
- **Positivo**: Intervenções são recuperáveis — o agente recebe feedback textual e continua o loop, não há hard abort.
- **Negativo**: Falsos positivos possíveis em tarefas naturalmente intensivas em código. O agente pode ser interrompido desnecessariamente, consumindo passos do budget. Mitigação parcial via `trustScore` que relaxa o gate para agentes comprovadamente confiáveis.

## Referências

- Task ID: [ec002a9]
- Implementação: `internal/agents/engine_helpers.go:77-125`, `internal/math/formulas.go:32-34,38-43`
