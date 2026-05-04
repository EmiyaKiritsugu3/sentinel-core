# Graph Report - .  (2026-05-04)

## Corpus Check
- Corpus is ~30,127 words - fits in a single context window. You may not need a graph.

## Summary
- 368 nodes · 585 edges · 33 communities detected
- Extraction: 63% EXTRACTED · 37% INFERRED · 0% AMBIGUOUS · INFERRED: 215 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Agent Tools & Definitions|Agent Tools & Definitions]]
- [[_COMMUNITY_Core Engine & Protocols|Core Engine & Protocols]]
- [[_COMMUNITY_Task & State Management|Task & State Management]]
- [[_COMMUNITY_Git Shield & Registry|Git Shield & Registry]]
- [[_COMMUNITY_Command Infrastructure|Command Infrastructure]]
- [[_COMMUNITY_AST Scanners & Schema|AST Scanners & Schema]]
- [[_COMMUNITY_Agent Execution Context|Agent Execution Context]]
- [[_COMMUNITY_Graph Engine & Utils|Graph Engine & Utils]]
- [[_COMMUNITY_ADR & Socratic Engine|ADR & Socratic Engine]]
- [[_COMMUNITY_C4 Visualization|C4 Visualization]]
- [[_COMMUNITY_Legacy TS System|Legacy TS System]]
- [[_COMMUNITY_Sovereign Validator|Sovereign Validator]]
- [[_COMMUNITY_Architecture Docs|Architecture Docs]]
- [[_COMMUNITY_Graph Data Types|Graph Data Types]]
- [[_COMMUNITY_Compliance & Standards|Compliance & Standards]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]

## God Nodes (most connected - your core abstractions)
1. `NewStartCmd` - 19 edges
2. `NewRootCmd` - 12 edges
3. `Engine` - 12 edges
4. `NewManager()` - 12 edges
5. `NewAuditCmd` - 11 edges
6. `NewInstructCmd` - 11 edges
7. `Manager` - 11 edges
8. `Engine` - 11 edges
9. `NewScanCmd` - 9 edges
10. `TestDecomposeTool()` - 9 edges

## Surprising Connections (you probably didn't know these)
- `NewAuditCmd` --implements--> `Project Sentinel Protocol`  [INFERRED]
  cmd/sentinel/commands/audit.go → GEMINI.md
- `Open-Closed Principle for Commands` --rationale_for--> `NewRootCmd`  [EXTRACTED]
  docs/superpowers/plans/PID-SENTINEL-HUB-DECOUPLING-V2.md → cmd/sentinel/commands/root.go
- `NewInstructCmd` --implements--> `Scout Protocol (Data-Driven Intent)`  [EXTRACTED]
  cmd/sentinel/commands/instruct.go → docs/process/sentinel-log.md
- `NewInstructCmd` --implements--> `Executable Contracts (ADR Verification)`  [EXTRACTED]
  cmd/sentinel/commands/instruct.go → docs/process/sentinel-log.md
- `NewStartCmd` --calls--> `RegisterCoreTools()`  [INFERRED]
  cmd/sentinel/commands/start.go → /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core/internal/agents/tools.go

## Hyperedges (group relationships)
- **Sovereign Governance Loop** — audit_newauditcmd, plan_newplancmd, status_newstatuscmd, gemini_sentinel_protocol [INFERRED 0.85]
- **Graph Extraction Engine** — engine_scanproject, scanner_treesitter_scan, scanner_go_scan, linker_linkdependencies [EXTRACTED 1.00]

## Communities (35 total, 18 thin omitted)

### Community 0 - "Agent Tools & Definitions"
Cohesion: 0.06
Nodes (9): ADRTool, AuditTool, DecomposeTool, GrepSearchTool, ReadFileTool, ReplaceTool, ScanTool, WriteFileTool (+1 more)

### Community 1 - "Core Engine & Protocols"
Cohesion: 0.07
Nodes (27): Loader, mockAuthProvider, Registry, Tool, TestSovereignAuthProvider_GetAPIKey(), init(), NewStartCmd(), Init() (+19 more)

### Community 2 - "Task & State Management"
Cohesion: 0.1
Nodes (19): NewAuditCmd, Runner, ADR, ContextNode, ContextPayload, Factory, Sentinel Audit Command, Sentinel Plan Command (+11 more)

### Community 3 - "Git Shield & Registry"
Cohesion: 0.09
Nodes (15): Dispatcher, GitShield, mockValidator, RegistryManager, RunTool, ADF Protocol, Evolution Insights, Sentinel Protocol (GEMINI.md) (+7 more)

