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
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin, got %q", acao)
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
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin, got %q", acao)
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

func TestSetCORS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		origin         string
		expectedHeader string
	}{
		{"no origin", "", ""},
		{"valid localhost", "http://localhost:5173", "http://localhost:5173"},
		{"valid 127.0.0.1", "http://127.0.0.1:8080", "http://127.0.0.1:8080"},
		{"invalid origin", "http://example.com", ""},
		{"malformed origin", "http://%42:8080", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			setCORS(rec, req)

			acao := rec.Header().Get("Access-Control-Allow-Origin")
			if acao != tt.expectedHeader {
				t.Errorf("expected %q, got %q", tt.expectedHeader, acao)
			}
		})
	}
}

func TestHandleGetGraph_Full(t *testing.T) {
	t.Parallel()
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	_, err = db.Conn.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			name TEXT,
			type TEXT,
			file_path TEXT,
			start_line INTEGER,
			end_line INTEGER,
			hash TEXT,
			last_indexed TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("create nodes table: %v", err)
	}
	_, err = db.Conn.ExecContext(context.Background(), `
		INSERT INTO nodes (id, name, type, file_path, start_line, end_line, hash, last_indexed)
		VALUES ('node1', 'Node 1', 'file', 'test.go', 1, 10, 'hash', '2023-01-01T00:00:00Z')
	`)
	if err != nil {
		t.Fatalf("insert nodes: %v", err)
	}

	_, err = db.Conn.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS edges (
			from_node_id TEXT,
			to_node_id TEXT,
			relation_type TEXT,
			PRIMARY KEY (from_node_id, to_node_id, relation_type)
		)
	`)
	if err != nil {
		t.Fatalf("create edges table: %v", err)
	}
	_, err = db.Conn.ExecContext(context.Background(), `
		INSERT INTO edges (from_node_id, to_node_id, relation_type)
		VALUES ('node1', 'node2', 'calls')
	`)
	if err != nil {
		t.Fatalf("insert edges: %v", err)
	}

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleListADR_Full(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetCode_MissingPath(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing path, got %d", rec.Code)
	}
}

func TestHandleGetCode_InvalidPath(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=../secret", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid path, got %d", rec.Code)
	}
}

func TestHandleGetADR_MissingPath(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing path, got %d", rec.Code)
	}
}

func TestHandleGetCode_FileFound(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetADR_FileFound(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-test-1.md", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetCode_StartEndParams(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=1&end=5", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=abc", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&end=abc", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetCode_InvalidEnd(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleGetCode(db)

	req := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&end=9999", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&end=-1", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
func TestHandleGetCode_StartEndBounds(t *testing.T) {
	t.Parallel()
	db := &sqlite.DB{}
	handler := handleGetCode(db)

	req := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=-1&end=0", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=9999", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
func TestHandleListADR_ValidDir(t *testing.T) {
	t.Parallel()

	db := &sqlite.DB{}
	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
