package liveview

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCORS(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectedVary   string
	}{
		{
			name:           "Empty Origin",
			origin:         "",
			expectedOrigin: "",
			expectedVary:   "",
		},
		{
			name:           "Invalid URL Origin",
			origin:         "://invalid",
			expectedOrigin: "",
			expectedVary:   "",
		},
		{
			name:           "Valid localhost Origin",
			origin:         "http://localhost:5173",
			expectedOrigin: "http://localhost:5173",
			expectedVary:   "Origin",
		},
		{
			name:           "Valid 127.0.0.1 Origin",
			origin:         "http://127.0.0.1:8080",
			expectedOrigin: "http://127.0.0.1:8080",
			expectedVary:   "Origin",
		},
		{
			name:           "Invalid Domain Origin",
			origin:         "http://example.com",
			expectedOrigin: "",
			expectedVary:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			setCORS(rec, req)

			if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != tt.expectedOrigin {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, acao)
			}
			if vary := rec.Header().Get("Vary"); vary != tt.expectedVary {
				t.Errorf("expected Vary %q, got %q", tt.expectedVary, vary)
			}
		})
	}
}