### Community 4 - "Command Infrastructure"
Cohesion: 0.07
Nodes (20): NewAggregator(), init(), init(), init(), Sentinel Report Command, init(), Execute(), NewRootCmd() (+12 more)

### Community 5 - "AST Scanners & Schema"
Cohesion: 0.09
Nodes (18): Sentinel Scan Command, init(), NewScanCmd(), Memory Integrity & Thread-Safety (Standard #10), Engine.ScanProject, GoScanner, NewGoScanner(), NewTreeSitterScanner() (+10 more)

### Community 6 - "Agent Execution Context"
Cohesion: 0.11
Nodes (11): AgentContext, AgentDefinition, AuthProvider, Engine, Message, SovereignAuthProvider, Specialist, SubTask (+3 more)

### Community 7 - "Graph Engine & Utils"
Cohesion: 0.13
Nodes (6): MutationEngine, NewIgnoreFilter(), Engine, CalculateHash(), Incremental Scanning Pattern, IgnoreFilter

### Community 8 - "ADR & Socratic Engine"
Cohesion: 0.12
Nodes (16): NewADRGenerator(), TestADRGenerator_Generate(), Sentinel Instruct Command, NewInstructCmd(), performDiagnostic(), runSocraticInterview(), God Objects Diagnosis, ADRData (+8 more)

### Community 9 - "C4 Visualization"
Cohesion: 0.24
Nodes (6): init(), NewVisualizeCmd(), Visualizer, NewVisualizer(), NewVisualizeCmd(), NewVisualizer()

### Community 10 - "Legacy TS System"
Cohesion: 0.28
Nodes (8): runBrainstorm(), audit(), forge(), main(), plan(), verifyPlan(), loadConfig(), forgePlan()

### Community 11 - "Sovereign Validator"
Cohesion: 0.24
Nodes (5): Sovereign Gate Pattern, Validator, Violation, isIgnored(), NewValidator()

### Community 12 - "Architecture Docs"
Cohesion: 0.32
Nodes (8): ADR-306b0ea4 Implementar Markdownlint, ADR-d0555ca9 Implementar Foundational Quality Gate, Engineering Standards, Sentinel Sovereign Roadmap, Sentinel Log, Sentinel Core System Design, Wiki: AST Language Expansion, Quality Firewall Implementation Plan

### Community 13 - "Graph Data Types"
Cohesion: 0.4
Nodes (4): Edge, FileScanner, Node, ScanResult

### Community 14 - "Compliance & Standards"
Cohesion: 0.5
Nodes (4): ADR-66d2618f Melhorar Performance, ADR-ad9933bf Refatorar Camada de Persistencia, ADR-fe2bb6f9 Decisão Crítica Caracteres Especiais, Compliance Dashboard

## Knowledge Gaps
- **65 isolated node(s):** `ADR`, `ContextNode`, `ContextPayload`, `Node`, `Edge` (+60 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewStartCmd` connect `Core Engine & Protocols` to `Agent Tools & Definitions`, `Task & State Management`, `Git Shield & Registry`, `Command Infrastructure`, `Agent Execution Context`, `Sovereign Validator`?**
  _High betweenness centrality (0.150) - this node is a cross-community bridge._
- **Why does `NewRootCmd` connect `Command Infrastructure` to `Core Engine & Protocols`, `Task & State Management`, `AST Scanners & Schema`, `ADR & Socratic Engine`, `C4 Visualization`?**
  _High betweenness centrality (0.095) - this node is a cross-community bridge._
- **Why does `NewAuditCmd` connect `Task & State Management` to `Core Engine & Protocols`, `Sovereign Validator`, `Command Infrastructure`?**
  _High betweenness centrality (0.065) - this node is a cross-community bridge._
- **Are the 17 inferred relationships involving `NewStartCmd` (e.g. with `NewRootCmd` and `NewManager()`) actually correct?**
  _`NewStartCmd` has 17 INFERRED edges - model-reasoned connections that need verification._
- **Are the 8 inferred relationships involving `NewRootCmd` (e.g. with `NewAuditCmd` and `NewPlanCmd`) actually correct?**
  _`NewRootCmd` has 8 INFERRED edges - model-reasoned connections that need verification._
- **Are the 11 inferred relationships involving `NewManager()` (e.g. with `NewAuditCmd` and `NewPlanCmd`) actually correct?**
  _`NewManager()` has 11 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `NewAuditCmd` (e.g. with `NewManager()` and `.GetActiveTask()`) actually correct?**
  _`NewAuditCmd` has 10 INFERRED edges - model-reasoned connections that need verification._