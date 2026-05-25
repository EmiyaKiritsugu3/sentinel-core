# internal/reflect

Structural validation and compliance checking against Sentinel engineering standards.

## Overview

The reflect package provides a `Validator` that scans project files for standards violations and validates agent-provided paths and commands against security rules. It implements two hard gates: path traversal prevention (Gate A argument validation) and shell injection prevention.

## Key Types

### `Validator`
Holds a `*sqlite.DB` reference for standards lookup.
- `NewValidator(db)` — constructor with DB validation
- `ValidateProject(root)` — walks project directory checking `.go` files against standards
- `ValidatePath(path)` — blocks absolute paths and `..` traversal
- `ValidateCommand(cmd)` — blocks shell metacharacters (`|`, `&&`, `;`, `>`, `>>`, `<`, `` ` ``, `$(`)

### `Violation`
Represents a standards violation: `StandardID`, `FilePath`, `Line`, `Reason`.

### `checkFile(path)`
Internal scanner that detects:
- **STD-01**: Use of `os.ReadFile` (should use `bufio`)
- **STD-05**: Raw `return nil, err` (should use `fmt.Errorf` wrapping)

Skips `legacy/` directories and `.go`-only files. Self-referential checking is disabled for the validator file itself (`internal/reflect/validator.go`).

## Security Enforcement

Both `ValidatePath` and `ValidateCommand` are called from `Engine.executeToolsWithResults` in the agents package as **Hard Gates** before any tool executes:

```go
case "path", "file", "filepath":
    if err := e.validator.ValidatePath(strVal); err != nil {
        return fmt.Errorf("hard gate: %w", err)
    }
case "command", "cmd":
    if err := e.validator.ValidateCommand(strVal); err != nil {
        return fmt.Errorf("hard gate: %w", err)
    }
```

## Dependencies

- `pkg/sqlite` — DB validation

## Usage

```go
v, _ := reflect.NewValidator(db)

// Project audit
violations, err := v.ValidateProject(".")

// Path validation (Gate A)
if err := v.ValidatePath(agentSuppliedPath); err != nil {
    // Block: path traversal attempt
}

// Command validation
if err := v.ValidateCommand("rm -rf /"); err != nil {
    // Block: forbidden characters
}
```
