# Sentinel Core 🛡️
**Deterministic Architectural Governance & Context Compiler Engine**

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Status](https://img.shields.io/badge/Status-Alpha-orange.svg)]()

> "Code without architecture is just noise. Action without verification is just a risk."

Sentinel is a high-performance **Governance Wrapper** designed for the era of AI-Native Software Engineering. It acts as the **Warden** of your codebase, ensuring that AI agents (like Gemini, Claude, or Copilot) follow strict architectural protocols without "hallucinations" or scope drift.

## 🌌 The Vision: The Sovereign Triad
Sentinel introduces the **Subagent Triad Architecture**, separating concerns to achieve maximum scalability and zero context bloat:

*   **The Warden (Sentinel Core Go)**: The source of truth. It maps your codebase using AST (Abstract Syntax Tree), manages state in a local SQLite ledger, and enforces **Hard Gates** via executable ADR contracts.
*   **The Chief Engineer (Main Agent)**: The orchestrator. It reads Sentinel's plans, performs **Data-Driven Diagnostics (Scout Protocol)**, and dispatches surgical tasks to Operators.
*   **The Operators (Subagents)**: Ephemeral, task-specific agents that execute code changes under strict Sentinel constraints.

## 🚀 Key Features
*   **Deterministic Context (AST + SQLite)**: Sentinel scans your code and builds a granular dependency graph. It feeds the AI only the "Surgical Context" (the exact functions/structs needed), saving 90% of token waste.
*   **The Scout (Data-Driven Diagnostics)**: Before creating tasks, Sentinel analyzes your `graph.db` to identify God Objects and architectural hotspots, transforming vague intents into precise plans.
*   **Executable ADRs**: Every decision is linked to a shell command (test/benchmark). No task is marked as `DONE` until the Warden verifies this protocol.

*   **Zero Dependencies**: Compiled in Go, it's a single static binary. No `node_modules`, no runtime friction.

## 🛠️ How it Works (The Sovereign Workflow)

1.  **Plan**: Forge a new architectural goal.
    ```bash
    sentinel plan "Add Auth Service" "go test ./internal/auth/..."
    ```
2.  **Scan**: Update the internal knowledge graph.
    ```bash
    sentinel scan
    ```
3.  **Visualize**: See your architecture in real-time.
    ```bash
    sentinel visualize
    ```
4.  **Audit**: Let the Warden verify the implementation.
    ```bash
    sentinel audit
    ```

## 🏗️ Architecture
Sentinel's core is built for speed and reliability:
*   **Language**: Go (Golang)
*   **Storage**: SQLite (CGO-free via `modernc.org/sqlite`)
*   **Analysis**: Native Go AST Parser (Extending to Tree-sitter)
*   **Workflow**: Finite State Machine (FSM)

## 📜 License
This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details. This ensures patent protection and corporate-grade safety for all contributors and users.

---
Built with 🛡️ by [EmiyaKiritsugu3](https://github.com/EmiyaKiritsugu3)
