# Wiki: AST Language Expansion (Tree-sitter) [PID-SENTINEL-AST]

Este Wiki detalha a infraestrutura de análise multi-linguagem do Sentinel, focando no motor Tree-sitter e padrões de extração semântica.

## 1. Visão Geral (The Multi-Language Orchestrator)

O Sentinel utiliza uma arquitetura de `FileScanners` desacoplados, onde cada linguagem possui seu driver especializado. O motor central (`Engine.go`) coordena o Worker Pool concorrente e a persistência no SQLite.

## 2. Padrões Tree-sitter (Standard #16 & #17)

A implementação de novos scanners deve seguir o padrão "Elite" de gerenciamento de recursos CGO:

### Gerenciamento de Memória (CGO Safe)

Árvores AST e Cursores são alocados em C. É obrigatório liberar esses recursos:

```go
tree := parser.Parse(source, nil)
defer tree.Close() // MANDATÓRIO

cursor := sitter.NewQueryCursor()
defer cursor.Close() // MANDATÓRIO
```

### Concorrência Soberana (sync.Pool)

Instâncias de `sitter.Parser` não são thread-safe. Para scans paralelos, utilize o pool global:

```go
parser := s.pool.Get().(*sitter.Parser)
defer s.pool.Put(parser)
```

## 3. Extração Semântica (S-expressions)

Utilizamos `sitter.Query` para capturar símbolos de forma declarativa. O query principal para TypeScript/TSX captura:

- **Imports**: `(import_statement source: (string) @import.path)`
- **Interfaces**: `(interface_declaration name: (identifier) @interface.name)`
- **Classes**: `(class_declaration name: (identifier) @class.name)`
- **Componentes React**: Capturados via `lexical_declaration` ou `function_declaration` utilizando a heurística de nomenclatura PascalCase.

## 4. Heurísticas de Componentes

Um símbolo é classificado como `component` se:

1. O nó AST for uma função ou declaração de variável.
2. O nome iniciar com letra Maiúscula (PascalCase).
3. (Futuro) Identificação de retorno de JSX/TSX.

## 5. Known Landmines

- **Relative Paths**: O extrator de imports captura o literal (ex: `./Component`). O `Linker` deve resolver estes caminhos para o ID canônico do arquivo no `graph.db`.
- **CGO Compilation**: O binário exige `CGO_ENABLED=1`. Verifique o ambiente de build para evitar erros de linkagem estática.

---
*Assinado: Sovereign Architect*
