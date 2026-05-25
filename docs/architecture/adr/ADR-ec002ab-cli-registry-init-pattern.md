---
task_id: "ec002ab"
title: "CLI registry.Register + init() Pattern"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002ab: CLI registry.Register + init() Pattern

## Contexto

O CLI do Sentinel (`cmd/sentinel/`) possui múltiplos subcomandos (`plan`, `scan`, `audit`, `live`, `debrief`, `start`, `status`, `visualize`, `report`, `pattern`, `instruct`), cada um implementado em seu próprio arquivo no pacote `commands`. Registrar manualmente cada comando no `root.go` criaria acoplamento e exigiria modificação do arquivo raiz a cada novo comando.

## Decisão

Implementamos o padrão **Registry + init() + Factory** com três componentes:

**Registry** (`internal/registry/commands.go`):
- `CommandFactory func(*sqlite.DB) *cobra.Command` — tipo de fábrica que recebe conexão DB e retorna comando.
- `Register(factory CommandFactory)` — adiciona fábrica ao slice global com `sync.Mutex`.
- `GetCommands() []CommandFactory` — retorna cópia defensiva do slice.

**init() registration** — cada arquivo de comando declara `func init() { registry.Register(NewXxxCmd) }`. Exemplo (`live.go`):
```go
func init() {
    registry.Register(NewLiveCmd)
}
```

**Root aggregation** (`root.go`):
```go
for _, factory := range registry.GetCommands() {
    root.AddCommand(factory(db))
}
```

Este padrão permite que novos comandos sejam adicionados simplesmente criando um novo arquivo com `init()` — zero modificações no `root.go`. A injeção de dependência (`*sqlite.DB`) é resolvida pela factory no momento da construção do comando, não no registro.

Thread-safety: `sync.Mutex` protege `Register` e `GetCommands` para permitir registro dinâmico em plugins futuros. Testes usam `ResetForTesting()` para limpar o registro global entre casos de teste.

## Consequências

- **Positivo**: Open/Closed Principle — comandos são adicionados sem modificar código existente.
- **Positivo**: Factory pattern permite injeção tardia de dependências (DB) no momento da construção do comando.
- **Positivo**: `GetCommands` retorna cópia defensiva, prevenindo race conditions entre iteração e registro.
- **Negativo**: `init()` é execução implícita — ordem de registro depende da ordem de carregamento de arquivos pelo compilador Go. Como comandos são independentes entre si, a ordem não afeta o comportamento.

## Referências

- Task ID: [ec002ab]
- Implementação: `internal/registry/commands.go`, `cmd/sentinel/commands/live.go:14-16`
