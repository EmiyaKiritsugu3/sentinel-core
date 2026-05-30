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

// createTasksTable creates the tasks table in the given DB for testing.
// Uses IF NOT EXISTS for idempotency across subtests.
func createTasksTable(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
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
}

func TestHandleGetStatus_NoTasks(t *testing.T) {
	t.Parallel()

	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	createTasksTable(t, rawDB)

	handler := handleGetStatus(db)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var status TaskStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if status.ID != "" {
		t.Errorf("expected empty task, got ID=%q", status.ID)
	}
	if status.Description != "" {
		t.Errorf("expected empty Description, got %q", status.Description)
	}
	if status.Status != "" {
		t.Errorf("expected empty Status, got %q", status.Status)
	}
	if status.Tier != nil {
		t.Errorf("expected Tier to be nil (omitted from JSON), got %v", *status.Tier)
	}
	if status.Verification != nil {
		t.Errorf("expected Verification to be nil (omitted from JSON), got %v", *status.Verification)
	}
	if status.CreatedAt != nil {
		t.Errorf("expected CreatedAt to be nil (omitted from JSON), got %v", *status.CreatedAt)
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
	createTasksTable(t, rawDB)

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
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
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
	if status.Description != "Add Auth Service" {
		t.Errorf("expected Description 'Add Auth Service', got %q", status.Description)
	}
	if status.Tier == nil || *status.Tier != "T1" {
		t.Errorf("expected Tier=T1, got %v", status.Tier)
	}
	if status.Verification == nil || *status.Verification != "go test ./internal/auth/..." {
		t.Errorf("expected Verification set, got %v", status.Verification)
	}
	if status.CreatedAt == nil || *status.CreatedAt == "" {
		t.Errorf("expected CreatedAt set by DEFAULT CURRENT_TIMESTAMP")
	}
}

func TestHandleGetStatus_DBError(t *testing.T) {
	t.Parallel()

	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	// Intentionally DO NOT create the tasks table.
	// QueryRow will fail with a SQL error (not ErrNoRows),
	// triggering the 500 branch in handleGetStatus.

	handler := handleGetStatus(db)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for missing table, got %d: body=%s", rec.Code, rec.Body.String())
	}

	// Verify the response Content-Type is still JSON
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json on error, got %q", ct)
	}

	// Verify the response body is valid JSON
	var errBody map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&errBody); err != nil {
		t.Fatalf("error response is not valid JSON: %v", err)
	}
	if errBody["error"] != "internal server error" {
		t.Errorf("expected error message, got %q", errBody["error"])
	}
}

func TestSetSecureCORS(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		expectAcao string
	}{
		{
			name:       "valid localhost",
			origin:     "http://localhost:3000",
			expectAcao: "http://localhost:3000",
		},
		{
			name:       "valid 127.0.0.1",
			origin:     "http://127.0.0.1:5173",
			expectAcao: "http://127.0.0.1:5173",
		},
		{
			name:       "valid ipv6 loopback",
			origin:     "http://[::1]:8080",
			expectAcao: "http://[::1]:8080",
		},
		{
			name:       "invalid external site",
			origin:     "https://malicious.com",
			expectAcao: "",
		},
		{
			name:       "invalid subdomain evasion",
			origin:     "http://localhost.malicious.com",
			expectAcao: "",
		},
		{
			name:       "empty origin",
			origin:     "",
			expectAcao: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/status", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := httptest.NewRecorder()

			setSecureCORS(rec, req)

			acao := rec.Header().Get("Access-Control-Allow-Origin")
			if acao != tc.expectAcao {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tc.expectAcao, acao)
			}
		})
	}
}
