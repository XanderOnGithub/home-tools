package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestSPA(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":       {Data: []byte("<html>app</html>")},
		"assets/app-1a.js": {Data: []byte("js")},
		"favicon.svg":      {Data: []byte("<svg/>")},
	}
	h := SPA(fsys)

	tests := []struct {
		name, path  string
		code        int
		body, cache string
	}{
		{"root", "/", 200, "app", "no-cache"},
		{"client route", "/routines/abc", 200, "app", "no-cache"},
		{"directory", "/assets/", 200, "app", "no-cache"},
		// net/http refuses ".." outright (the mux would redirect it first).
		{"escape attempt", "/../../etc/passwd", 400, "invalid URL path", ""},
		{"hashed asset", "/assets/app-1a.js", 200, "js", "public, max-age=31536000, immutable"},
		{"other file", "/favicon.svg", 200, "<svg/>", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != tt.code {
				t.Fatalf("status = %d, want %d", rec.Code, tt.code)
			}
			if !strings.Contains(rec.Body.String(), tt.body) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.body)
			}
			if got := rec.Header().Get("Cache-Control"); got != tt.cache {
				t.Errorf("Cache-Control = %q, want %q", got, tt.cache)
			}
		})
	}
}
