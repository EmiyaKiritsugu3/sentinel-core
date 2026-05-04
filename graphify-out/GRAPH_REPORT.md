# Graph Report - sentinel-core  (2026-05-04)

## Corpus Check
- 62 files · ~61,371 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 362 nodes · 588 edges · 38 communities detected
- Extraction: 58% EXTRACTED · 42% INFERRED · 0% AMBIGUOUS · INFERRED: 245 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]

## God Nodes (most connected - your core abstractions)
1. `NewStartCmd()` - 19 edges
2. `Engine` - 13 edges
3. `NewRootCmd()` - 12 edges
4. `NewManager()` - 12 edges
5. `NewAuditCmd()` - 11 edges
6. `Register()` - 10 edges
7. `NewInstructCmd()` - 9 edges
8. `NewEngine()` - 9 edges
9. `TestDecomposeTool()` - 9 edges
10. `Engine` - 9 edges

## Surprising Connections (you probably didn't know these)
- `NewAuditCmd()` --calls--> `NewRunner()`  [INFERRED]
  cmd/sentinel/commands/audit.go → internal/audit/runner.go
- `NewAuditCmd()` --implements--> `Project Sentinel Protocol`  [INFERRED]
  cmd/sentinel/commands/audit.go → GEMINI.md
- `NewRootCmd()` --calls--> `GetCommands()`  [INFERRED]
  cmd/sentinel/commands/root.go → internal/registry/commands.go
- `init()` --calls--> `Register()`  [INFERRED]
  cmd/sentinel/commands/scan.go → internal/registry/commands.go
- `NewStartCmd()` --calls--> `RegisterCoreTools()`  [INFERRED]
  cmd/sentinel/commands/start.go → internal/agents/tools.go

## Hyperedges (group relationships)
- **Sovereign Governance Loop** — audit_newauditcmd, plan_newplancmd, status_newstatuscmd, gemini_sentinel_protocol [INFERRED 0.85]
- **Graph Extraction Engine** — engine_scanproject, scanner_treesitter_scan, scanner_go_scan, linker_linkdependencies [EXTRACTED 1.00]

## Communities

### Community 0 - "Community 0"
Cohesion: 0.06
Nodes (9): ADRTool, AuditTool, DecomposeTool, GrepSearchTool, ReadFileTool, ReplaceTool, ScanTool, WriteFileTool (+1 more)

### Community 1 - "Community 1"
Cohesion: 0.09
Nodes (23): NewAggregator(), NewAuditCmd(), Factory, Cognitive Loop, TestDecomposeTool(), Antigravity AI Assistant, Project Sentinel Protocol, Sovereign Council (Warden, Auditor) (+15 more)

### Community 2 - "Community 2"
Cohesion: 0.1
Nodes (17): Dispatcher, Loader, RegistryManager, Init(), InitAtPath(), NewDispatcher(), TestDispatcher_ReconcileEvents(), setupTestDB() (+9 more)

### Community 3 - "Community 3"
Cohesion: 0.08
Nodes (16): mockAuthProvider, Registry, Tool, ADR, ContextNode, ContextPayload, NewEngine(), NewRegistry() (+8 more)

### Community 4 - "Community 4"
Cohesion: 0.12
Nodes (11): handleGetGraph(), Runner, TestSovereignAuthProvider_GetAPIKey(), NewLiveCmd(), GraphSnapshot, Server, Technical Gate Pattern, NewRunner() (+3 more)

### Community 5 - "Community 5"
Cohesion: 0.13
Nodes (6): MutationEngine, NewIgnoreFilter(), Engine, CalculateHash(), Incremental Scanning Pattern, IgnoreFilter

### Community 6 - "Community 6"
Cohesion: 0.14
Nodes (10): NewADRGenerator(), TestADRGenerator_Generate(), GitShield, NewGitShield(), TestGitShield_CreateWorktree(), ADRData, ADRGenerator, EscapeYAML() (+2 more)

### Community 7 - "Community 7"
Cohesion: 0.11
Nodes (11): init(), GetCommands(), Register(), init(), init(), init(), CommandFactory, init() (+3 more)

### Community 8 - "Community 8"
Cohesion: 0.16
Nodes (3): Engine, mockValidator, RunTool

### Community 9 - "Community 9"
Cohesion: 0.28
Nodes (8): runBrainstorm(), audit(), forge(), main(), plan(), verifyPlan(), loadConfig(), forgePlan()

### Community 10 - "Community 10"
Cohesion: 0.38
Nodes (3): Visualizer, NewVisualizeCmd(), NewVisualizer()

### Community 11 - "Community 11"
Cohesion: 0.24
Nodes (5): Sovereign Gate Pattern, Validator, Violation, isIgnored(), NewValidator()

### Community 12 - "Community 12"
Cohesion: 0.18
Nodes (8): AgentContext, AgentDefinition, Message, Specialist, SubTask, TokenBudget, Validator, NewAgentContext()

### Community 13 - "Community 13"
Cohesion: 0.32
Nodes (8): ADR-306b0ea4 Implementar Markdownlint, ADR-d0555ca9 Implementar Foundational Quality Gate, Engineering Standards, Sentinel Sovereign Roadmap, Sentinel Log, Sentinel Core System Design, Wiki: AST Language Expansion, Quality Firewall Implementation Plan

### Community 14 - "Community 14"
Cohesion: 0.4
Nodes (4): Edge, FileScanner, Node, ScanResult

### Community 15 - "Community 15"
Cohesion: 0.5
Nodes (3): EventType, GraphEvent, Observer

### Community 16 - "Community 16"
Cohesion: 0.5
Nodes (2): AuthProvider, SovereignAuthProvider

