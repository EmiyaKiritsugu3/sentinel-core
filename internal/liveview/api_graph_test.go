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

func createNodesAndEdgesTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
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
	`)
	if err != nil {
		t.Fatalf("create nodes/edges tables: %v", err)
	}
}

func TestHandleGetGraph_Success(t *testing.T) {
	t.Parallel()

	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	createNodesAndEdgesTables(t, rawDB)

	_, err = rawDB.Exec(`
		INSERT INTO nodes (id, name, type, file_path, start_line, end_line, hash, last_indexed)
		VALUES ('n1', 'Node1', 'function', 'main.go', 1, 10, 'hash1', '2023-01-01');
		INSERT INTO edges (from_node_id, to_node_id, relation_type)
		VALUES ('n1', 'n2', 'calls');
	`)
	if err != nil {
		t.Fatalf("insert data: %v", err)
	}

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var snapshot GraphSnapshot
	if err := json.NewDecoder(rec.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}

	if len(snapshot.Nodes) != 1 || snapshot.Nodes[0].ID != "n1" {
		t.Errorf("expected 1 node with ID n1, got %+v", snapshot.Nodes)
	}
	if len(snapshot.Edges) != 1 || snapshot.Edges[0].From != "n1" || snapshot.Edges[0].To != "n2" || snapshot.Edges[0].Rel != "calls" {
		t.Errorf("expected 1 edge n1->n2 (calls), got %+v", snapshot.Edges)
	}
}

func TestHandleGetGraph_NodesError(t *testing.T) {
	t.Parallel()
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	// Missing nodes table -> 500 error
	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHandleGetGraph_EdgesError(t *testing.T) {
	t.Parallel()
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	// Create nodes but NOT edges table
	_, _ = rawDB.Exec(`CREATE TABLE nodes (id TEXT, name TEXT, type TEXT, file_path TEXT, start_line INTEGER, end_line INTEGER, hash TEXT, last_indexed TEXT)`)

	handler := handleGetGraph(db)
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
