# Sentinel Cognitive Evolution System (SEC) — Design Spec
**Date:** 2026-05-06
**Status:** Draft

---

## 1. Problem Statement

AI agents typically learn via "Passive Documentation" (writing lessons into a log file). This creates two critical failures:
1. **Context Obesity:** As logs grow, injecting them into prompts wastes tokens, slows down execution, and degrades AI focus (Lost in the Middle syndrome).
2. **Symptom Myopia:** Agents fix the resulting *code error* (e.g., adding a nil-check) but fail to fix the *cognitive reasoning pattern* that caused the error (e.g., "Trust-by-Context" bias).

**Goal:** Transform Sentinel from a reactive code-fixer into a proactive intelligence that edits its own "Cognitive Lines" (reasoning algorithms) while strictly adhering to the principle of "Resource Minimalism" (using the smallest possible context for maximum leverage).

---

## 2. Core Philosophy: The Dual-Layer Protocol

When an error or friction occurs, Sentinel must process the resolution on two frequencies:
- **Execution Frequency:** Fix the bug in the code.
- **Evolution Frequency:** Identify the flawed reasoning pattern (e.g., "Greedy Matching") and replace it with an Elite Cognitive Strategy (e.g., "Collision Reasoning").

---

## 3. Architecture: The 3 Pillars of SEC

### Pillar 1: The Trigger (Triage)
Not every typo is a lesson. To prevent noise and bloat, evolution is triggered only if a "Friction Threshold" is crossed.
- **Condition:** A Sprint/Task is marked `FP > 5` (High Complexity) OR external audits (like CodeRabbit or test suites) flag > 2 conceptual issues.
- **Action:** At the end of the sprint, the `sentinel evolve` command is invoked.

### Pillar 2: The Injection (Contextual RAG)
We must avoid loading the entire history into the AI's prompt. 
- **Mechanism:** The system leverages the `ContextRouter` (built in Phase 6). 
- **Logic:** `COGNITIVE-DNA.md` is no longer a flat file loaded 100% of the time. It is a structured taxonomy. When the agent starts a task (e.g., `Intent: Diagnose` on `auth.go`), the router extracts *only* the Elite Cognitive Strategies related to debugging and security.

### Pillar 3: The Compaction (Garbage Collection)
Negative constraints ("Don't do X") confuse LLMs.
- **Mechanism:** During `sentinel evolve`, the system reviews the current `COGNITIVE-DNA.md`.
- **Logic:** If multiple similar reasoning flaws exist, the agent refatora (compacts) them into a single, positive **Elite Cognitive Checklist**. This ensures the DNA shrinks in volume but grows in conceptual density.

---

## 4. Implementation Details

### `sentinel evolve` Command
A new CLI command responsible for auditing the session.

**Input:**
- Current `docs/process/sentinel-log.md` (Session history)
- Diff of the current Sprint/PR.

**Process:**
1. Evaluates Friction Threshold.
2. Prompts the Agent (LLM) to perform a "Cognitive Audit": *What reasoning model failed? What elite model replaces it?*
3. Updates `docs/process/COGNITIVE-DNA.md` with the new Cognitive Strategy.
4. Runs "DNA Compaction" to merge redundant rules into positive checklists.

### Integration with `internal/agents/dispatcher.go`
When the Subagent Dispatcher (Phase 4) assigns a task to an expert agent (e.g., Implementer), it attaches the relevant Cognitive Strategy directly to the system prompt, enforcing the desired reasoning algorithm before code generation begins.

---

## 5. Security & Verification
- The `evolve` process itself must run as an isolated subagent to prevent mutating the DNA incorrectly.
- All changes to `COGNITIVE-DNA.md` must be committed and visible in the Git history for human review.
