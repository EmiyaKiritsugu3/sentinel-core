# Design Process Standards [PID-SENTINEL]

Regras de processo para design de features no Sentinel. Aplicam-se a qualquer nova feature ou subsistema, independente de escopo.

---

## Audit Depth Rule

Antes de implementar, determine quantos rounds de auditoria são necessários com base em dois eixos:

```
              Fácil reverter   Difícil reverter
Baixo impacto    0 rounds          1 round
Médio impacto    1 round           2 rounds
Alto impacto     1 round           2 rounds + sign-off
```

**Blast Radius (impacto):**
- Baixo: 1 arquivo, 1 flag, 1 função isolada
- Médio: 1 pacote, 1 interface pública
- Alto: padrão usado em todo o codebase, schema de DB, contrato entre pacotes

**Reversibilidade:**
- Fácil: rename, lógica interna, flags
- Difícil: schema de DB, interfaces públicas, decisões arquiteturais

**Condição de parada absoluta** (sobrescreve a matriz):
Se audit round N encontrou 0 issues críticas E 0 issues major → para imediatamente.

**Classificação de issues:**
- **Crítica:** race condition, security flaw, breaking change sem migração, modelo de dados errado
- **Major:** package boundary incorreto, caminho de erro ausente, API que não consegue evoluir
- **Minor:** naming, style, preferência — nunca justifica outro round

---

## Time Budget Rule

Se o tempo de design exceder **30% do tempo estimado de implementação** → commitar o design e implementar.

A implementação revela problemas mais rápido que discussão adicional de design. Auditar além do ponto de retorno decrescente é "Infinite Optimization" — o projeto não avança.

**Sinal de alerta:** se você está auditando a auditoria, ou buscando problemas em detalhes que testes unitários cobrem trivialmente, o budget foi excedido.

---

## Formato de Spec

Todo spec vai para `docs/superpowers/specs/YYYY-MM-DD-<feature>-design.md`.

Estrutura mínima:
1. **Problem** — o que está quebrado ou faltando
2. **Scope** — in scope / out of scope explícitos
3. **Architecture** — diagrama ASCII com fluxo temporal separado por pontos de execução
4. **Components** — interfaces, structs, assinaturas de funções
5. **Error handling** — tabela de cenários de falha com comportamento definido
6. **Known limitations** — o que foi conscientemente deixado de fora do MVP
7. **Verification gate** — comandos exatos que devem passar antes do PR

Diagramas: ASCII para specs de trabalho (legível em terminal), Mermaid para docs publicadas em `docs/architecture/`.

---

## Quando NÃO auditar

- Mudanças pontuais em código existente (bug fix de 1 arquivo)
- Configuração ou boilerplate
- Renomeação sem mudança de comportamento
- Features com blast radius baixo E fáceis de reverter

A auditoria é uma ferramenta, não um ritual. Aplicar indiscriminadamente dilui o valor.

---

*Origem: S19 — Prompt Intelligence Design Session (2026-05-04)*
*Atualizado em: 2026-05-05*
