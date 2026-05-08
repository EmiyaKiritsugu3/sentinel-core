# Phase 7.2: Real-Time Entropy Monitor — Design Spec
**Date:** 2026-05-07
**Status:** Approved by Sovereign Council

## 1. Executive Summary

This specification outlines the architecture for the **Real-Time Entropy Monitor** (Phase 7.2). Since raw `logprobs` are currently unavailable via the Go SDK for Gemini 3.1 streams, the Sentinel employs a **Hybrid Security Funnel** to detect and interrupt model hallucinations during code generation. This hybrid approach combines mathematical analysis of the model's reasoning effort with structural validation of its output.

## 2. Architecture: The Hybrid Funnel

The Circuit Breaker operates in two sequential phases within `internal/agents/engine.go`.

### 2.1. Gate A: Cognitive Averaging (Thinking Entropy)

This gate exploits Gemini 3.1's `ThinkingConfig` to measure the density of the reasoning process before accepting the execution.

* **The Metric ($\lambda$):** $\lambda = \frac{\text{Tokens de Ação (Código Gerado)}}{\text{Tokens de Raciocínio (Thought)}}$
* **Zero-Thought Handling:** `CalculateLambda(action_tokens, thought_tokens)` must return a defined value when `thought_tokens == 0`; the current implementation uses `action_tokens` as the finite high-risk lambda proxy. If that value exceeds the threshold, clients surface it as a Gate A interruption rather than a division-by-zero failure.
* **Mechanism:** If $\lambda$ exceeds a defined threshold, it mathematically indicates "high predictive uncertainty" (lazy thought vs. massive output). The engine interrupts the execution.
* **StepBudget Accounting:** Gate A consumes 1 `StepBudget` unit when it halts execution due to high lambda. Gate B consumes 1 `StepBudget` unit when it blocks filesystem tool execution after isomorphism = 0. If both gates trigger in separate phases of the same task, each consumes its own unit.
* **Dynamic Sovereignty:** The maximum allowed lambda ($\lambda_{max}$) is not static. It is defined per specialist profile in `AgentDefinition`.
  * *Implementer*: High tolerance ($\lambda_{max} = 5.0$)
  * *Security Auditor*: Low tolerance ($\lambda_{max} = 0.5$)
* **Implementation:** The `CalculateLambda` function will be added to `internal/math/formulas.go`.

### 2.2. Gate B: Isomorphism Proxy (Structural Validation)

This is the ultimate "Hard Gate." It ensures that syntactically broken code never reaches the file system.

* **Mechanism:** When Gate A passes and the AI attempts to use a filesystem tool (`replace`, `write_file`), the content is intercepted in-memory.
* **In-Memory Parsing:** Based on the file extension (`.go`, `.ts`), the content is parsed using the existing Tree-sitter scanners (developed in Phase 3). Unsupported extensions pass through with a warning flag/log entry instead of semantic validation, so downstream consumers can revalidate or block later.
* **Isomorphism Evaluation:** If the generated AST contains an `ERROR` or `MISSING` node (native Tree-sitter error indicators), isomorphism = 0.
* **Line Reporting:** For Tree-sitter `ERROR` nodes, derive user-facing `line X` from the node `start_point.row + 1`; prefer `start_point` over `end_point`, with byte-offset-to-position conversion as fallback.
* **Feedback Loop:** The tool execution is blocked, and an error is injected back into the LLM context: `"Structural Audit Failed: Code generates invalid AST near line X. Fix the syntax before writing."` This consumes a unit of the `StepBudget`.

## 3. Implementation Path

1. **Metric Integration:** Add `CalculateLambda` to `internal/math`.
2. **Configuration:** Extend `AgentDefinition` to include `MaxLambda`.
3. **Engine Interception:** Update `Engine.Execute` to intercept `part.Thought` and calculate $\lambda$ during streaming responses.
4. **Tool Wrapping:** Create a secure wrapper around filesystem tools that injects the in-memory Tree-sitter validation (Gate B) before allowing disk writes.
5. **Testing:** Implement unit tests for both Gate A calculations and Gate B AST interceptions.
