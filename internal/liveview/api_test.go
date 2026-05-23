package liveview

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestHandleGetStatus_NoTasks(t *testing.T) {
	t.Parallel()

	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	// Run graph schema to create tasks table.
	_, err = db.Conn.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			description TEXT NOT NULL,
			status TEXT NOT NULL,
			tier TEXT,
			verification_command TEXT,
			commit_hash TEXT,
			latency_ms REAL DEFAULT 0,
			tokens_used INTEGER DEFAULT 0,
			api_cost REAL DEFAULT 0,
			math_delta REAL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("create tasks table: %v", err)
	}

	handler := handleGetStatus(db)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: body=%s", rec.Code, rec.Body.String())
	}

	var status TaskStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if status.ID != "" {
		t.Errorf("expected empty task, got ID=%q", status.ID)
	}
}

func TestHandleGetStatus_WithTask(t *testing.T) {
	t.Parallel()

	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	ctx := context.Background()
	_, err = db.Conn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			description TEXT NOT NULL,
			status TEXT NOT NULL,
			tier TEXT,
			verification_command TEXT,
			commit_hash TEXT,
			latency_ms REAL DEFAULT 0,
			tokens_used INTEGER DEFAULT 0,
			api_cost REAL DEFAULT 0,
			math_delta REAL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("create tasks table: %v", err)
	}

	_, err = db.Conn.ExecContext(ctx,
		"INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)",
		"task-1", "Add Auth Service", "IN_PROGRESS", "T1", "go test ./internal/auth/...",
	)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	handler := handleGetStatus(db)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var status TaskStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if status.ID != "task-1" {
		t.Errorf("expected ID task-1, got %q", status.ID)
	}
	if status.Status != "IN_PROGRESS" {
		t.Errorf("expected IN_PROGRESS, got %q", status.Status)
	}
}
