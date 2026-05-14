# Sentinel Cognitive DNA — Modus Operandi & Evolution [PID-SENTINEL-DNA]

Este documento codifica o sistema de autoaprimoramento constante do Sentinel. Ele não rastreia bugs de código, mas sim **Falhas nas Linhas de Ação** e seus respectivos **Patches Cognitivos**.

## 🧬 Biblioteca de Anti-Padrões de Pensamento (AP)

| ID | Nome | Linha de Ação Falha | Por Que? (Motivação) |
|---|---|---|---|
| **[AP-01]** | Miopia de Superfície | Escolher a ferramenta mais rápida que resolve o caso óbvio (ex: `strings.Contains`). | Desejo de reduzir esforço de parsing e acelerar a entrega (Heuristic Bias). |
| **[AP-02]** | Inércia de Wiring | Confiar que dependências existem porque foram "fundeados" no Engine. | Assunção de que o sistema é um monólito de confiança, ignorando limites de borda (Defensive Gap). |
| **[AP-03]** | Colapso Binário | Reduzir lógica multi-fatorial a `if/else` ou switches simples. | Tentativa de simplificar o fluxo de decisão às custas da precisão incremental (Signal Loss). |
| **[AP-04]** | Anestesia do Silêncio | Ignorar erro de parsing/formação com `_` porque "a superfície funciona" (zero-value parece aceitável). | Desejo de evitar ruído em código que "parece" funcionar, ignorando que zero-value é dado corrompido (Invisibility Bias). |
| **[AP-05]** | Miragem do Sintoma Único | Assumir que um erro visível (exit code 3) tem causa única, quando na verdade é cascata de falhas. | Pressa de resolver o CI urgente, assumindo causalidade única onde há encadeamento de falhas (Stack Blindness). |
| **[AP-06]** | Inércia de Ramificação | Manter branches e worktrees ativos após o merge do código. | Preguiça de limpar — "não atrapalha" até alguém tentar usar o branch (Hygiene Debt). |

## 🛠️ Patches de Modus Operandi (PMO)

Estes patches alteram como o Sentinel "pensa" e "age" a partir de agora.

### PMO-01: O Teste do Falso Positivo (Anti-AP-01)
- **Regra:** Proibido implementar lógica de busca/match sem listar 3 casos que NÃO devem dar match.
- **Modus Operandi:** O Sentinel agora desconfia de substrings por padrão. A ferramenta padrão para intenções é a **Tokenização + Exact Match**.

### PMO-02: A Regra do Componente Hostil (Anti-AP-02)
- **Regra:** Todo método exportado/público deve validar `nil` em suas dependências injetadas.
- **Modus Operandi:** Nenhum componente é considerado "amigo". A segurança deve ser intrínseca e isolada (Sovereignty).
- **Implementation:** `sqlite.ErrNilDB` sentinel error + `ValidateDB` function (`pkg/sqlite/validation.go`). All exported methods now systematically validate DB dependency via `fmt.Errorf("%s: %w", caller, ErrNilDB)`.

### PMO-03: O Padrão Acumulador (Anti-AP-03)
- **Regra:** Lógica com >2 fatores de decisão DEVE usar pipelines de pesos aditivos.
- **Modus Operandi:** Evitar "Switches" de decisão. O Sentinel agora constrói resultados por acumulação de evidências.

### PMO-04: O Erro Nunca é Opcional (Anti-AP-04)
- **Regra:** Todo `time.Parse` ou operação de scanning que retorne `(T, error)` DEVE ter o erro tratado, não silenciado com `_`. A única exceção é quando o zero-value for semanticamente válido E documentado com justificativa inline.
- **Modus Operandi:** O Sentinel agora trata erro de parsing como corrupção de dado, não como ruído tolerável. Nenhum `_` em operações de conversão é aceito sem justificativa explícita.
- **Implementation:** `internal/state/manager.go` — erro de `time.Parse` propagado com `fmt.Errorf` em vez de silenciado.

### PMO-05: Isolamento de Camadas (Anti-AP-05)
- **Regra:** Ao depurar falha em pipeline (CI, build, deploy), listar EXPLICITAMENTE todos os layers na ordem de execução e testar CADA layer individualmente ANTES de tentar consertar o sintoma visível.
- **Modus Operandi:** O Sentinel não aceita "exit code N" como diagnóstico isolado — exige verificação de cada etapa do pipeline de baixo para cima.
- **Template de debug CI:** (1) `go build` (2) `go test` (3) `go test -race` (4) coverage file existe? (5) scanner CLI roda manualmente?

### PMO-06: Higiene de Branches (Anti-AP-06)
- **Regra:** Após o merge de um branch (especialmente squash merge), deletar branch local + remote IMEDIATAMENTE e limpar worktrees associados. Branches com mais de 7 dias sem atividade entram no backlog de limpeza automática.
- **Modus Operandi:** O Sentinel considera branches não deletados após merge como "dívida técnica de processo" — acumulam sujeira (arquivos não commitados, conflitos latentes, docs desatualizados).

---

## 📈 Ciclo de Evolução (SEC)

1. **Ação:** Executar tarefa.
2. **Impacto:** Se houver erro ou correção externa (Review), identificar o AP correspondente.
3. **Mutação:** Se for um AP novo, adicioná-lo à biblioteca e criar um PMO.
4. **Verificação:** A próxima tarefa DEVE citar qual PMO está sendo aplicado preventivamente.

*"Quando encontrar um bom movimento, procure um melhor. Quando cometer um erro, mude a forma como você procura."*
