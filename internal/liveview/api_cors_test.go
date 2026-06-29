package liveview

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetLocalCORS(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		expectCORS string
		expectVary string
	}{
		{"no origin", "", "", ""},
		{"invalid url", ":invalid", "", ""},
		{"valid localhost", "http://localhost:5173", "http://localhost:5173", "Origin"},
		{"valid 127.0.0.1", "http://127.0.0.1:3000", "http://127.0.0.1:3000", "Origin"},
		{"invalid host", "http://example.com", "", ""},
		{"subdomain spoof", "http://localhost.example.com", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			setLocalCORS(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.expectCORS {
				t.Errorf("expected CORS %q, got %q", tt.expectCORS, got)
			}
			if got := rec.Header().Get("Vary"); got != tt.expectVary {
				t.Errorf("expected Vary %q, got %q", tt.expectVary, got)
			}
		})
	}
}
