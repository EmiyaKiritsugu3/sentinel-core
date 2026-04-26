package main

import (
	"fmt"
	"os"

	"github.com/EmiyaKiritsugu3/sentinel-core/cmd/sentinel/commands"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func main() {
	// 1. Inicializa o Cérebro (SQLite)
	db, err := sqlite.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize sentinel brain: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Executa o CLI injetando o banco
	commands.Execute(db)
}
