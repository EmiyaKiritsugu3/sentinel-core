# internal/registry

Thread-safe global command registry for the Sentinel CLI.

## Overview

The registry package provides a simple global registration system for `cobra.Command` factories. Each subcommand package registers its factory via `init()` and the CLI root command collects them at startup.

## Key Types

### `CommandFactory`
```go
type CommandFactory func(*sqlite.DB) *cobra.Command
```
A factory function that receives the database handle and returns a configured `*cobra.Command`. This pattern enables each command to independently configure its own subcommands, flags, and runners without importing other commands.

### `Register(factory CommandFactory)`
Thread-safe registration (`sync.Mutex`). Called from package-level `init()` functions in each command file. Registration order is preserved (appended to slice).

### `GetCommands() []CommandFactory`
Returns a defensive copy of all registered factories. Used by `NewRootCmd` in `cmd/sentinel/commands/root.go` to dynamically build the complete CLI tree.

### `ResetForTesting()`
Clears the global registry for test isolation. Should only be called in `*_test.go` files.

## Dependencies

- `pkg/sqlite` — database handle passed to factories
- `github.com/spf13/cobra` — CLI framework

## Usage

```go
// In a command file (e.g., internal/x/command.go):
package x

import (
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
    "github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
    "github.com/spf13/cobra"
)

func init() {
    registry.Register(func(db *sqlite.DB) *cobra.Command {
        return &cobra.Command{
            Use:   "mycmd",
            Short: "My command",
            RunE: func(cmd *cobra.Command, args []string) error {
                // use db
                return nil
            },
        }
    })
}
```

## Registered Commands

The following packages register commands via `init()`: `debrief`, `scan`, `plan`, `audit`, `status`, `start`, `instruct`, `live`, `visualize`, `report`, `pattern`.
