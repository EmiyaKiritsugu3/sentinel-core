package liveview

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestCORSHeaders(t *testing.T) {
	t.Parallel()

	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	// Just need handlers to run past CORS headers. For GetStatus, we need the table.
	createTasksTable(t, rawDB)
	_, _ = rawDB.Exec(`CREATE TABLE nodes (id TEXT)`) // For GetGraph
	_, _ = rawDB.Exec(`CREATE TABLE edges (from_node_id TEXT)`) // For GetGraph

	tests := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"GetGraph", handleGetGraph(db), "/api/graph"},
		{"GetStatus", handleGetStatus(db), "/api/status"},
		{"GetCode", handleGetCode(db), "/api/code?path=api.go"},
		{"ListADR", handleListADR(db), "/api/adr"},
		{"GetADR", handleGetADR(db), "/api/adr/ADR-001.md"},
	}

	origins := []struct {
		origin string
		valid  bool
	}{
		{"http://localhost:5173", true},
		{"http://127.0.0.1:8080", true},
		{"https://evil.com", false},
		{"", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, ot := range origins {
				ot := ot
				t.Run(ot.origin, func(t *testing.T) {
					t.Parallel()
					req := httptest.NewRequest(http.MethodGet, tt.path, nil)
					if ot.origin != "" {
						req.Header.Set("Origin", ot.origin)
					}
					rec := httptest.NewRecorder()
					tt.handler.ServeHTTP(rec, req)

					acao := rec.Header().Get("Access-Control-Allow-Origin")
					if ot.valid {
						if acao != ot.origin {
							t.Errorf("expected Access-Control-Allow-Origin %q, got %q", ot.origin, acao)
						}
					} else {
						if acao != "" {
							t.Errorf("expected empty Access-Control-Allow-Origin, got %q", acao)
						}
					}
				})
			}
		})
	}
}
