package liveview

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func TestHandleGetCode(t *testing.T) {
	// Need to run from a clean temp directory to use relative paths
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	rawDB, _ := sql.Open("sqlite", ":memory:")
	defer rawDB.Close()
	db := &sqlite.DB{Conn: rawDB}
	handler := handleGetCode(db)

	filePath := "test.go"
	content := []byte("line1\nline2\nline3\nline4\n")
	_ = os.WriteFile(filePath, content, 0644)

	emptyFilePath := "empty.go"
	_ = os.WriteFile(emptyFilePath, []byte(""), 0644)

	tests := []struct {
		name       string
		path       string
		start      string
		end        string
		wantCode   int
		wantLines  []string
	}{
		{
			name:      "missing path",
			path:      "",
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "invalid path abs",
			path:      "/etc/passwd",
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "invalid path traversal",
			path:      "../test.go",
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "file not found",
			path:      "nonexistent.go",
			wantCode:  http.StatusNotFound,
		},
		{
			name:      "valid file full content",
			path:      filePath,
			wantCode:  http.StatusOK,
			wantLines: []string{"line1", "line2", "line3", "line4"},
		},
		{
			name:      "valid file with start and end",
			path:      filePath,
			start:     "2",
			end:       "3",
			wantCode:  http.StatusOK,
			wantLines: []string{"line2", "line3"},
		},
		{
			name:      "invalid start param",
			path:      filePath,
			start:     "abc",
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "invalid end param",
			path:      filePath,
			end:       "abc",
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "start param bounds under",
			path:      filePath,
			start:     "-5",
			wantCode:  http.StatusOK,
			wantLines: []string{"line1", "line2", "line3", "line4"},
		},
		{
			name:      "end param bounds over",
			path:      filePath,
			end:       "10",
			wantCode:  http.StatusOK,
			wantLines: []string{"line1", "line2", "line3", "line4"},
		},
		{
			name:      "start param over bounds",
			path:      filePath,
			start:     "10",
			wantCode:  http.StatusOK,
			wantLines: []string{"line4"},
		},
		{
			name:      "end param under bounds",
			path:      filePath,
			end:       "0",
			wantCode:  http.StatusOK,
			wantLines: []string{"line1"},
		},
		{
			name:      "end before start",
			path:      filePath,
			start:     "3",
			end:       "2",
			wantCode:  http.StatusOK,
			wantLines: []string{"line3"},
		},
		{
			name:      "empty file",
			path:      emptyFilePath,
			wantCode:  http.StatusOK,
			wantLines: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqUrl := "/api/code?path=" + tt.path
			if tt.start != "" {
				reqUrl += "&start=" + tt.start
			}
			if tt.end != "" {
				reqUrl += "&end=" + tt.end
			}

			req := httptest.NewRequest(http.MethodGet, reqUrl, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("got code %d, want %d", rec.Code, tt.wantCode)
			}

			if tt.wantCode == http.StatusOK {
				var resp map[string]any
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("decode resp: %v", err)
				}

				linesInterface, ok := resp["lines"].([]any)
				if !ok {
					t.Fatalf("lines field missing or not array: %v", resp)
				}

				var lines []string
				for _, l := range linesInterface {
					lines = append(lines, l.(string))
				}

				if len(lines) != len(tt.wantLines) {
					t.Errorf("got %d lines, want %d", len(lines), len(tt.wantLines))
				} else {
					for i, l := range lines {
						if l != tt.wantLines[i] {
							t.Errorf("line %d: got %q, want %q", i, l, tt.wantLines[i])
						}
					}
				}
			}
		})
	}
}
