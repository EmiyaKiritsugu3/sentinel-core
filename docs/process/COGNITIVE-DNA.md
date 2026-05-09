# Sentinel Cognitive DNA — Modus Operandi & Evolution [PID-SENTINEL-DNA]

Este documento codifica o sistema de autoaprimoramento constante do Sentinel. Ele não rastreia bugs de código, mas sim **Falhas nas Linhas de Ação** e seus respectivos **Patches Cognitivos**.

## 🧬 Biblioteca de Anti-Padrões de Pensamento (AP)

| ID | Nome | Linha de Ação Falha | Por Que? (Motivação) |
|---|---|---|---|
| **[AP-01]** | Miopia de Superfície | Escolher a ferramenta mais rápida que resolve o caso óbvio (ex: `strings.Contains`). | Desejo de reduzir esforço de parsing e acelerar a entrega (Heuristic Bias). |
| **[AP-02]** | Inércia de Wiring | Confiar que dependências existem porque foram "fundeados" no Engine. | Assunção de que o sistema é um monólito de confiança, ignorando limites de borda (Defensive Gap). |
| **[AP-03]** | Colapso Binário | Reduzir lógica multi-fatorial a `if/else` ou switches simples. | Tentativa de simplificar o fluxo de decisão às custas da precisão incremental (Signal Loss). |

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

---

## 📈 Ciclo de Evolução (SEC)

1. **Ação:** Executar tarefa.
2. **Impacto:** Se houver erro ou correção externa (Review), identificar o AP correspondente.
3. **Mutação:** Se for um AP novo, adicioná-lo à biblioteca e criar um PMO.
4. **Verificação:** A próxima tarefa DEVE citar qual PMO está sendo aplicado preventivamente.

*"Quando encontrar um bom movimento, procure um melhor. Quando cometer um erro, mude a forma como você procura."*
