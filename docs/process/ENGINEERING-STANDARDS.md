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

---
*Última Atualização: 2026-04-26*
