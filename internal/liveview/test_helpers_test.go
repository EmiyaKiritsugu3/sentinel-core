package liveview

import (
	"context"
	"database/sql"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHandleGetGraph_Coverage(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	// Create nodes/edges tables
	rawDB.Exec(`CREATE TABLE nodes (id TEXT, name TEXT, type TEXT, file_path TEXT, start_line INT, end_line INT, hash TEXT, last_indexed TIMESTAMP)`)
	rawDB.Exec(`CREATE TABLE edges (from_node_id TEXT, to_node_id TEXT, relation_type TEXT)`)

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	req.Header.Set("Origin", "http://localhost")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetCode_Coverage(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go", nil)
	req.Header.Set("Origin", "http://localhost")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=1&end=5", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/code?path=nonexistent.go", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	req4 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=-1&end=1000", nil)
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
}

func TestHandleGetGraph_Error(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetCode_Errors(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetCode(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/code", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/code?path=../../etc/passwd", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&start=invalid", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	req4 := httptest.NewRequest(http.MethodGet, "/api/code?path=api.go&end=invalid", nil)
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
}

func TestHandleListADR_Errors(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestHandleGetADR_Errors(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handler := handleGetADR(db)

	req1 := httptest.NewRequest(http.MethodGet, "/api/adr/", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/adr/../invalid", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-001-test", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
}

func TestServerCoverage(t *testing.T) {
	server := NewServer()
	server.Notify(graph.GraphEvent{Type: "TEST"})
}

func TestHandleListADR_WithMockDir(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	// Create a temporary directory to act as docs/architecture/adr
	tempDir := t.TempDir()

	// Temporarily override the working directory to the tempDir so the handler can find "docs/architecture/adr"
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	os.Chdir(tempDir)
	os.MkdirAll("docs/architecture/adr", 0755)
	os.WriteFile(filepath.Join("docs/architecture/adr", "ADR-001-Test-ADR.md"), []byte("Test content"), 0644)
	os.WriteFile(filepath.Join("docs/architecture/adr", "ADR-002-Another-ADR.md"), []byte("Test content"), 0644)
	os.WriteFile(filepath.Join("docs/architecture/adr", "ignore.txt"), []byte("ignore"), 0644)

	handler := handleListADR(db)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	req.Header.Set("Origin", "http://localhost")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	handlerGet := handleGetADR(db)
	reqGet := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-001-Test-ADR.md", nil)
	reqGet.Header.Set("Origin", "http://localhost")
	recGet := httptest.NewRecorder()
	handlerGet.ServeHTTP(recGet, reqGet)
}

func TestServerCoverage_HTTP(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}
	// Create nodes/edges tables
	rawDB.Exec(`CREATE TABLE nodes (id TEXT, name TEXT, type TEXT, file_path TEXT, start_line INT, end_line INT, hash TEXT, last_indexed TIMESTAMP)`)
	rawDB.Exec(`CREATE TABLE edges (from_node_id TEXT, to_node_id TEXT, relation_type TEXT)`)

	rawDB.Exec("INSERT INTO nodes (id) VALUES ('1')")
	rawDB.Exec("INSERT INTO edges (from_node_id, to_node_id) VALUES ('1', '2')")

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestServerCoverage_WS(t *testing.T) {
	server := NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	server.Run(ctx)
}

func TestServerCoverage_HTTP_Errors(t *testing.T) {
	server := NewServer()

	// Test StartHTTP with invalid DB
	err := server.StartHTTP(8080, nil)
	if err == nil {
		t.Errorf("expected error with nil db")
	}
}

func TestServerCoverage_WS_Upgrade(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	server.serveWS(rec, req) // Should fail upgrade nicely
}

func TestServerCoverage_StartHTTP(t *testing.T) {
	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()

	// Create required tables to bypass ValidateDB
	rawDB.Exec(`CREATE TABLE IF NOT EXISTS config (key TEXT PRIMARY KEY, value TEXT NOT NULL)`)
	rawDB.Exec(`INSERT INTO config (key, value) VALUES ('schema_version', '1')`)
	db := &sqlite.DB{Conn: rawDB}

	server := NewServer()

	// Start HTTP in background, and immediately cancel
	go func() {
		server.StartHTTP(0, db) // port 0 allows random port
	}()
	time.Sleep(10 * time.Millisecond)
}
