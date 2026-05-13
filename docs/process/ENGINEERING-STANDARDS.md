# Sentinel Engineering Standards [PID-SENTINEL]

Este documento define os padrões técnicos inegociáveis do Sentinel Core. Toda nova função ou módulo deve aderir a estes padrões desde a primeira linha.

## 💾 I/O & Memory Management

* **Standard #01 (Buffered Reads)**: Nunca utilize `os.ReadFile` para leitura de arquivos de código ou logs. Utilize sempre `bufio.Scanner` ou `io.ReadAll` (para arquivos pequenos como AST chunks) garantindo eficiência de memória.
* **Standard #02 (Stream Processing)**: Processamento de grandes volumes de dados (como scans de AST) deve ser feito via canais (Channels) para desacoplar leitura de escrita.
* **Standard #16 (CGO Memory Integrity)**: Ao utilizar bindings CGO (como Tree-sitter), é MANDATÓRIO o fechamento explícito de objetos via `defer obj.Close()`. Árvores e Cursores AST residem em memória não gerenciada e causam vazamentos sistêmicos se não liberados.

## 🗄️ Database & Consistency

* **Standard #03 (Atomic Transactions)**: Toda operação de escrita que envolva mais de um comando SQL ou que impacte o estado de um arquivo deve ser envolvida em uma transação (`db.BeginTx`).
* **Standard #04 (WAL Mode)**: O SQLite deve operar sempre em modo WAL para permitir alta concorrência entre workers de scan e leituras da UI.

## 🛡️ Error Governance

* **Standard #05 (Error Wrapping)**: Erros devem ser retornados com contexto utilizando `fmt.Errorf("contexto: %w", err)`. Nunca silencie erros com `_` a menos que seja um descarte intencional e documentado.
* **Standard #06 (Fail-Fast)**: Em operações concorrentes, utilize `errgroup` ou monitoramento de canais para interromper a execução imediatamente após a primeira falha crítica.

## 🚀 CLI & UX

* **Standard #07 (Command Isolation)**: A lógica de execução dos comandos não deve residir no `main.go`. Cada comando deve ter seu próprio arquivo no pacote `internal/commands` ou `cmd/sentinel/commands`.
* **Standard #09 (Clean Shutdown)**: Nunca use `log.Fatalf` ou `os.Exit` dentro de subcomandos. Use `RunE` para retornar o erro ao `Execute()` do Cobra, garantindo que os `PersistentPostRun` e `defers` de limpeza de banco sejam executados.

## 🔒 Security & Portability

* **Standard #10 (Shell-Less Execution)**: Evite `exec.Command("sh", "-c", ...)`. Use parsers de argumentos (como `shlex`) para invocar binários diretamente. Isso previne Command Injection e garante compatibilidade cross-platform (Windows/Linux).
* **Standard #11 (Explicit DB State)**: Nunca assuma o estado padrão do SQLite. Toda conexão deve explicitamente ativar `PRAGMA foreign_keys = ON` e configurar `busy_timeout` para evitar bloqueios em ambiente concorrente.
* **Standard #17 (Concurrency Sovereignty)**: Ferramentas que utilizam estado interno mutável ou memória C (como parsers) NÃO são thread-safe por padrão. Utilize `sync.Pool` para gerenciar instâncias de parsers em ambientes multi-worker, garantindo isolamento total de goroutine.

## 🏛️ Continuous Learning & Self-Correction

...

* **Standard #08 (The Sovereign Audit Framework)**: É MANDATÓRIO que toda conclusão de tarefa ou sprint seja acompanhada de um Relatório de Auditoria de 5 pontos:
    1. ✨ **The Good**: Vitórias técnicas e o que agora é sólido.
    2. ⚠️ **The Bad**: Dívidas técnicas ou hacks aceitos conscientemente.
    3. 💥 **The Ugly**: Fragilidades, riscos e inconsistências detectadas.
    4. 💡 **The Lesson**: O aprendizado universal ou novo padrão extraído.
    5. 🚀 **The Next**: O próximo passo de otimização ou evolução arquitetural.

* **Standard #13 (Verify, Never Assume)**: É PROIBIDO realizar afirmações de sucesso ou conclusão sem evidência física verificável. Toda resposta de "OK" deve vir acompanhada de um Log, Exit Code ou Prova de Estado.

* **Standard #14 (Proportional Documentation)**: A documentação deve ser atualizada de forma proporcional ao impacto da tarefa (Tiers):
  * **T1 (Trivial)**: Documentação opcional. O commit semântico é evidência suficiente.
  * **T2 (Feature)**: Atualização obrigatória do `sentinel-log.md`.
  * **T3 (Arquitetura)**: Atualização obrigatória de Logs, Evolution Insights e ADRs.
    Nenhuma mudança estrutural pode ser selada sem o registro de sua pegada intelectual.

* **Standard #12 (Deterministic ADRs)**: Todo Registro de Decisão Arquitetural (ADR) deve conter obrigatoriamente um **Protocolo de Verificação** (comando shell). Um ADR sem comando de prova é considerado uma "Dívida de Documentação" (T3).

* **Standard #15 (The Hard Gate Enforcement)**: O status `COMPLETED` de uma tarefa é um portão de ferro. Ele é bloqueado sistematicamente até que o comando contido no ADR vinculado retorne `Exit Code 0`. O progresso é medido por evidência física, não por intenção.

## 🧹 Code Quality & Linting

* **Standard #18 (Zero Linter Debt)**: Todo código enviado para `main` deve passar por `golangci-lint run` com zero issues. As verificações obrigatórias são:
  * `gocyclo` ≤ 15 para toda função — complexidade ciclomática alta é refatorada antes do merge.
  * `revive/exported` — todo símbolo exportado (type, func, const, method) deve ter doc comment no formato `// SymbolName verb phrase.`.
  * `noctx` — toda chamada de banco ou rede deve usar variante Context (QueryRowContext, QueryContext, etc.).
  * `errcheck` — nenhum erro pode ser descartado sem `//nolint` explícito e justificado.
  * `gosec` — permissões de arquivo 0600/0750 (G306), sem G304/G204 não justificados.
* **Standard #19 (Thread Safety)**: Todo código que acessa estado compartilhado (`map`, slice sem sincronização) deve ser protegido por `sync.RWMutex` ou `sync.Mutex`. A suíte `go test -race ./...` deve passar sem warnings. O padrão `sync.Map` é preferível em cenários de leitura-heavy com writes esparsos.

---
*Última Atualização: 2026-05-13*
