package liveview

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestSetCORS_InvalidURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	// Injecting an invalid URL that fails url.Parse
	// Typically, url.Parse is quite permissive, but we can use control characters.
	req.Header.Set("Origin", "http://[::1]:namedport")
	rec := httptest.NewRecorder()

    setCORS(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin, got %q", acao)
	}
}


func TestSetCORS_ValidURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

    setCORS(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:3000" {
		t.Errorf("expected http://localhost:3000 Access-Control-Allow-Origin, got %q", acao)
	}
}

func TestSetCORS_EmptyOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

    setCORS(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin, got %q", acao)
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
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

    // Creating tables just so query won't error out before CORS is tested
    _, err = db.Conn.ExecContext(context.Background(), "CREATE TABLE nodes (id TEXT, name TEXT, type TEXT, file_path TEXT, start_line INTEGER, end_line INTEGER, hash TEXT, last_indexed INTEGER)")
    if err != nil {
        t.Fatalf("create nodes table: %v", err)
    }
    _, err = db.Conn.ExecContext(context.Background(), "CREATE TABLE edges (from_node_id TEXT, to_node_id TEXT, relation_type TEXT)")
    if err != nil {
        t.Fatalf("create edges table: %v", err)
    }

    handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
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
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=main.go", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
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
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
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
	req := httptest.NewRequest(http.MethodGet, "/api/adr/ADR-001-Test.md", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
	}
}