### Community 17 - "Community 17"
Cohesion: 0.67
Nodes (4): ADF Protocol, Evolution Insights, Sentinel Protocol (GEMINI.md), Wiki Index

### Community 18 - "Community 18"
Cohesion: 0.5
Nodes (4): ADR-66d2618f Melhorar Performance, ADR-ad9933bf Refatorar Camada de Persistencia, ADR-fe2bb6f9 Decisão Crítica Caracteres Especiais, Compliance Dashboard

### Community 21 - "Community 21"
Cohesion: 1.0
Nodes (2): Project Master Architecture, Graphify Report

### Community 22 - "Community 22"
Cohesion: 1.0
Nodes (2): ADR-841fa0a2 Self-Audit for Standard Compliance, Memory Query: Manager Connection

### Community 28 - "Community 28"
Cohesion: 1.0
Nodes (1): God Objects Diagnosis

### Community 29 - "Community 29"
Cohesion: 1.0
Nodes (1): Project README

### Community 30 - "Community 30"
Cohesion: 1.0
Nodes (1): Technical Debt Log

### Community 31 - "Community 31"
Cohesion: 1.0
Nodes (1): ADR-3c3075f2 Documentation Evolution

### Community 32 - "Community 32"
Cohesion: 1.0
Nodes (1): ADR-0baf39c0 Pre-Flight Check Local

### Community 33 - "Community 33"
Cohesion: 1.0
Nodes (1): Git Shield Implementation Plan

### Community 34 - "Community 34"
Cohesion: 1.0
Nodes (1): Sovereign Factory Plan Phase 5.2

### Community 35 - "Community 35"
Cohesion: 1.0
Nodes (1): Agent: Sovereign Architect

### Community 36 - "Community 36"
Cohesion: 1.0
Nodes (1): Graphify HTML Visualizer

### Community 37 - "Community 37"
Cohesion: 1.0
Nodes (1): Protocolo de Epifania (Filtros A, B, C)

### Community 38 - "Community 38"
Cohesion: 1.0
Nodes (1): Scout Protocol (Data-Driven Intent)

### Community 39 - "Community 39"
Cohesion: 1.0
Nodes (1): Executable Contracts (ADR Verification)

### Community 40 - "Community 40"
Cohesion: 1.0
Nodes (1): Fase 3: The Language Expansion

### Community 41 - "Community 41"
Cohesion: 1.0
Nodes (1): Fase 4: The Agentic State Machine

### Community 42 - "Community 42"
Cohesion: 1.0
Nodes (1): Fase 5: The Visual Sovereign

### Community 43 - "Community 43"
Cohesion: 1.0
Nodes (1): Open-Closed Principle for Commands

### Community 44 - "Community 44"
Cohesion: 1.0
Nodes (1): Memory Integrity & Thread-Safety (Standard #10)

## Knowledge Gaps
- **59 isolated node(s):** `ADR`, `ContextNode`, `ContextPayload`, `Node`, `Edge` (+54 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 16`** (4 nodes): `AuthProvider`, `SovereignAuthProvider`, `.GetAPIKey()`, `auth_provider.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 21`** (2 nodes): `Project Master Architecture`, `Graphify Report`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 22`** (2 nodes): `ADR-841fa0a2 Self-Audit for Standard Compliance`, `Memory Query: Manager Connection`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 28`** (1 nodes): `God Objects Diagnosis`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 29`** (1 nodes): `Project README`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 30`** (1 nodes): `Technical Debt Log`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 31`** (1 nodes): `ADR-3c3075f2 Documentation Evolution`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 32`** (1 nodes): `ADR-0baf39c0 Pre-Flight Check Local`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 33`** (1 nodes): `Git Shield Implementation Plan`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 34`** (1 nodes): `Sovereign Factory Plan Phase 5.2`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 35`** (1 nodes): `Agent: Sovereign Architect`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 36`** (1 nodes): `Graphify HTML Visualizer`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 37`** (1 nodes): `Protocolo de Epifania (Filtros A, B, C)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 38`** (1 nodes): `Scout Protocol (Data-Driven Intent)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 39`** (1 nodes): `Executable Contracts (ADR Verification)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 40`** (1 nodes): `Fase 3: The Language Expansion`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 41`** (1 nodes): `Fase 4: The Agentic State Machine`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 42`** (1 nodes): `Fase 5: The Visual Sovereign`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 43`** (1 nodes): `Open-Closed Principle for Commands`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 44`** (1 nodes): `Memory Integrity & Thread-Safety (Standard #10)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewStartCmd()` connect `Community 2` to `Community 0`, `Community 1`, `Community 3`, `Community 6`, `Community 7`, `Community 11`, `Community 12`?**
  _High betweenness centrality (0.126) - this node is a cross-community bridge._
- **Why does `NewRootCmd()` connect `Community 1` to `Community 10`, `Community 2`, `Community 3`, `Community 7`?**
  _High betweenness centrality (0.067) - this node is a cross-community bridge._
- **Why does `NewAuditCmd()` connect `Community 1` to `Community 11`, `Community 4`, `Community 7`?**
  _High betweenness centrality (0.047) - this node is a cross-community bridge._
- **Are the 18 inferred relationships involving `NewStartCmd()` (e.g. with `NewManager()` and `.StartTask()`) actually correct?**
  _`NewStartCmd()` has 18 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `NewRootCmd()` (e.g. with `GetCommands()` and `Factory`) actually correct?**
  _`NewRootCmd()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 11 inferred relationships involving `NewManager()` (e.g. with `NewAuditCmd()` and `NewPlanCmd()`) actually correct?**
  _`NewManager()` has 11 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `NewAuditCmd()` (e.g. with `NewManager()` and `.GetActiveTask()`) actually correct?**
  _`NewAuditCmd()` has 10 INFERRED edges - model-reasoned connections that need verification._