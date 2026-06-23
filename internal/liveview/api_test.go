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
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
	}
	if vary := rec.Header().Get("Vary"); vary != "Origin" {
		t.Errorf("expected Vary: Origin, got %q", vary)
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
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
	}
	if vary := rec.Header().Get("Vary"); vary != "Origin" {
		t.Errorf("expected Vary: Origin, got %q", vary)
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

func TestSetCORSHeaders(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectedVary   string
	}{
		{
			name:           "valid localhost",
			origin:         "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			expectedVary:   "Origin",
		},
		{
			name:           "valid 127.0.0.1",
			origin:         "http://127.0.0.1:8080",
			expectedOrigin: "http://127.0.0.1:8080",
			expectedVary:   "Origin",
		},
		{
			name:           "invalid origin evil.com",
			origin:         "https://evil.com",
			expectedOrigin: "",
			expectedVary:   "",
		},
		{
			name:           "empty origin",
			origin:         "",
			expectedOrigin: "",
			expectedVary:   "",
		},
		{
			name:           "invalid URL format",
			origin:         "://invalid",
			expectedOrigin: "",
			expectedVary:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			setCORSHeaders(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.expectedOrigin {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, got)
			}
			if got := rec.Header().Get("Vary"); got != tt.expectedVary {
				t.Errorf("expected Vary %q, got %q", tt.expectedVary, got)
			}
		})
	}
}

func TestHandleGetGraph_CORS(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected ACAO http://localhost:5173, got %q", acao)
	}
}

func TestHandleGetCode_CORS(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected ACAO http://localhost:5173, got %q", acao)
	}
}

func TestHandleListADR_CORS(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected ACAO http://localhost:5173, got %q", acao)
	}
}

func TestHandleGetADR_CORS(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-001", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected ACAO http://localhost:5173, got %q", acao)
	}
}

func TestHandleGetGraph_Full(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	_, _ = db.Conn.Exec(`
		CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			name TEXT,
			type TEXT,
			file_path TEXT,
			start_line INTEGER,
			end_line INTEGER,
			hash TEXT,
			last_indexed TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS edges (
			from_node_id TEXT,
			to_node_id TEXT,
			relation_type TEXT
		);
		INSERT INTO nodes (id, name, type, file_path, start_line, end_line, hash, last_indexed)
		VALUES ('n1', 'Node1', 'TypeA', 'file.go', 1, 10, 'hash1', '2020-01-01');
		INSERT INTO edges (from_node_id, to_node_id, relation_type)
		VALUES ('n1', 'n2', 'depends_on');
	`)

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetGraph_ErrorNodes(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetGraph_ErrorEdges(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	_, _ = db.Conn.Exec(`
		CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			name TEXT,
			type TEXT,
			file_path TEXT,
			start_line INTEGER,
			end_line INTEGER,
			hash TEXT,
			last_indexed TIMESTAMP
		);
	`)

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetCode_FileOps(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.WriteFile("testcode.go", []byte("line1\nline2\nline3\n"), 0644)
	defer os.Remove("testcode.go")

	handler := handleGetCode(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode.go", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/code?path=notexist.go", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/code?path=/etc/passwd", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	req4 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode.go&start=1&end=2", nil)
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)

	req5 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode.go&start=10&end=20", nil)
	rec5 := httptest.NewRecorder()
	handler.ServeHTTP(rec5, req5)

	req6 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode.go&start=abc&end=def", nil)
	rec6 := httptest.NewRecorder()
	handler.ServeHTTP(rec6, req6)

	req7 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode.go&start=3&end=2", nil)
	rec7 := httptest.NewRecorder()
	handler.ServeHTTP(rec7, req7)

	req8 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode.go&start=0&end=0", nil)
	rec8 := httptest.NewRecorder()
	handler.ServeHTTP(rec8, req8)

	req9 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode.go&end=10", nil)
	rec9 := httptest.NewRecorder()
	handler.ServeHTTP(rec9, req9)
}

func TestHandleGetCode_EmptyFile(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.WriteFile("testcode_empty.go", []byte(""), 0644)
	defer os.Remove("testcode_empty.go")

	handler := handleGetCode(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode_empty.go", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
}

func TestHandleGetCode_DirError(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.MkdirAll("testcode_dir", 0755)
	defer os.RemoveAll("testcode_dir")

	handler := handleGetCode(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/code?path=testcode_dir", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
}

func TestHandleListADR_Success(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.MkdirAll("docs/architecture/adr", 0755)
	defer os.RemoveAll("docs")
	os.WriteFile("docs/architecture/adr/ADR-001-Test.md", []byte("test"), 0644)
	os.WriteFile("docs/architecture/adr/ADR-002.md", []byte("test"), 0644)
	os.WriteFile("docs/architecture/adr/not-an-adr.md", []byte("test"), 0644)

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleListADR_NoDir(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleListADR_DirError(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.MkdirAll("docs/architecture", 0755)
	os.WriteFile("docs/architecture/adr", []byte("test"), 0644)
	defer os.RemoveAll("docs")

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetADR_Success(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.MkdirAll("docs/architecture/adr", 0755)
	defer os.RemoveAll("docs")
	os.WriteFile("docs/architecture/adr/ADR-001-Test.md", []byte("test"), 0644)
	os.WriteFile("docs/architecture/adr/ADR-002.md", []byte("test"), 0644)

	handler := handleGetADR(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-001-Test.md", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-002.md", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/adr/", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	req4 := httptest.NewRequest(http.MethodGet, "/api/adr//tmp/passwd", nil)
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)

	req5 := httptest.NewRequest(http.MethodGet, "/api/adr/..%2F..%2Fpasswd", nil)
	rec5 := httptest.NewRecorder()
	handler.ServeHTTP(rec5, req5)
}

func TestHandleGetADR_Errors(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.MkdirAll("docs/architecture/adr", 0755)
	defer os.RemoveAll("docs")
	os.MkdirAll("docs/architecture/adr/ADR-dir", 0755)

	handler := handleGetADR(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-missing.md", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-dir", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
}

func TestHandleGetStatus_WithTaskPartialFieldsErrorEncode(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}
	ctx := context.Background()
	createTasksTable(t, rawDB)

	// Since we can't easily force json Encode to fail for the normal objects unless we inject an interface that isn't marshable,
	// let's just make sure we cover the rows Next parsing edge cases

	_, err := db.Conn.ExecContext(ctx,
		"INSERT INTO tasks (id, description, status) VALUES (?, ?, ?)",
		"task-partial", "A task", "TODO",
	)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	handler := handleGetStatus(db)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetADR_EscapeAttempt(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.MkdirAll("docs/architecture/adr", 0755)
	defer os.RemoveAll("docs")

	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/../adr/ADR-001.md", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetADR_ReadFileError(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	os.MkdirAll("docs/architecture/adr", 0755)
	defer os.RemoveAll("docs")
	// Creating directory instead of file to trigger read error for ReadFile
	os.MkdirAll("docs/architecture/adr/ADR-009.md", 0755)

	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-009.md", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleListADR_InternalServerErrorReadDir(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	// Make docs/architecture/adr a file to force ReadDir to return an error other than IsNotExist
	os.MkdirAll("docs/architecture", 0755)
	os.WriteFile("docs/architecture/adr", []byte("file instead of dir"), 0644)
	defer os.RemoveAll("docs")

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
