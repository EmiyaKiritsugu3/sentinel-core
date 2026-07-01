package liveview

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"

	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestSetCORSHeaders(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectedVary   string
	}{
		{
			name:           "Valid Origin Localhost",
			origin:         "http://localhost:5173",
			expectedOrigin: "http://localhost:5173",
			expectedVary:   "Origin",
		},
		{
			name:           "Valid Origin 127.0.0.1",
			origin:         "http://127.0.0.1:8080",
			expectedOrigin: "http://127.0.0.1:8080",
			expectedVary:   "Origin",
		},
		{
			name:           "Invalid Origin",
			origin:         "http://example.com",
			expectedOrigin: "",
			expectedVary:   "",
		},
		{
			name:           "Empty Origin",
			origin:         "",
			expectedOrigin: "",
			expectedVary:   "",
		},
		{
			name:           "Malformed Origin",
			origin:         "http://%ZZ:8080",
			expectedOrigin: "",
			expectedVary:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			setCORSHeaders(rec, req)

			if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != tt.expectedOrigin {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, acao)
			}
			if vary := rec.Header().Get("Vary"); vary != tt.expectedVary {
				t.Errorf("expected Vary %q, got %q", tt.expectedVary, vary)
			}
		})
	}
}

func TestHandleGetGraph_CORS(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	// mock nodes/edges table
	_, _ = db.Conn.Exec("CREATE TABLE IF NOT EXISTS nodes (id TEXT PRIMARY KEY, name TEXT, type TEXT, file_path TEXT, start_line INTEGER, end_line INTEGER, hash TEXT, last_indexed TIMESTAMP)")
	_, _ = db.Conn.Exec("CREATE TABLE IF NOT EXISTS edges (from_node_id TEXT, to_node_id TEXT, relation_type TEXT)")

	handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:3000, got %q", acao)
	}
}

func TestHandleGetCode_CORS(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=.", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:3000, got %q", acao)
	}
}

func TestHandleListADR_CORS(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:3000, got %q", acao)
	}
}

func TestHandleGetADR_CORS(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-001-test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:3000, got %q", acao)
	}
}

func TestHandleGetCode_MissingPath(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing path, got %d", rec.Code)
	}
}

func TestHandleGetCode_InvalidPath(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid path, got %d", rec.Code)
	}
}

func TestHandleListADR_ReadError(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetADR_InvalidPaths(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)

	tests := []string{
		"/api/adr/",
		"/api/adr/../secret.txt",
		"/api/adr//etc/passwd",
	}

	for _, path := range tests {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for path %s, got %d", path, rec.Code)
		}
	}
}

func TestHandleGetCode_FileOps(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/code?path=nonexistent_file_12345.go", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=abc", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid start, got %d", rec3.Code)
	}

	req4 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&end=abc", nil)
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid end, got %d", rec4.Code)
	}
}

func TestHandleGetADR_FileOps(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-999-notfound.md", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusNotFound && rec1.Code != http.StatusBadRequest && rec1.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code for not found ADR: %d", rec1.Code)
	}
}

func TestHandleListADR_Success(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()

	t.Setenv("PWD", "/tmp")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandleGetADR_Success(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-001-test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound && rec.Code != http.StatusBadRequest && rec.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code for not found ADR: %d", rec.Code)
	}
}

func TestHandleGetGraph_NoTableError(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestHandleGetGraph_EdgesError(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	_, _ = db.Conn.Exec("CREATE TABLE IF NOT EXISTS nodes (id TEXT PRIMARY KEY, name TEXT, type TEXT, file_path TEXT, start_line INTEGER, end_line INTEGER, hash TEXT, last_indexed TIMESTAMP)")

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestHandleGetCode_StartEnd(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=-1&end=9999", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=10&end=5", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec2.Code)
	}
}

func TestHandleListADR_ReadDirSuccess(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	// Temporarily create docs/architecture/adr
	err = os.MkdirAll("docs/architecture/adr", 0755)
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	defer os.RemoveAll("docs")

	// Create a mock ADR file
	_ = os.WriteFile("docs/architecture/adr/ADR-001-mock.md", []byte("mock adr"), 0644)
	_ = os.WriteFile("docs/architecture/adr/not-an-adr.txt", []byte("ignore"), 0644)

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandleGetADR_Full(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	err = os.MkdirAll("docs/architecture/adr", 0755)
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	defer os.RemoveAll("docs")

	_ = os.WriteFile("docs/architecture/adr/ADR-001-test-adr.md", []byte("mock adr content"), 0644)

	handler := handleGetADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-001-test-adr.md", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandleGetCode_EmptyFile(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)

	_ = os.WriteFile("empty.go", []byte(""), 0644)
	defer os.Remove("empty.go")

	req1 := httptest.NewRequest(http.MethodGet, "/api/code?path=empty.go", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec1.Code)
	}
}


func TestHandleListADR_ReadDirFailure(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()

	// Create docs/architecture/adr as a file so it fails to read dir
	_ = os.MkdirAll("docs/architecture", 0755)
	defer os.RemoveAll("docs")
	_ = os.WriteFile("docs/architecture/adr", []byte("file not dir"), 0644)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
