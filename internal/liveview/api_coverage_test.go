package liveview

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestHandleGetGraph_Coverage(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	// Create necessary tables
	_, err = db.Conn.Exec(`
		CREATE TABLE nodes (id TEXT, name TEXT, type TEXT, file_path TEXT, start_line INT, end_line INT, hash TEXT, last_indexed TEXT);
		CREATE TABLE edges (from_node_id TEXT, to_node_id TEXT, relation_type TEXT);
	`)
	if err != nil {
		t.Fatalf("create tables: %v", err)
	}

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestHandleGetCode_Coverage(t *testing.T) {
	handler := handleGetCode(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/code?path=nonexistent", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestHandleListADR_Coverage(t *testing.T) {
	handler := handleListADR(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestHandleGetADR_Coverage(t *testing.T) {
	handler := handleGetADR(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/adr/invalid", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
