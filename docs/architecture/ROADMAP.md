# Sentinel Sovereign Roadmap [PID-SENTINEL]

Este documento define a trajetória oficial de desenvolvimento do Sentinel Core. Nenhuma tarefa deve ser iniciada sem estar mapeada neste roadmap.

## 🏁 Milestones Alcançados

- [x] **Fase 1: The Fail-Safe Foundation**
  - Implementação de Timeouts de Auditoria.
  - Governança de Erros com Wrapping.
  - Definição da Tríade (Warden, Chief Engineer, Operator).
- [x] **Fase 2: The Context Engine**
  - Scanner de AST (Go) paralelo via Worker Pool.
  - Persistência em SQLite (CGO-free).
  - Extração cirúrgica de linhas de código real.
- [x] **Fase 2.10: Sovereign Hardening**
  - Refatoração para Injeção de Dependência (Fim das Globais).
  - Blindagem de Segurança (shlex, Foreign Keys, Transactions).
  - Implementação do Sovereign Validator (Hard Gates).
- [x] **Fase 2.11: Linter Cleanup & Quality Firewall**
  - Complexidade ciclomática reduzida: 6 funções refatoradas (gocyclo ≤ 15).
  - 131 exported doc comments adicionados em 20+ arquivos.
  - Thread safety no Registry: `sync.RWMutex` + `go test -race` pass.
  - Context threading no Disambiguator (noctx fix).
  - `golangci-lint run`: 0 issues (de 137).
- [x] **Fase 2.12: SonarCloud Code Quality Remediation**
  - Tradução de ~100 comentários PT→EN em 30 arquivos.
  - Redução de complexidade cognitiva em 2 funções (executeQuery 31→≤15, buildNodeToContainerMap 19→≤15).
  - Duplicate literals extraídos para constantes (S1192).
  - Parâmetros consecutivos agrupados (S8209).
  - Fix do timestamp RFC3339 para compatibilidade com modernc.org/sqlite.
  - CI action version reparada (SHA quebrado → v7.1.0).
  - `golangci-lint run`: 0 issues, `SonarCloud Scan`: ✅ PASS.

## 🚀 Próximas Frentes (O Plano Concreto)

### Fase 3: The Language Expansion (AST Evolution)

O Sentinel deve ser capaz de gerir projetos Web de vanguarda.

- [x] Abstração da Engine Multi-Linguagem (Orchestrator).
- [x] Integração com **Tree-sitter** (smacker/go-tree-sitter).
- [x] Scanner AST para **TypeScript/TSX**.
- [x] Mapeamento de dependências entre arquivos e componentes.
*Critério de Sucesso: `sentinel scan` mapeia projetos Go e TypeScript com 100% de conectividade.*

### Fase 4: The Agentic State Machine (Proactive Governance)

Transformar o Sentinel em um guia proativo para o usuário.

- [x] Saneamento de Grafo via .gitignore (Hybrid Filter).
- [x] Modo Entrevista (Comando `instruct` blindado para CI/CD).
- [x] **Auto-ADR**: Gera rascunhos técnicos precisos e baseados em dados do grafo (Protocolo Scout).
- [x] **Hard Gate Verification**: Vincula comandos de verificação aos ADRs para garantir progresso sólido.
- [x] **Dashboard Visibility**: Vincula fisicamente tarefas aos ADRs no relatório.
- [x] **Subagent Dispatcher**: Infraestrutura nativa para orquestração de especialistas (Fase 5.8 Early Access).
*Critério de Sucesso: Criação de uma feature completa apenas via diálogo, sem intervenção manual no plano.*

### Fase 5: The Visual Sovereign (Live UI)

A visualização de arquitetura deve ser interativa.

- [x] **C4 Container View**: Geração automática de diagramas de Nível 2 (Containers).
- [x] **Sentinel Live View**: Servidor WebSocket em Go que envia o Grafo para uma UI Web (Sprints 0-2: WebSocket server, StatusHUD, FilterToolbar, EventLog).
- [x] **Interactive C4**: Clique no nó do diagrama para abrir o código ou ver o ADR relacionado (Sprint 3: InfoPanel, code viewer, ADR links, path traversal hardening).
*Critério de Sucesso: Visualização em tempo real no browser enquanto o código muda.*

### Fase 6: The Prompt Intelligence Layer

Tornar o sentinel consciente da intenção do usuário e do contexto semântico.

