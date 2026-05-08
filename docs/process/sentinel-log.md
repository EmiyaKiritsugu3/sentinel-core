# Sentinel Log — Compiled Brain [PID-SENTINEL]

## [2026-05-07] Milestone: Entropy Monitor Gate B [PID-SENTINEL-ENTROPY-GATE-B]

**Status**: COMPLETED ✅
**Impact**: HIGH (Structural Security & Hallucination Hard-Stop)

### 🔍 Analysis (Epiphanies)

1. **In-Memory AST Validation**: Implemented Gate B (`internal/agents/ast_validator.go`) as a hard gate in `tools.go` that intercepts `write_file` / `replace` calls before they reach the filesystem. Uses `go/parser` for `.go` files and Tree-sitter for `.ts`/`.tsx`, injecting a structured error back into the LLM context on `ERROR` or `MISSING` AST nodes.
2. **Feedback Loop Discipline**: The error message returned to the LLM is deterministic: `"Structural Audit Failed: Code generates invalid AST near line X."` This consumes a StepBudget unit, ensuring the circuit-breaker eventually escalates rather than looping forever.
3. **Separation of Concerns**: Gate B lives in its own file (`ast_validator.go`) rather than inline in `tools.go`, enforcing the Rule of Single Responsibility across the interceptor surface.

### 💡 Key Learning

"A última barreira contra alucinações estruturais não é o raciocínio do modelo — é uma validação determinística em memória antes do commit no disco. O Gate B transforma o filesystem em território soberano."

---

## 🏁 SOVEREIGN HANDOVER [S25-ENTROPY-GATE-B -> S26-TRUST-CALIBRATION]

**Status**: STABLE 🛡️
**Success Rate**: 100% (Gates A + B both live, all tests green)

### 🚀 Current Vector

O Hybrid Funnel (Phase 7.2) está completo: Gate A filtra por entropia cognitiva (λ), Gate B filtra por integridade estrutural (AST). O próximo vetor é a **Bayesian Trust Calibration** (Phase 7, Pillar D): ajuste dinâmico de `MaxLambda` baseado no histórico de erros por agente.

### ⚠️ Technical Snag

Os parâmetros `probHallucination` e `bugWeight` em `CalculateDelta` ainda são estáticos (0.5 e 5.0). A Trust Calibration irá alimentar esses valores dinamicamente a partir do `TrustScore` histórico.

### 🎯 Chief's Priority (First Command)

**"Sentinel, implemente o Bayesian Trust Calibration (Pillar D). Adicione TrustScore tracking ao SQLite, implemente CalculateTrustScore em internal/math/formulas.go e wire no Engine para ajustar MaxLambda dinamicamente."**

---

## [2026-04-29] Milestone: Sovereign Audit & Recovery Cycle

**Status**: COMPLETED
**Impact**: HIGH (System Integrity & Development Strategy)

### 🔍 Analysis (Epiphanies)

1. **CLI Resilience**: Identified a critical input blocking bug in Gemini CLI v0.40.0 (Linux/WSL). Validated `Ctrl + J` as the sovereign bypass for non-responsive Enter keys in interactive prompts.
2. **Standards Enforcement**: Applied automated remediation for Standards STD-01 (Buffered I/O) and STD-05 (Error Governance). The audit gate is now 100% compliant.
3. **VCS Modernization**: Renamed default branch from `master` to `main` and pruned legacy branches, achieving architectural purity in the Git record.

### 💡 Key Learning

"A tool that cannot be audited is a liability. A process that cannot be simplified is a trap. Stability is the precursor to autonomy."

---

## [2026-04-29] Milestone: The KISS Pivot [PID-SENTINEL-5.8]

**Status**: IN_PROGRESS
**Impact**: STRATEGIC (High)

### 🔍 Analysis (Epiphanies)

1. **Premature Optimization Trauma**: Identified a move towards over-engineering (Protocolo Bonsai/SOLID Module) before delivering basic functionality.
2. **Pareto Alignment**: Realized that 80% of project value currently resides in functional task decomposition, not in automated refactoring of non-existent code.
3. **Strategic Realignment**: Pivoted the Phase 5.8 plan to a "KISS & Deliver" model, focusing on the minimum viable tool for task breakdown.

### 💡 Key Learning

"Do not build a shipyard before you have a ship. Architecture must emerge from necessity, not from a desire for infinite polish."

---

## [2026-05-03] Sovereign Loop Audit (Security Session)

**Status**: AUDITED 🛡️
**Impact**: HIGH (Process Stability)

### 🔍 Analysis (Findings)

