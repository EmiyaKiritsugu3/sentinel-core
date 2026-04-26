# Graph Report - .  (2026-04-26)

## Corpus Check
- Corpus is ~10,442 words - fits in a single context window. You may not need a graph.

## Summary
- 117 nodes · 139 edges · 9 communities detected
- Extraction: 80% EXTRACTED · 20% INFERRED · 0% AMBIGUOUS · INFERRED: 28 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_AI Bridge & Prompt Factory|AI Bridge & Prompt Factory]]
- [[_COMMUNITY_AST Scanning & Hashing|AST Scanning & Hashing]]
- [[_COMMUNITY_Architecture Visualization|Architecture Visualization]]
- [[_COMMUNITY_Legacy TypeScript Core|Legacy TypeScript Core]]
- [[_COMMUNITY_Governance & Audit Runner|Governance & Audit Runner]]
- [[_COMMUNITY_Task & State Management|Task & State Management]]
- [[_COMMUNITY_Reporting & Aggregation|Reporting & Aggregation]]
- [[_COMMUNITY_CLI Entry Point & Root|CLI Entry Point & Root]]
- [[_COMMUNITY_SQLite Infrastructure|SQLite Infrastructure]]

## God Nodes (most connected - your core abstractions)
1. `Manager` - 10 edges
2. `Visualizer` - 8 edges
3. `loadConfig()` - 8 edges
4. `GoScanner` - 7 edges
5. `main()` - 7 edges
6. `Factory` - 6 edges
7. `extractLines()` - 5 edges
8. `Execute()` - 4 edges
9. `Validator` - 4 edges
10. `Aggregator` - 4 edges

## Surprising Connections (you probably didn't know these)
- `Fase 4: The Agentic State Machine` --conceptually_related_to--> `Factory`  [INFERRED]
  docs/architecture/ROADMAP.md → /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core/internal/bridge/prompt_factory.go
- `Fase 3: The Language Expansion` --conceptually_related_to--> `GoScanner`  [INFERRED]
  docs/architecture/ROADMAP.md → /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core/internal/graph/scanner_go.go
- `Fase 5: The Visual Sovereign` --conceptually_related_to--> `Visualizer`  [INFERRED]
  docs/architecture/ROADMAP.md → /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core/internal/graph/visualizer.go
- `Instruct Command` --calls--> `Factory`  [EXTRACTED]
  cmd/sentinel/commands/instruct.go → /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core/internal/bridge/prompt_factory.go
- `Scan Command` --calls--> `GoScanner`  [EXTRACTED]
  cmd/sentinel/commands/scan.go → /home/emiyakiritsugu/Projetos_Antigravity/sentinel-core/internal/graph/scanner_go.go

## Communities

### Community 0 - "AI Bridge & Prompt Factory"
Cohesion: 0.17
Nodes (10): ADR, ContextNode, Factory, PromptData, Instruct Command, NewManager(), extractLines(), Fase 4: The Agentic State Machine (+2 more)

### Community 1 - "AST Scanning & Hashing"
Cohesion: 0.16
Nodes (9): Scan Command, edgeData, GoScanner, nodeData, scanResult, CalculateHash(), Fase 3: The Language Expansion, Rationale for Web Avant-garde Support (+1 more)

### Community 2 - "Architecture Visualization"
Cohesion: 0.19
Nodes (7): Visualize Command, Edge, Node, Visualizer, Fase 5: The Visual Sovereign, Rationale for Interactive Architecture, SanitizeID()

### Community 3 - "Legacy TypeScript Core"
Cohesion: 0.28
Nodes (8): runBrainstorm(), audit(), forge(), main(), plan(), verifyPlan(), loadConfig(), forgePlan()

### Community 4 - "Governance & Audit Runner"
Cohesion: 0.2
Nodes (5): Runner, Audit Command, Validator, Violation, isIgnored()

### Community 5 - "Task & State Management"
Cohesion: 0.2
Nodes (5): Plan Command, Start Command, Status Command, Manager, Task

### Community 6 - "Reporting & Aggregation"
Cohesion: 0.29
Nodes (3): Report Command, Aggregator, ProjectStats

### Community 7 - "CLI Entry Point & Root"
Cohesion: 0.5
Nodes (3): commands.Execute, Root Command, main()

### Community 8 - "SQLite Infrastructure"
Cohesion: 0.67
Nodes (1): DB

## Knowledge Gaps
- **22 isolated node(s):** `ADR`, `ContextNode`, `PromptData`, `scanResult`, `nodeData` (+17 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `SQLite Infrastructure`** (3 nodes): `Init()`, `db.go`, `DB`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Manager` connect `Task & State Management` to `AI Bridge & Prompt Factory`, `Governance & Audit Runner`?**
  _High betweenness centrality (0.094) - this node is a cross-community bridge._
- **Are the 7 inferred relationships involving `loadConfig()` (e.g. with `runBrainstorm()` and `plan()`) actually correct?**
  _`loadConfig()` has 7 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `main()` (e.g. with `runBrainstorm()` and `loadConfig()`) actually correct?**
  _`main()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **What connects `ADR`, `ContextNode`, `PromptData` to the rest of the system?**
  _22 weakly-connected nodes found - possible documentation gaps or missing edges._