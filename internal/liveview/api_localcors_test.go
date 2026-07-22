package liveview

import (
	"net/http/httptest"
	"testing"
)

func TestSetLocalCORS_InvalidURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	// %ZZ is an invalid escape sequence, will cause url.Parse to fail
	req.Header.Set("Origin", "http://%ZZlocalhost")
	rec := httptest.NewRecorder()

	setLocalCORS(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected empty string for invalid URL, got %q", got)
	}
}