1. **Generation Protection**: Validated that `MaxSteps` in `Engine.go` atua como um disjuntor (circuit breaker) infalível contra loops de tentativa-erro em agentes.
2. **Audit Safety**: O `AuditRunner` foi confirmado com timeout de 30s, prevenindo travamento por comandos bloqueantes ou interativos.
3. **Recursive Vulnerability**: Identificado que o sistema carece de um limite de profundidade para sub-tarefas geradas recursivamente. Risco adicionado ao `TECHNICAL-DEBT.md`.

### 💡 Key Learning

"A arquitetura de agentes é uma árvore de recursão. Sem um limite de profundidade, a autonomia se torna instabilidade. O próximo nível de maturidade exige 'Depth Governance'."

---

## [2026-05-05] Milestone: Input Disambiguation Audit (Task 5) [PID-SENTINEL-DISAMBIGUATOR-AUDIT]

**Status**: AUDITED 🛡️
**Impact**: MEDIUM (Input Integrity & DX)

### 🔍 Analysis (Findings)

1. **Spec Compliance**: A implementação do `Disambiguator` e `VaguenessScore` segue rigorosamente a Spec da Task 5. Os 4 sinais (Length, Verb, Pronoun, Anchor) e as duas fases do Anchor Signal foram validados no código.
2. **Anchor Phase 2 Verification**: Confirmado que o `anchorSignal` utiliza corretamente o `matched_ratio` consultando o SQLite para reduzir o score de vagueza quando o usuário menciona símbolos existentes no grafo.
3. **Suggestion Engine**: O método `queryGraph` limita corretamente a 5 sugestões ancoradas no grafo, prevenindo poluição visual no prompt do usuário.
4. **Test Gap Resolved**: Adicionado `disambiguator_db_test.go` para cobrir o cenário de análise ancorada em DB, que estava ausente na suíte original do implementador.

### 💡 Key Learning

"Um sistema de disambiguação é tão inteligente quanto sua âncora na realidade (o grafo). Validar que o sistema reconhece o código existente como uma âncora de clareza é o que separa um chatbot de um engenheiro autônomo."

---

**Status**: COMPLETED
**Impact**: STRATEGIC (High - Process Governance)

### 🔍 Analysis (Epiphanies)

1. **Vagueness vs. Evidence**: Implemented the "Scout" protocol. The Sentinel now uses the `graph.db` (SQLite) to find "God Objects" and hotspots when the user provides vague intents (e.g., "improve performance"), transforming abstract ideas into data-driven proposals.
2. **Executable Governance**: Transitioned ADRs from static documentation to "Executable Contracts". Every ADR now requires a `Verification Protocol` (shell command).
3. **The Hard Gate**: A task's completion is now deterministic. It requires the ADR's verification command to return Exit Code 0, preventing "guessing" and ensuring solid, step-by-step progress.

### 💡 Key Learning (continued)

"Documentation is only as strong as its ability to be verified. An ADR that cannot be tested is merely a suggestion. A protocol that enforces its own rules is the foundation of true autonomy."

---

---

## [2026-05-03] Milestone: AST Expansion & Start Snag Resolution [PID-SENTINEL-AST-EXPANSION]

**Status**: COMPLETED 🛡️
**Impact**: HIGH (Architectural Capability & Operational Stability)

### 🔍 Analysis (Epiphanies)

