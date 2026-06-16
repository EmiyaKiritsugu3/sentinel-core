package liveview

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestHandlers_HitCORS(t *testing.T) {
	t.Parallel()
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}

	handlers := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"handleGetGraph", handleGetGraph(db), "/api/graph"},
		{"handleGetStatus", handleGetStatus(db), "/api/status"},
		{"handleGetCode", handleGetCode(db), "/api/code?path=x"},
		{"handleListADR", handleListADR(db), "/api/adr"},
		{"handleGetADR", handleGetADR(db), "/api/adr/ADR-001.md"},
	}

	for _, tt := range handlers {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Origin", "http://localhost:5173")
			rec := httptest.NewRecorder()

			tt.handler.ServeHTTP(rec, req)

			// We don't care about the response code here,
			// just that the CORS header was set correctly by the new line.
			if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
				t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
			}
		})
	}
}
