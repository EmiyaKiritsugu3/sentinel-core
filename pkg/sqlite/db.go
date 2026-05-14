// Package sqlite provides the SQLite database wrapper and initialization.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a database/sql connection.
type DB struct {
	Conn *sql.DB
}

// Init establishes SQLite connection and configures Elite Pragmas
func Init() (*DB, error) {
	return InitAtPath(".sentinel/graph.db")
}

// InitAtPath establishes SQLite connection at a specific path
func InitAtPath(dbPath string) (*DB, error) {
	sentinelDir := filepath.Dir(dbPath)
	if _, err := os.Stat(sentinelDir); os.IsNotExist(err) && sentinelDir != "." {
		err := os.MkdirAll(sentinelDir, 0750)
		if err != nil {
			return nil, fmt.Errorf("sqlite: could not create directory %s: %w", sentinelDir, err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("sqlite: could not open db: %w", err)
	}

	// Pragma Configuration for Performance and Integrity
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = NORMAL;",
	}

	ctx := context.Background()

	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			return nil, fmt.Errorf("sqlite: failed to apply pragma %s: %w", p, err)
		}
	}

	// Pool Configuration for Concurrency (WAL allows multiple readers)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("sqlite: could not ping db: %w", err)
	}

	return &DB{Conn: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.Conn.Close()
}
