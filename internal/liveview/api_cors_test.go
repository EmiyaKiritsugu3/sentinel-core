package liveview

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCorsHeaders(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectedHeader string
	}{
		{
			name:           "Empty Origin",
			origin:         "",
			expectedHeader: "",
		},
		{
			name:           "Valid Origin Localhost",
			origin:         "http://localhost:5173",
			expectedHeader: "http://localhost:5173",
		},
		{
			name:           "Valid Origin 127.0.0.1",
			origin:         "http://127.0.0.1:3000",
			expectedHeader: "http://127.0.0.1:3000",
		},
		{
			name:           "Invalid Origin Hostname",
			origin:         "http://evil.com",
			expectedHeader: "",
		},
		{
			name:           "Invalid URL Parse",
			origin:         "://invalid-url",
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			setCorsHeaders(rec, req)

			actualHeader := rec.Header().Get("Access-Control-Allow-Origin")
			if actualHeader != tt.expectedHeader {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedHeader, actualHeader)
			}
		})
	}
}
