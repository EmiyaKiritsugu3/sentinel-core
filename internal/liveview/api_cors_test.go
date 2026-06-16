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
		wantHeader string
	}{
		{
			name:       "empty origin",
			origin:     "",
			wantHeader: "",
		},
		{
			name:       "valid localhost",
			origin:     "http://localhost:5173",
			wantHeader: "http://localhost:5173",
		},
		{
			name:       "valid 127.0.0.1",
			origin:     "http://127.0.0.1:8080",
			wantHeader: "http://127.0.0.1:8080",
		},
		{
			name:       "invalid origin parsing error",
			origin:     "://invalid-url",
			wantHeader: "",
		},
		{
			name:       "malicious subdomain",
			origin:     "http://localhost.evil.com",
			wantHeader: "",
		},
		{
			name:       "arbitrary domain",
			origin:     "https://example.com",
			wantHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			setLocalCORS(rec, req)

			got := rec.Header().Get("Access-Control-Allow-Origin")
			if got != tt.wantHeader {
				t.Errorf("setLocalCORS() Access-Control-Allow-Origin = %v, want %v", got, tt.wantHeader)
			}
		})
	}
}

func TestUpgraderCheckOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{
			name:   "empty origin",
			origin: "",
			want:   true, // Upgrader allows empty origin for non-browser clients
		},
		{
			name:   "valid localhost",
			origin: "http://localhost:5173",
			want:   true,
		},
		{
			name:   "valid 127.0.0.1",
			origin: "http://127.0.0.1:8080",
			want:   true,
		},
		{
			name:   "invalid origin parsing error",
			origin: "://invalid-url",
			want:   false,
		},
		{
			name:   "malicious subdomain",
			origin: "http://localhost.evil.com",
			want:   false,
		},
		{
			name:   "arbitrary domain",
			origin: "https://example.com",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			if got := upgrader.CheckOrigin(req); got != tt.want {
				t.Errorf("upgrader.CheckOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
