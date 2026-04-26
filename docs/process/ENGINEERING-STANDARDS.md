# Sentinel Engineering Standards [PID-SENTINEL]

Este documento define os padrões técnicos inegociáveis do Sentinel Core. Toda nova função ou módulo deve aderir a estes padrões desde a primeira linha.

## 💾 I/O & Memory Management
*   **Standard #01 (Buffered Reads)**: Nunca utilize `os.ReadFile` para leitura de arquivos de código ou logs. Utilize sempre `bufio.Scanner` com buffer configurado para 1MB para garantir eficiência de memória.
*   **Standard #02 (Stream Processing)**: Processamento de grandes volumes de dados (como scans de AST) deve ser feito via canais (Channels) para desacoplar leitura de escrita.

## 🗄️ Database & Consistency
*   **Standard #03 (Atomic Transactions)**: Toda operação de escrita que envolva mais de um comando SQL ou que impacte o estado de um arquivo deve ser envolvida em uma transação (`db.BeginTx`).
*   **Standard #04 (WAL Mode)**: O SQLite deve operar sempre em modo WAL para permitir alta concorrência entre workers de scan e leituras da UI.
## 🛡️ Error Governance
*   **Standard #05 (Error Wrapping)**: Erros devem ser retornados com contexto utilizando `fmt.Errorf("contexto: %w", err)`. Nunca silencie erros com `_` a menos que seja um descarte intencional e documentado.
*   **Standard #06 (Fail-Fast)**: Em operações concorrentes, utilize `errgroup` ou monitoramento de canais para interromper a execução imediatamente após a primeira falha crítica.

## 🚀 CLI & UX
*   **Standard #07 (Command Isolation)**: A lógica de execução dos comandos não deve residir no `main.go`. Cada comando deve ter seu próprio arquivo no pacote `internal/commands` ou `cmd/sentinel/commands`.
*   **Standard #09 (Clean Shutdown)**: Nunca use `log.Fatalf` ou `os.Exit` dentro de subcomandos. Use `RunE` para retornar o erro ao `Execute()` do Cobra, garantindo que os `PersistentPostRun` e `defers` de limpeza de banco sejam executados.

## 🔒 Security & Portability
*   **Standard #10 (Shell-Less Execution)**: Evite `exec.Command("sh", "-c", ...)`. Use parsers de argumentos (como `shlex`) para invocar binários diretamente. Isso previne Command Injection e garante compatibilidade cross-platform (Windows/Linux).
*   **Standard #11 (Explicit DB State)**: Nunca assuma o estado padrão do SQLite. Toda conexão deve explicitamente ativar `PRAGMA foreign_keys = ON` e configurar `busy_timeout` para evitar bloqueios em ambiente concorrente.

## 🏛️ Continuous Learning & Self-Correction
...

*   **Standard #08 (The Sovereign Audit Framework)**: É MANDATÓRIO que toda conclusão de tarefa ou sprint seja acompanhada de um Relatório de Auditoria de 5 pontos:
    1. ✨ **The Good**: Vitórias técnicas e o que agora é sólido.
    2. ⚠️ **The Bad**: Dívidas técnicas ou hacks aceitos conscientemente.
    3. 💥 **The Ugly**: Fragilidades, riscos e inconsistências detectadas.
    4. 💡 **The Lesson**: O aprendizado universal ou novo padrão extraído.
    5. 🚀 **The Next**: O próximo passo de otimização ou evolução arquitetural.

*   **Standard #13 (Verify, Never Assume)**: É PROIBIDO realizar afirmações de sucesso ou conclusão sem evidência física verificável. Toda resposta de "OK" deve vir acompanhada de um Log, Exit Code ou Prova de Estado.

*   **Standard #14 (Just-In-Time Documentation)**: A documentação (sentinel-log, evolution-insights, ADRs) deve ser atualizada OBRIGATORIAMENTE ao final de cada tarefa individual, e não apenas ao final de sprints ou sessões. Nenhuma tarefa é considerada concluída enquanto sua pegada intelectual não estiver persistida no repositório.

---
*Última Atualização: 2026-04-26*
