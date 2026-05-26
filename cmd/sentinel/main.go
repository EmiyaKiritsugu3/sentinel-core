package main

import (
    "context"
    "fmt"
    "os"

    "github.com/EmiyaKiritsugu3/sentinel-core/cmd/sentinel/commands"
    "github.com/EmiyaKiritsugu3/sentinel-core/internal/knowledge"
    "github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func main() {
    // 1. Initialize the Brain (SQLite)
    db, err := sqlite.Init()
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to initialize sentinel brain: %v\n", err)
        os.Exit(1)
    }
    defer func() { _ = db.Close() }()

    // 2. Auto-debrief on exit (runs before db.Close due to LIFO defer)
    defer func() {
        if knowledge.GlobalBuffer.Len() == 0 {
            return
        }
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return
        }
        svc := knowledge.NewDebriefService(knowledge.GlobalBuffer, db, homeDir+"/knowledge")
        if _, _, err := svc.Save(context.Background()); err != nil {
            fmt.Fprintf(os.Stderr, "auto-debrief failed: %v\n", err)
        }
    }()

    // 3. Execute the CLI with injected database
    commands.Execute(db)
}
