package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	Conn *sql.DB
}

// Init inicializa a conexão com o SQLite e cria as pastas necessárias
func Init() (*DB, error) {
	sentinelDir := ".sentinel"
	if _, err := os.Stat(sentinelDir); os.IsNotExist(err) {
		err := os.Mkdir(sentinelDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("could not create .sentinel directory: %w", err)
		}
	}

	dbPath := filepath.Join(sentinelDir, "graph.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open sqlite db: %w", err)
	}

	// 1. Habilita o modo WAL para permitir leitura simultânea com escrita
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// 2. Configura os limites do Pool para evitar contenção no SQLite
	db.SetMaxOpenConns(1) // SQLite é single-writer, garantimos 1 conexão para escrita atômica

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping sqlite db: %w", err)
	}

	return &DB{Conn: db}, nil
}

// Close fecha a conexão com o banco
func (db *DB) Close() error {
	return db.Conn.Close()
}
