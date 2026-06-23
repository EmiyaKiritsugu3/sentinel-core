#!/bin/bash
sed -i '/req := httptest.NewRequest(http.MethodGet, "\/api\/status", nil)/a \\treq.Header.Set("Origin", "http://localhost:5173")' internal/liveview/api_test.go
sed -i 's/acao != "\*"/acao != "http:\/\/localhost:5173"/' internal/liveview/api_test.go
sed -i 's/expected Access-Control-Allow-Origin \*, got %q/expected Access-Control-Allow-Origin http:\/\/localhost:5173, got %q/' internal/liveview/api_test.go
sed -i '/t.Errorf("expected Access-Control-Allow-Origin http:\/\/localhost:5173, got %q", acao)/!b;n;a\	if vary := rec.Header().Get("Vary"); vary != "Origin" {\n\t\tt.Errorf("expected Vary: Origin, got %q", vary)\n\t}' internal/liveview/api_test.go

cat << 'INNEREOF' >> internal/liveview/api_test.go

func TestSetCORSHeaders(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectedVary   string
	}{
		{
			name:           "valid localhost",
			origin:         "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			expectedVary:   "Origin",
		},
		{
			name:           "valid 127.0.0.1",
			origin:         "http://127.0.0.1:8080",
			expectedOrigin: "http://127.0.0.1:8080",
			expectedVary:   "Origin",
		},
		{
			name:           "invalid origin evil.com",
			origin:         "https://evil.com",
			expectedOrigin: "",
			expectedVary:   "",
		},
		{
			name:           "empty origin",
			origin:         "",
			expectedOrigin: "",
			expectedVary:   "",
		},
		{
			name:           "invalid URL format",
			origin:         "://invalid",
			expectedOrigin: "",
			expectedVary:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			setCORSHeaders(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.expectedOrigin {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, got)
			}
			if got := rec.Header().Get("Vary"); got != tt.expectedVary {
				t.Errorf("expected Vary %q, got %q", tt.expectedVary, got)
			}
		})
	}
}
INNEREOF
go fmt internal/liveview/api_test.go
