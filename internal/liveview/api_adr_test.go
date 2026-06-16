package liveview

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestHandleListADR(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	adrDir := filepath.Join("docs", "architecture", "adr")
	os.MkdirAll(adrDir, 0755)

	// Create some ADR files
	os.WriteFile(filepath.Join(adrDir, "ADR-001-initial-design.md"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(adrDir, "ADR-002.md"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(adrDir, "not-an-adr.txt"), []byte("content"), 0644)

	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}
	handler := handleListADR(db)

	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)

	adrs, ok := resp["adrs"].([]any)
	if !ok || len(adrs) != 2 {
		t.Fatalf("expected 2 adrs, got %v", resp["adrs"])
	}
}

func TestHandleListADR_NoDir(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}
	handler := handleListADR(db)

	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleGetADR(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	adrDir := filepath.Join("docs", "architecture", "adr")
	os.MkdirAll(adrDir, 0755)
	os.WriteFile(filepath.Join(adrDir, "ADR-001-initial-design.md"), []byte("content 1"), 0644)

	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}
	handler := handleGetADR(db)

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"valid", "/api/adr/ADR-001-initial-design.md", http.StatusOK},
		{"not found", "/api/adr/ADR-002.md", http.StatusNotFound},
		{"traversal", "/api/adr/../ADR-001-initial-design.md", http.StatusBadRequest},
		{"empty", "/api/adr/", http.StatusBadRequest},
		{"absolute", "/api/adr//etc/passwd", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("got code %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}
