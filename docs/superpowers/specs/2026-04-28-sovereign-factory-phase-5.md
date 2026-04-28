# Design Doc: The Sovereign Factory (Phase 5) [PID-SENTINEL]

## 1. Vision & Purpose
To transform Sentinel Core from a proactive assistant into a **State-of-the-Art Digital Engineering Firm**. The system must autonomously transform ideas into high-quality software products by strictly adhering to vanguarde engineering standards, deterministic protocols, and multi-perspective deliberation.

## 2. Structural Divisions (The Production Line)

### 2.1 Solution Discovery (The Architect)
- **Objective**: Convert vague intent into elite documentation.
- **Workflow**: Socratic dialogue to identify requirement gaps.
- **Mandatory Output Suite (Tier 3)**:
    - **PRD**: Business value and criteria for success.
    - **RFC/Design Doc**: Architectural alternatives and trade-offs.
    - **Smart ADR**: Permanent record of the selected decision.
    - **Technical Spec**: Detailed interfaces, data schemas, and **Mandatory Mermaid Diagrams** (Sequence and Class/Component).
- **Hard Gate**: Tier 3 execution only starts after the Chief Engineer approves the "Visual Contract" (Mermaid diagrams).

### 2.2 Execution Planning (The Chief Engineer)
- **Objective**: Create a deterministic blueprint for implementation.
- **Workflow**: Map tasks into sequential steps with explicit **Acceptance Criteria (AC)**.
- **Visual Progress**: Progress is mapped via Mermaid Gantt/State diagrams in the `implementation_plan.md`.
- **Hard Gate**: Steps must be verifiable via physical evidence (AST state, build status, or test results).

### 2.3 Specialist Execution (The Operators)
- **Objective**: Surgery-level code implementation.
- **Model Hierarchy**:
    - **Standard Operation**: `gemini-1.5-flash` for routine coding and boilerplate.
    - **Crisis Operation**: `gemini-1.5-pro` for architectural shifts and complex debugging.
- **Standard Adherence**: Mandatory use of `bufio.Scanner` (Standard #01) and safe concurrency (Standard #06).

### 2.4 Sovereign QA (The Auditor)
- **Objective**: Final validation and epigenetic learning.
- **Outputs**: Sovereign Audit Report (5-point framework).

## 3. Resilience & Recovery: Protocolo de Angulação Crítica (PAC)

When an agent fails a Hard Gate 3 times, the **PAC [PID-SENTINEL-PAC]** is activated:

1.  **Trauma Decomposition**: Analyze logs and AST Diffs to identify the root cause.
2.  **Tripartite Deliberation (Multi-Perspective)**:
    - **Angle A (Minimalist)**: Seek simpler, YAGNI-compliant solutions.
    - **Angle B (Structuralist)**: Evaluate if the Execution Plan itself needs a pivot.
    - **Angle C (Auditor)**: Check for hidden system locks or environment conflicts.
3.  **Intelligence Escalation**: Promote the task to `gemini-1.5-pro` for a final "Sovereign Pivot" attempt.

## 4. Sovereign Stress & Reliability (Certainty Testing)
... (AST Load Stress, Crisis Injection, Concurrency Pressure)

## 5. Seamless Authentication: The Sovereign Identity
...

## 6. Git-Native Orchestration: The Versioning Shield
To ensure code integrity and local tool authority, the Sentinel Core will implement a **Git-Driven Lifecycle**:
- **Local Authority**: Subagents execute tools natively on the host machine (no sandboxing) for maximum performance and tool access.
- **Ephemeral Branches**: Every task starts in an automatically generated branch (e.g., `sentinel/task-{ID}`).
- **Atomic Commit Pattern**: Successful completion of an Implementation Step (Hard Gate Passed) triggers an automatic semantic commit (e.g., `feat(core): ...`).
- **Sovereign Merge**: Final integration into the main branch only occurs after the **Sovereign Audit Framework** is verified.

## 7. Implementation Roadmap (Phase 5.2 & 5.3)
- [ ] **Step 1**: Implement the Model Escalation Trigger (Flash -> Pro).
- [ ] **Step 2**: Create the `sentinel:get_context` and `sentinel:execute` toolset for subagents.
- [ ] **Step 3**: Build the PAC deliberative state machine in `internal/agents/engine.go`.
- [ ] **Step 4**: Integrate automated Stress Tests into the `audit` command.

---
*Assinado: Sovereign Council & Chief Architect José Inamar*