1. **Engine Resilience**: O desacoplamento da inicialização da engine no comando `sentinel start` permite que o sistema opere em modo "Local-First" quando o `GOOGLE_API_KEY` está ausente. Isso remove o bloqueio técnico para progressão manual de tarefas.
2. **Tree-sitter Integration**: A transição de regex para o motor real Tree-sitter (`smacker/go-tree-sitter`) elevou a precisão da extração de símbolos em TypeScript/TSX. O sistema agora mapeia componentes React, interfaces e classes com fidelidade sintática.
3. **Elite Concurrency (Standard #10)**: Identificada e corrigida uma condição de corrida (Race Condition) no scanner AST. Implementado `sync.Pool` para gerenciar instâncias de `sitter.Parser`, garantindo que múltiplos workers operem simultaneamente sem corromper a memória CGO.
4. **Memory Integrity (Standard #07)**: Implementada a liberação explícita de memória CGO via `tree.Close()` e `cursor.Close()`. Sem isso, o scanner vazaria memória proporcional ao tamanho do projeto.
5. **Declarative Extraction**: Migrado de travessia manual de árvore para `Query-based extraction`. O uso de S-expressions (Lisp-like) permite capturar padrões complexos como interfaces e componentes com maior robustez e menor overhead de código.

### 💡 Key Learning

"A inteligência de um scanner é proporcional à sua compreensão da gramática, não de padrões de texto. Em sistemas multi-thread que utilizam CGO, a segurança de memória e a concorrência não são opcionais—são os alicerces da soberania tecnológica."

---

## 🏁 SOVEREIGN HANDOVER [S08-AST-EXPANSION -> S09-COMPONENT-RELATIONS]

**Status**: STABLE 🛡️
**Success Rate**: 100% (Elite AST Gates Operational)

### 🚀 Current Vector

Capacidade de análise multi-linguagem fortificada. O Sentinel agora processa Go e TypeScript com segurança de memória e alta performance (concorrente). O próximo vetor estratégico é o mapeamento profundo de dependências entre componentes (Relações de Renderização).

### ⚠️ Technical Snag

Embora a extração de símbolos seja agora 'Elite', o linker ainda requer a resolução de caminhos relativos em imports de TypeScript para IDs de arquivo canônicos.

### 🎯 Chief's Priority (First Command)

**"Sentinel, implemente o Linker de Dependências para TypeScript para converter 'import:path' em relações reais entre nós de arquivo no SQLite."**

---

## [2026-05-03] Milestone: Dependency Linker & Grammar Resilience [PID-SENTINEL-LINKER]

**Status**: COMPLETED 🛡️
**Impact**: HIGH (Graph Connectivity & Deep Context)

### 🔍 Analysis (Epiphanies)

1. **Grammar Resilience**: Refatorado o `TreeSitterScanner` para utilizar Queries genéricas sem dependência de nomes de campos rígidos (ex: `name:`). A extração agora percorre os filhos nomeados em busca de identificadores, tornando o Sentinel compatível com múltiplas versões das gramáticas de TypeScript/TSX.
2. **Sovereign Linking (S11)**: Implementado o `Linker` no motor central. O Sentinel agora resolve caminhos relativos em imports e cria relações reais de dependência entre arquivos no SQLite. Isso elimina as "ilhas de código" e permite uma análise de impacto sistêmica.
3. **CGO Stability**: Corrigidos erros de segmentação em ambientes concorrentes através da validação rigorosa de ponteiros nulos em objetos Tree-sitter (`tree`, `query`, `cursor`).

### 💡 Key Learning

"Um grafo de dependências sem links é apenas um inventário. A inteligência reside na conexão. Ao resolver imports brutos em relações canônicas, transformamos o banco de dados em um mapa de navegação real para agentes autônomos."

---

## 🏁 SOVEREIGN HANDOVER [S11-DEPENDENCY-LINKER -> S12-C4-VISUALIZATION]

**Status**: STABLE 🛡️
**Success Rate**: 100% (Dependency Linker Operational)

### 🚀 Current Vector

Capacidade de mapeamento relacional completa. O Sentinel agora entende não apenas o que está dentro dos arquivos, mas como eles se conectam. O próximo vetor estratégico é a geração automática de diagramas C4 dinâmicos baseados nestas relações reais.

### ⚠️ Technical Snag

A resolução de imports atuais foca em arquivos locais. Imports de pacotes externos (ex: `inquirer`, `fs`) são mantidos como `unresolved_import`. Futuras versões podem integrar um mapeamento de `node_modules` ou `go.mod`.

### 🎯 Chief's Priority (First Command)

**"Sentinel, utilize o novo grafo conectado para gerar um diagrama C4 de Nível 2 (Container) em tempo real."**

---

## [2026-05-03] Milestone: C4 Visualization & Go AST Expansion [PID-SENTINEL-C4-VIS]

**Status**: COMPLETED 🛡️
**Impact**: HIGH (Architecture Visibility & DX)

### 🔍 Analysis (Epiphanies)

1. **C4 Level 2 Automation**: Implementada a geração automática de diagramas C4 de Nível 2 (Container) via Mermaid. O Sentinel agora agrega milhares de micro-relações em um mapa de alto nível, permitindo visualizar a interação entre `CLI`, `Agents`, `Graph`, `Audit` e `Database`.
2. **Go AST Maturity**: Expandido o `GoScanner` para extrair declarações de importação. Isso permite que o Sentinel mapeie dependências internas do projeto Go com a mesma precisão do TypeScript, unificando a visão sistêmica.
3. **Cross-Language Linking**: O `Linker` foi aprimorado para resolver módulos Go (baseados no `go.mod`) e caminhos relativos TS/JS, garantindo que o grafo de containers reflita a realidade física do repositório.

### 💡 Key Learning

"A transparência arquitetural é o maior inimigo da entropia. Ao transformar o grafo bruto em uma visão C4, permitimos que tanto humanos quanto agentes compreendam o impacto de mudanças sistêmicas antes da primeira linha de código ser escrita."

---

## [2026-05-04] Milestone: Linker Integration & Test Resilience [PID-SENTINEL-LINKER-TESTS]

**Status**: COMPLETED 🛡️
**Impact**: HIGH (Graph Reliability & Testability)

### 🔍 Analysis (Epiphanies)

1. **Testable Persistence**: Refatorado o pacote `sqlite` para suportar `InitAtPath`, permitindo a criação de bancos de dados isolados em diretórios temporários. Isso resolve o problema de poluição do banco de dados de produção durante a execução de testes.
2. **FileSystem Isolation**: Implementada a prática de utilizar `t.TempDir()` combinada com `os.Chdir()` para simular estruturas de diretórios complexas. Isso garante que as heurísticas de resolução de import do `Linker` sejam validadas em um ambiente controlado e reproduzível.
3. **Deep Resolution Coverage**: A suíte de testes agora valida resoluções de múltiplos níveis (`../../`), index patterns (`/index.ts`), e variadas extensões Web (`.tsx`, `.ts`). O `Linker` foi provado resiliente contra estruturas de pastas profundas.
4. **Database Integrity Audit**: Validado que o `Linker` limpa corretamente os nós de `unresolved_import` apenas após a criação bem-sucedida das relações reais, mantendo a consistência do grafo.

### 💡 Key Learning

"A confiabilidade de um grafo é ditada pela precisão de seus links. Uma suíte de testes que simula a realidade física do projeto é o único 'Hard Gate' aceitável para garantir que a inteligência relacional do Sentinel não degrade com a evolução das gramáticas."

---

## [2026-05-04] Milestone: Live View WebSocket Infrastructure [PID-SENTINEL-LIVE-VIEW-INFRA]

**Status**: COMPLETED 🛡️
**Impact**: HIGH (Interactive Architecture Visibility)

### 🔍 Analysis (Epiphanies)

1. **Non-Blocking Observability**: Implementamos o padrão Observer no `Engine`, porém garantimos que o envio para os observadores seja não-bloqueante (`select { case s.broadcast <- event: default: }`). Isso é crucial; se o servidor WebSocket estiver lento ou clientes acumularem, a análise AST não será penalizada. A performance do motor CGO prevalece.
2. **Concurrent Data Safety**: O uso de `sync.RWMutex` tanto no gerenciamento de observadores no `Engine` quanto nas conexões de clientes no `liveview.Server` previne as condições de corrida validadas no `go test -race`.
3. **Graceful Degradation (Standard #05)**: Falhas de `json.Marshal` nos eventos GraphEvent são interceptadas e logadas isoladamente, impedindo o *panic* do hub de eventos inteiro. A integridade do servidor foi fortificada contra payloads malformados.

### 💡 Key Learning

"A visualização em tempo real não pode comprometer a velocidade da extração de dados subjacente. A dissociação através de canais não-bloqueantes garante que a UI atue como um espelho assíncrono da realidade, e não como uma âncora que atrasa o progresso."

---

## [2026-05-04] Milestone: Live View Frontend (S15) [PID-SENTINEL-LIVE-FRONTEND]

**Status**: COMPLETED 🛡️
**Impact**: HIGH (Interactive Architecture Visibility & DX)

### 🔍 Analysis (Epiphanies)

1. **Frictionless DX (KISS Protocol)**: Decidimos não usar um servidor Node/Next.js separado em desenvolvimento/produção. Em vez disso, usamos Vite + React para compilar arquivos estáticos (`web/dist`) que o próprio servidor Go serve (`http.FileServer`). Isso garante que o usuário rode apenas um comando (`sentinel live`) para subir backend e frontend.
2. **Race Condition Prevention (Buffer Protocol)**: Para garantir que nenhum evento seja perdido, o frontend usa um React hook (`useSentinelData`) que: 1. Conecta no WebSocket. 2. Coloca os eventos `NODE_UPSERTED` e `EDGE_CREATED` num array temporário. 3. Faz o fetch HTTP de `/api/graph`. 4. Renderiza tudo. 5. Dá o flush do array.
3. **DOM Performance (Cytoscape vs React)**: Para suportar milhares de eventos sem congelar a UI, não usamos state do React para gerir os nós. O hook usa `useRef` e chama os métodos mutáveis nativos do Cytoscape (`cy.add()`), pulando o Virtual DOM do React totalmente para atualizações de grafo em tempo real.

### 💡 Key Learning

"A integração entre backend de alta performance (Go) e frontends declarativos (React) pode gerar gargalos se não respeitarmos a natureza de cada engine. Pular o Virtual DOM e escrever diretamente no Canvas do Cytoscape foi essencial para a viabilidade do Live View sob carga maciça."

---

## 🏁 SOVEREIGN HANDOVER [S15-LIVE-FRONTEND -> S16-SENTINEL-ORCHESTRATOR]

**Status**: STABLE 🛡️
**Success Rate**: 100% (Frontend and Backend Integrated)

### 🚀 Current Vector

A Fase 5 (Visual Sovereign) tem seu esqueleto totalmente funcional. O usuário pode rodar `sentinel live` e visualizar o grafo auto-atualizável no navegador via React e Cytoscape. O próximo passo lógico do *ROADMAP.md* é expandir o **Subagent Dispatcher** (Fase 5.8 Early Access), também conhecido como o sistema de orquestração de especialistas (Agentic State Machine) baseado na tríade Warden, Auditor, Chief.

### ⚠️ Technical Snag

A UI do Cytoscape está funcional mas ainda básica visualmente. Layouts automáticos como 'cose' ou 'concentric' podem sobrepor nós se o grafo for excessivamente denso sem uma categorização (compound nodes) de diretórios.

### 🎯 Chief's Priority (First Command)

**"Sentinel, revise a infraestrutura de agentes (`internal/agents`) para garantir que o dispatcher consiga lidar com múltiplos subagentes concorrentes durante a resolução de problemas complexos."**

---
Related: [ROADMAP.md](../architecture/ROADMAP.md) | [PID-SENTINEL-QUALITY-FIREWALL](../superpowers/plans/PID-SENTINEL-QUALITY-FIREWALL.md) | [ADR-d0555ca9](../architecture/adr/ADR-d0555ca9-2139-4b67-8261-5a64afd44e24-implementar-golangci-lint-foundational-layer.md)

---

## [2026-05-04] Milestone: Security Hardening & CodeRabbit Autofix [PID-SENTINEL-CODERABBIT-S16]

**Status**: COMPLETED 🛡️
**Impact**: HIGH (Security & Concurrency Correctness)

### 🔍 Analysis (Epiphanies)

1. **WebSocket Single-Writer Protocol**: O padrão correto do Gorilla WebSocket exige que apenas **um goroutine** escreva em cada `*websocket.Conn`. Quando o hub (`Run`) e o `writePump` escrevem simultaneamente, cria-se uma race condition silenciosa detectável apenas com `go test -race`. A solução canônica é introduzir uma struct `wsClient` com `send chan []byte` por cliente — o hub escreve no canal, e o `writePump` (único goroutine escritor) drena o canal.
2. **Prefix-Match como Falsa Segurança**: `strings.HasPrefix(origin, "http://localhost")` aceita `http://localhost.evil.com`. Validação de origem via URL **sempre deve usar `url.Parse` + `u.Hostname()`** para comparação exata, nunca prefixo de string.
3. **Autofix Introduz Novos Bugs**: Durante o autofix do CodeRabbit, o fix do `CheckOrigin` introduziu um bypass de segurança que não estava no código original. Isso valida que **todo fix de segurança requer auditoria independente** — o revisor de código capturou o bypass antes do merge.
4. **Testes Determinísticos vs. Sleep**: `time.Sleep` em testes não é um mecanismo de sincronização — é uma aposta. O padrão correto é um **poll loop com deadline**: `for { check condition; select { case <-deadline: t.Fatal(...) default: time.Sleep(1ms) } }`. Isso elimina flakiness sem adicionar dependências.

### 💡 Key Learning

"Segurança e concorrência não são camadas opcionais. Um `return true` no `CheckOrigin` e um `_ =` silenciando erros de rollback são duas formas distintas de dívida técnica que se acumulam silenciosamente até um incidente. O Semgrep e o race detector são os únicos 'Hard Gates' confiáveis para esses vetores."

---

## 🏁 SOVEREIGN HANDOVER [S16-SECURITY-HARDENING -> S17]

**Status**: STABLE 🛡️
**Success Rate**: 100% (All CVEs patched, race conditions resolved)

### 🚀 Current Vector

PR #6 mergeado. A infraestrutura de LiveView está segura e livre de race conditions. CVEs críticos do Dependabot (grpc + oauth2) foram corrigidos. O próximo vetor é expandir o **Subagent Dispatcher** conforme definido no ROADMAP.md (S17+).

### ⚠️ Technical Snag

`liveview.Server.Run` não fecha conexões ativas quando o contexto é cancelado — goroutines de `readPump`/`writePump` sobrevivem até TCP timeout. Adicionado ao TECHNICAL-DEBT.md.

### 🎯 Chief's Priority (First Command)

**"Verificar se os alertas Dependabot fecharam automaticamente após o merge do PR #6, e iniciar o próximo ciclo de desenvolvimento conforme o ROADMAP.md."**

---

## [2026-05-04] Milestone: Educational Questionnaire Automation & Truncation Resilience [PID-SENTINEL-QUIZ-S17]

**Status**: COMPLETED 🛡️
**Impact**: MEDIUM (Process Resilience & Context Management)

### 🔍 Analysis (Epiphanies)

1. **Atomic Message Failure**: Identificado o erro `could not convert a single message before hitting truncation`. Isso ocorre quando um único tool output (ex: `take_snapshot(verbose: true)`) excede o limite físico de processamento do pipeline do Gemini, impedindo até o truncamento automático.
2. **Surgical Research Protocol (S17)**: Formalizada a transição de "Full Snapshot" para "Targeted DOM Scraping" via `evaluate_script` em páginas web massivas. Extrair apenas JSON filtrado com os textos e IDs necessários reduziu o consumo de tokens em ~95% e eliminou os crashes de truncamento.
3. **Dialectical Learning (Trial & Error)**: O processo de aprendizado nesta sessão foi iterativo. A falha inicial em persistir a marcação das respostas levou ao refinamento da técnica de clique (coordenadas e verificação de estado), provando que o agente pode adaptar sua estratégia de interação dinamicamente.
4. **Quiz Domain Logic**: Resolvido questionário de 6 questões sobre Engenharia de Software (Paradigmas Orientado a Objetos, Estruturado e Componentes).

### 💡 Key Learning

"A obesidade de contexto é a morte da agência. Em ambientes web complexos, a 'visão total' é um risco técnico; a inteligência reside na filtragem cirúrgica. Além disso, a falha é o gatilho da adaptação: um agente soberano não apenas repete comandos, ele refina sua própria heurística de interação após cada atrito."

---

## [2026-05-04] Milestone: Prompt Intelligence Design & Audit Process Formalization [PID-SENTINEL-S19]

**Status**: COMPLETED 🧠
**Impact**: HIGH (Architecture Quality & Development Process)

### 🔍 Analysis (Epiphanies)

1. **Audit Depth Rule — Design Quality is Calculable**: Revisões de design sem auditoria explícita acumulam dívida técnica silenciosa. Toda sessão deve aplicar a matriz `BlastRadius × Reversibility` para determinar quantos rounds de auditoria são necessários. Alto impacto + difícil reversão = 2 rounds. Condição de parada absoluta: se um round não encontra issues críticas ou majors, para-se imediatamente.

2. **Confidence como Razão de Palavras é Sempre Errado**: O algoritmo `matched_keywords / total_words` produz confidence ≈ 0.14 para qualquer descrição descritiva longa — derrota o propósito do heurístico. O algoritmo correto é `presença + ambiguidade`: 1 categoria matchada → confidence 0.85 (heurístico vence); 2+ categorias → 0.30 (AI fallback); 0 categorias → 0.00 (AI fallback). Este padrão se aplica a qualquer sistema de classificação tiered.

3. **Package Boundary Detectável em Design**: `internal/bridge/` constrói payloads para AI. Input do usuário (Disambiguator) pertence a `internal/intake/`. Violar este boundary em design cria acoplamento que testa coesão do pacote e dificulta testes unitários isolados. Regra: se o pacote precisa importar algo que o usuário digita, ele está no layer errado.

4. **Infinite Optimization Anti-Pattern**: Otimizar o design além do ponto de retorno decrescente impede o MVP. Regra de orçamento: se o tempo de design excede 30% do tempo estimado de implementação, commitar o design e implementar. A implementação revela problemas mais rápido que discussão adicional de design.

5. **GenAI Client Lifecycle como Responsabilidade do Engine**: Dois `genai.Client` com ciclos de vida independentes = resource leak garantido. O `Engine` já cria e possui o cliente. Qualquer componente downstream (classifier, factory) deve receber o cliente via injeção, nunca criar o próprio.

6. **Token Optimization via Padrões de Sessão**: Sessões longas com AI geram ~10% de conteúdo durável. O valor não está em indexar transcrições — está em capturar padrões de falha (raciocínio sem dado, confidence algorithm errado, package boundary violado) como conhecimento transferível entre projetos. O Audit Depth Rule desta sessão é um exemplo: um insight de 10 linhas que evita horas de debug em qualquer feature futura.

### 💡 Key Learning

"Design sem auditoria explícita é confirmação de viés em velocidade de desenvolvimento. A auditoria não procura perfeição — ela troca o modo cognitivo de construtivo para destrutivo, expondo o que o cérebro preenche automaticamente durante a criação. Dois rounds de auditoria desta sessão encontraram: 1 race condition, 1 algoritmo de confidence invertido, 2 package boundaries incorretos, 1 resource leak de cliente AI — todos antes de uma linha de código ser escrita."

---

## [2026-05-05] Milestone: GeminiClassifier Hardening (Code Quality) [PID-SENTINEL-CLASSIFIER-HARDENING]

**Status**: COMPLETED 🛡️
**Impact**: MEDIUM (Stability & Observability)

### 🔍 Analysis (Epiphanies)

1. **Nil Pointer Defense**: Identified a potential panic in `GeminiClassifier` where `resp.Candidates[0].Content` was accessed without validation. Defensive programming is mandatory when dealing with external AI provider responses.
2. **Type-Safe Extraction**: Replaced fragile `fmt.Sprintf` with explicit type assertions for `genai.Text`. Interface-based systems require strict type verification to ensure data integrity.
3. **Observability Gap**: Unrecognized model outputs are no longer silently ignored. Added warning logs to `os.Stderr` to facilitate debugging of AI classification drifts.

### 💡 Key Learning

"Data from external APIs must be treated as untrusted. Nil checks and type assertions are the 'Hard Gates' that prevent remote failures from becoming local panics."

---

## 🏁 SOVEREIGN HANDOVER [S20-CLASSIFIER-HARDENING -> S21]

**Status**: STABLE 🧠
**Success Rate**: 100% (Spec + Plan commitados, 2 rounds de auditoria completos)

### 🚀 Current Vector

Spec (`docs/superpowers/specs/2026-05-04-prompt-intelligence-design.md`) e plano de implementação (`docs/superpowers/plans/2026-05-04-prompt-intelligence.md`) commitados em main. 6 tasks definidas em dois blocos independentes: Block B (Smart Context Routing — Tasks 1-4) e Block A (Input Disambiguation — Tasks 5-6). Block B pode ser mergeado antes de Block A ser iniciado.

### ⚠️ Technical Snag

`KNOWN_LIMITATION_03`: `--refine` em contexto não-interativo (pipe, CI sem `--no-suggest`) causa hang no stdin. Fix futuro: detectar TTY via `golang.org/x/term`. Documentado no spec.

### 🎯 Chief's Priority (First Command)

**"Implementar Block B do Prompt Intelligence System (Tasks 1-4): classifier.go, gemini_classifier.go, router.go, e wiring no Engine+Factory. Usar superpowers:subagent-driven-development."**

---

## 🏁 SOVEREIGN HANDOVER [S17-QUIZ-RESILIENCE -> S18-ORCHESTRATOR-DISPATCH]

**Status**: STABLE 🛡️
**Success Rate**: 100% (Quiz complete, Truncation issue documented, Strategy refined)

### 🚀 Current Vector

Questionário concluído com 100% de precisão. A resiliência contra erros de truncamento e a capacidade de adaptação em interações web foram fortificadas. O próximo vetor é retornar ao foco do *ROADMAP.md*: o **Subagent Dispatcher** e a orquestração de especialistas no `internal/agents`.

### ⚠️ Technical Snag

O `take_snapshot` padrão ainda é perigoso em páginas com milhares de nós. O sistema deve emitir um aviso ou falhar graciosamente se detectar um DOM excessivamente profundo antes de gerar a mensagem.

### 🎯 Chief's Priority (First Command)

**"Sentinel, agora que o questionário está resolvido e o processo de aprendizado documentado, retome o ROADMAP.md e comece a implementação do Dispatcher de Agentes no pacote `internal/agents`."**

---

## [2026-05-05] Milestone: Prompt Intelligence System Deployment [PID-SENTINEL-COMPLETE]

**Status**: COMPLETED 🧠
**Impact**: HIGH (AI Quality & UX)

### 🔍 Analysis (Epiphanies)

1. **Context Efficiency**: Smart Context Routing (Subsystem B) is now live. The Sentinel selects context (ADRs, tests, debt) based on detected intent (diagnose/implement/refactor/review), reducing token waste and improving agent focus.
2. **User Sovereignty**: Input Disambiguation (Subsystem A) with `--refine` flag allows users to anchor vague tasks in the graph before persistence. This ensures the \"Task Objective\" is high-signal from the start.
3. **Graceful Resilience**: The system handles missing AI keys or graph data by falling back to deterministic heuristics, ensuring the CLI remains functional in all environments.

### 💡 Key Learning

\"Intention is the anchor of intelligence. By classifying user intent before execution and disambiguating input before storage, we transform the AI pipeline from a reactive loop into a goal-oriented architectural guide.\"

---

## 🏁 SOVEREIGN HANDOVER [S21-PROMPT-INTELLIGENCE -> S22]

**Status**: STABLE 🛡️
**Success Rate**: 100% (Block A & B fully implemented and verified)

### 🚀 Current Vector

Prompt Intelligence System (Fase 6) concluído. O Sentinel agora é consciente da intenção e do grafo no momento do input. O próximo vetor é retornar ao ROADMAP para a **Fase 4: Subagent Dispatcher** (Agentic State Machine).

### ⚠️ Technical Snag

O linkador de dependências Go ainda não mapeia tipos externos (`third_party`) com a mesma profundidade do TypeScript.

### 🎯 Chief's Priority (First Command)

**"Sentinel, com a inteligência de prompt estabilizada, inicie a expansão do Subagent Dispatcher (internal/agents/dispatcher.go) para suportar a orquestração da tríade Warden/Auditor/Chief."**

---

## [2026-05-06] Milestone: Sovereign Math Engine Phase 1 [PID-SENTINEL-SME-P1]

**Status**: COMPLETED 💎
**Impact**: HIGH (Mathematical Observability & Stability)

### 🔍 Analysis (Epiphanies)

1. **Atomic Persistence (STD-03)**: Identified that multi-step migrations without transactions lead to partial states. Hardened the persistence layer with `BeginTx/Commit/Rollback`, ensuring that the "Mathematical Proof" (schema) is never corrupted.
2. **Deterministic Metric Collection**: Integrated high-precision timing and token tracking into the core Engine loop. The system no longer "guesses" efficiency; it calculates it via the Net Gain Equation ($\Delta$).
3. **Semantic Error Filtering**: Implemented "Smart Error Governance" for migrations, ignoring specifically "duplicate column name" while reporting all other SQL failures. This balances idempotency with transparency.

### 💡 Key Learning

"A inteligência artificial é probabilística por natureza. Tentar governá-la com lógica puramente binária é um erro de categoria. A Soberania Matemática permite que o Sentinel meça a entropia da criação e a eficiência da correção em tempo real."

---

## 🏁 SOVEREIGN HANDOVER [S23-SME-P1 -> S24]

**Status**: STABLE 🛡️
**Success Rate**: 100% (Foundational metrics and persistence live)

### 🚀 Current Vector

A base matemática está sólida. O Sentinel agora coleta latência, tokens e custo de cada sub-agente. O próximo vetor estratégico é a **Fase 7.2: Real-Time Entropy Monitor**, para interromper alucinações via análise de incerteza preditiva (Shannon Entropy).

### ⚠️ Technical Snag

Os parâmetros de "Probabilidade de Alucinação" e "Peso do Bug" na fórmula de $\Delta$ são atualmente valores estáticos. Precisam ser movidos para a `AgentDefinition` na próxima fase.

### 🎯 Chief's Priority (First Command)

**"Sentinel, inicie a Fase 7.2: Real-Time Entropy Monitor. O foco é implementar o Gate A (λ threshold) e o Gate B (AST validation) no loop da Engine para interromper alucinações em tempo real."**

---

## [2026-05-07] Milestone: Entropy Monitor Gate A [PID-SENTINEL-ENTROPY-GATE-A]

**Status**: COMPLETED (Partial) 🛡️
**Impact**: HIGH (Cognitive Security & Halucination Prevention)

### 🔍 Analysis (Epiphanies)

1. **Validation Vulnerability (Pointer Semantics)**: Identified that declaring `MaxLambda` as a primitive `float64` bypassed validation rules. Since `omitempty` treats an explicitly provided `0.0` as "empty", the `min=0.1` rule was silently skipped. Converted to `*float64` to seal this loophole, reinforcing the principle that validation parameters must support nullability to differentiate omitted values from explicit zeroes.
2. **Cognitive Averaging**: Successfully implemented Gate A ($\lambda$) calculation inside `engine.go`. The system now inspects streams for thought patterns (`<think>`) and calculates the ratio of action vs. thought tokens.
3. **Execution Interruption**: If the generated code volume massively exceeds the reasoning volume, the engine preemptively stops execution and forces the model to re-plan, consuming a step budget.

### 💡 Key Learning

"A type system is only as safe as its boundary validations. In Go, relying on primitive types for optional validation is a silent failure waiting to happen. The transition to pointer-based configuration for critical security metrics (MaxLambda) ensures that ignorance and malicious intent are distinct and handleable states."

---

## 🏁 SOVEREIGN HANDOVER [S24-ENTROPY-GATE-A -> S25-ENTROPY-GATE-B]

**Status**: PAUSED 🛑
**Success Rate**: 75% (Plan audited, Task 1-3 complete, Gate A live)

### 🚀 Current Vector

A matemática de Entropia (CalculateLambda) e o "Gate A" foram injetados no laço da Engine (`engine.go`). O sistema já intercepta alucinações (muito código, pouco pensamento). A execução foi pausada a pedido do Chief antes da inicialização do Gate B.

### ⚠️ Technical Snag

A implementação do Gate B requer manipulação rigorosa da árvore AST com `go-tree-sitter` (para TS/TSX) e `go/parser` (para Go). Foi identificada uma restrição prévia (ausência do módulo golang tree-sitter local) que foi devidamente abordada no plano atualizado (`docs/superpowers/plans/2026-05-07-real-time-entropy-monitor.md`), separando o validador em um arquivo dedicado (`ast_validator.go`).

### 🎯 Chief's Priority (First Command)

**"Sentinel, retome a execução do plano do Monitor de Entropia a partir da Task 4. O foco exclusivo é implementar o Gate B (Structural Validation) e os testes integrados de interceptação no `tools.go`."**
