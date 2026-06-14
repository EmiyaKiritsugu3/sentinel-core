package liveview

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestEndpointsCORS(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}

	tests := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"handleGetGraph", handleGetGraph(db), "/api/graph"},
		{"handleGetCode", handleGetCode(db), "/api/code?path=."},
		{"handleListADR", handleListADR(db), "/api/adr"},
		{"handleGetADR", handleGetADR(db), "/api/adr/dummy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Origin", "http://localhost:5173")
			rec := httptest.NewRecorder()
			tt.handler.ServeHTTP(rec, req)

			if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://localhost:5173" {
				t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", acao)
			}
		})
	}
}
