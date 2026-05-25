---
task_id: "ec002a3"
title: "DebriefService text/template Approach"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a3: DebriefService text/template Approach

## Contexto

Após cada sessão de agente, o Sentinel gera um arquivo Markdown de debrief contendo decisões tomadas, erros encontrados, padrões observados e arquivos modificados. O formato precisa ser consistente, extensível e de fácil leitura humana. A geração deve ser determinística — duas chamadas com o mesmo buffer devem produzir saída idêntica.

## Decisão

Utilizamos `text/template` da biblioteca padrão Go como motor de renderização do debrief. O template é definido como constante `debriefTemplate` no arquivo `debrief.go`, contendo seções fixas (`## Decisions Made`, `## Patterns Observed`, `### Anti-Patterns`, `### Success Patterns`, `## Files Changed`, `## Domain Tags`, `## Follow-ups`) com iteração sobre slices via `{{range}}`.

`DebriefService.Generate()` popula a struct `DebriefData` extraindo eventos do buffer por tipo (`Decisions()`, `Errors()`, `Patterns()`, `ByType(EventFileChange)`) e computa domínios únicos ordenados alfabeticamente via `sort.Strings` — garantindo determinismo na seção `## Domain Tags`.

Fallback: se o parsing ou execução do template falhar (ex.: erro de sintaxe na constante), o método `renderFallback` gera Markdown equivalente usando `strings.Builder` com `fmt.Sprintf`, garantindo que o debrief nunca falhe silenciosamente. O erro do template é prefixado como comentário HTML (`<!-- template error: ... -->`), preservando a rastreabilidade sem corromper a renderização Markdown.

## Consequências

- **Positivo**: Template declarativo facilita manutenção e customização do formato de saída sem alterar lógica de negócio.
- **Positivo**: Fallback determinístico garante geração de debrief mesmo sob falha de template.
- **Positivo**: `sort.Strings` nos domínios garante ordem alfabética estável — outputs idênticos para mesmos dados (verificado por `TestDebriefService_Generate_DomainsDeterministic`).
- **Negativo**: `text/template` não oferece auto-escaping HTML — irrelevante para saída Markdown, mas exigiria `html/template` se o output fosse servido como HTML.

## Referências

- Task ID: [ec002a3]
- Implementação: `internal/knowledge/debrief.go`
