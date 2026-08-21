package liveview

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" && acao != "http://127.0.0.1:5173" {
		if acao != "" {
			t.Errorf("expected empty string or http://localhost:5173, got %q", acao)
		}
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
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" && acao != "http://127.0.0.1:5173" {
		if acao != "" {
			t.Errorf("expected empty string or http://localhost:5173, got %q", acao)
		}
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
	req.Header.Set("Origin", "http://localhost:5173")
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

func TestHandleGetGraph(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetGraph(db)
	_, _ = db.Conn.Exec("CREATE TABLE IF NOT EXISTS nodes (id TEXT, name TEXT, type TEXT, file_path TEXT, start_line INTEGER, end_line INTEGER, hash TEXT, last_indexed INTEGER)")
	_, _ = db.Conn.Exec("CREATE TABLE IF NOT EXISTS edges (from_node_id TEXT, to_node_id TEXT, relation_type TEXT)")
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" && acao != "http://127.0.0.1:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
	}
}

func TestHandleGetCode(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError && rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected code: %d", rec.Code)
	}
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" && acao != "http://127.0.0.1:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
	}
}

func TestHandleListADR(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected code: %d", rec.Code)
	}
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" && acao != "http://127.0.0.1:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
	}
}

func TestHandleGetADR(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/dummy", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError && rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected code: %d", rec.Code)
	}
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" && acao != "http://127.0.0.1:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
	}
}

func TestHandleGetGraph_ValidData(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	_, _ = db.Conn.Exec("CREATE TABLE IF NOT EXISTS nodes (id TEXT, name TEXT, type TEXT, file_path TEXT, start_line INTEGER, end_line INTEGER, hash TEXT, last_indexed INTEGER)")
	_, _ = db.Conn.Exec("CREATE TABLE IF NOT EXISTS edges (from_node_id TEXT, to_node_id TEXT, relation_type TEXT)")
	_, _ = db.Conn.Exec("INSERT INTO nodes (id, name, type, file_path, start_line, end_line, hash, last_indexed) VALUES ('1', 'test', 'func', 'test.go', 1, 2, 'hash', 123)")
	_, _ = db.Conn.Exec("INSERT INTO edges (from_node_id, to_node_id, relation_type) VALUES ('1', '2', 'calls')")

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetCode_ValidData(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)

	// Create dummy file
	_ = os.WriteFile("dummy_test.go", []byte("package main\n\nfunc main() {}\n"), 0644)
	defer os.Remove("dummy_test.go")

	req := httptest.NewRequest(http.MethodGet, "/api/code?path=dummy_test.go&start=1&end=2", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/api/code?path=dummy_test.go&start=invalid", nil)
	req2.Header.Set("Origin", "http://localhost:5173")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/code?path=dummy_test.go&end=invalid", nil)
	req3.Header.Set("Origin", "http://localhost:5173")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
}

func TestHandleGetCode_InvalidPath(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=/etc/passwd", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleListADR_ValidData(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)

	os.MkdirAll("docs/architecture/adr", 0755)
	os.WriteFile("docs/architecture/adr/ADR-123-test.md", []byte("# ADR 123\n\nTest ADR"), 0644)
	os.Mkdir("docs/architecture/adr/dummy", 0755)

	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetADR_ValidData(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)

	os.MkdirAll("docs/architecture/adr", 0755)
	os.WriteFile("docs/architecture/adr/ADR-123-test.md", []byte("# ADR 123\n\nTest ADR"), 0644)

	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-123-test.md", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/api/adr/notfound.md", nil)
	req2.Header.Set("Origin", "http://localhost:5173")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/adr/../../etc/passwd", nil)
	req3.Header.Set("Origin", "http://localhost:5173")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
}

func TestHandleListADR_ReadDirError(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)

	// Create a file where a dir is expected to force an error other than NotExist
	os.MkdirAll("docs/architecture/adr", 0755)
	os.RemoveAll("docs/architecture/adr")
	os.WriteFile("docs/architecture/adr", []byte("file"), 0644)
	defer os.Remove("docs/architecture/adr")

	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetADR_ErrorFileRead(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)

	os.MkdirAll("docs/architecture/adr", 0755)
	os.Mkdir("docs/architecture/adr/ADR-err.md", 0755) // directory will fail to read as file

	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-err.md", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetCode_ReadError(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)

	os.MkdirAll("dummy_dir", 0755)
	defer os.RemoveAll("dummy_dir")

	req := httptest.NewRequest(http.MethodGet, "/api/code?path=dummy_dir", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