- [x] **Subsystem B — Smart Context Routing**: `IntentClassifier` (heurístico + Gemini fallback) + `ContextRouter` que seleciona nodes por intent (diagnose/implement/refactor/review). Spec: `docs/superpowers/specs/2026-05-04-prompt-intelligence-design.md`.
- [x] **Subsystem A — Input Disambiguation**: `Disambiguator` com `VaguenessScore` que detecta descrições vagas e sugere alternativas ancoradas no grafo. Flags `--refine` / `--no-suggest` no `sentinel plan`.
*Critério de Sucesso: `sentinel plan "fix bug"` exibe sugestão ancorada no grafo; `GeneratePayload` injeta contexto diferenciado por intent.*

---

### Fase 7: The Mathematical Sovereignty (The Final Frontier)

Elevar o Sentinel ao estado de Oráculo Matemático via Prova de Estabilidade.

- [x] **Sovereign Math Engine (SME)**: Motor central de cálculo de métricas e estabilidade.
- [x] **Real-Time Entropy Monitor**: Interrupção de alucinações via Gates A (λ) e B (AST).
- [x] **Lyapunov Divergence Detection**: Detecção de "Logic Drift" antes da renderização. *Implemented: `checkGateA5` in `internal/agents/engine_helpers.go`; wired in `engine.go` line 193.*
- [x] **Bayesian Trust Calibration**: Ajuste dinâmico de rigor baseado em evidência histórica. *Implemented: `CalculateTrustScore`/`TrustToDynamicLambda` in `internal/math/formulas.go`; `readPriorTrust`/`persistTrust` in `internal/agents/engine_helpers.go`; `agent_trust` table in `internal/graph/schema.go`.*
*Critério de Sucesso: Interrupção autônoma de 90% das alucinações baseada em métricas de probabilidade.*

---

## 🔭 Fase Futura: The Knowledge Sovereign

Ideias com mérito técnico validado, sem data definida. Cada uma tem uma **pré-condição explícita** — não iniciar sem validá-la.

### 🔨 In Progress: Pattern Capture — taxonomia de falhas e sucessos de workflow

**O que é:** Estrutura de `docs/patterns/` que registra padrões de desenvolvimento que funcionaram e que falharam — não prompts específicos, mas princípios: "diagnóstico sem dado empírico = loop", "especificar artefato de saída reduz iterações", "auditoria troca modo cognitivo de construtivo para destrutivo".

**Por que é útil:** Os mesmos anti-padrões reaparecem em projetos diferentes. Capturar uma vez, aplicar sempre.

**Pré-condição:** Nenhuma técnica — só disciplina de captura ao final de cada sessão. Iniciado em 2026-05-24 com catálogo inicial de 12 padrões do histórico do projeto.

### Semantic Search — `sentinel search "query"`

**O que é:** Busca semântica por similaridade vetorial nos documentos do projeto (ADRs, knowledge base, sentinel-log, TECHNICAL-DEBT). Usa Gemini Embedding API + SQLite para armazenar vetores. Retorna documentos relacionados por significado, não por keyword.

**Por que é útil:** Quando a base de documentos cresce além de ~100 arquivos, `grep` começa a retornar ruído. Busca semântica encontra "race condition em goroutine" mesmo que o documento use "concorrência sem mutex".

**Pré-condição obrigatória:** Base de conhecimento com 100+ documentos curados. Construir infraestrutura de busca para base vazia é desperdício de manutenção.

**Stack mínima:** `gemini.EmbedContent()` → vetor em SQLite BLOB → cosine similarity em Go (sem nova dependência). Complementa Obsidian + Smart Connections, não substitui.

### Session Debrief — `sentinel debrief`

**O que é:** Comando que roda ao final de cada sessão de desenvolvimento. Extrai insights da sessão atual (padrões que funcionaram, padrões que falharam, decisões arquiteturais), categoriza por domínio (hardware/methodology/tools/systems) e persiste em `~/knowledge/`. Alimenta o banco de conhecimento que torna o Semantic Search útil.

**Por que é útil:** Sessões longas com AI geram ~10% de conteúdo durável. Sem captura estruturada, esse conhecimento morre com o contexto. Com debrief sistemático, cada sessão alimenta a próxima.

**Pré-condição:** Definir estrutura de `~/knowledge/` e template de debrief antes de automatizar. A automação sem curadoria manual validada indexa ruído.

### Smart CLAUDE.md — `sentinel context "query"`

**O que é:** Dado um domínio ou task, o sentinel seleciona automaticamente os arquivos de `~/knowledge/` mais relevantes e os injeta no CLAUDE.md do projeto. O AI começa cada sessão com contexto real do histórico de decisões, não do zero.

**Por que é útil:** Elimina o "AI esquece tudo entre sessões" — o contexto histórico fica no arquivo, não na memória do modelo. Especialmente valioso em projetos com múltiplos domínios sobrepostos.

**Pré-condição:** Semantic Search implementado + base de conhecimento com 50+ documentos.

---

*Atualizado em: 2026-05-24*
*Assinado: Sovereign Council*
