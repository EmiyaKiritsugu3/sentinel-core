#!/bin/bash
cat << 'INNER_EOF' >> internal/liveview/api_test.go

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
INNER_EOF
