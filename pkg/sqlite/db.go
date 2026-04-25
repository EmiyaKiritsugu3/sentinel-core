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

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping sqlite db: %w", err)
	}

	return &DB{Conn: db}, nil
}

// Close fecha a conexão com o banco
func (db *DB) Close() error {
	return db.Conn.Close()
}
