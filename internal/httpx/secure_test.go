package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecure(t *testing.T) {
	ran := false
	h := Secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ran = true
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name    string
		method  string
		headers map[string]string
		allowed bool
	}{
		{"same-origin POST (the app itself)", "POST", map[string]string{"Sec-Fetch-Site": "same-origin"}, true},
		{"no browser headers (Discord bot, curl)", "POST", nil, true},
		{"cross-site POST (CSRF)", "POST", map[string]string{"Sec-Fetch-Site": "cross-site"}, false},
		{"another tool's subdomain", "PUT", map[string]string{"Sec-Fetch-Site": "same-site"}, false},
		{"old browser, other Origin", "DELETE", map[string]string{"Origin": "https://evil.example"}, false},
		{"cross-site GET (links work)", "GET", map[string]string{"Sec-Fetch-Site": "cross-site"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ran = false
			req := httptest.NewRequest(tt.method, "https://games.example.com/api/servers/mc/console", strings.NewReader(`{"command":"op eve"}`))
			req.Header.Set("Content-Type", "text/plain") // a "simple" request: no preflight
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if ran != tt.allowed {
				t.Errorf("handler ran = %v, want %v (status %d)", ran, tt.allowed, rec.Code)
			}
			if !tt.allowed && rec.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403", rec.Code)
			}
			for _, k := range []string{"Content-Security-Policy", "X-Content-Type-Options", "Referrer-Policy"} {
				if rec.Header().Get(k) == "" {
					t.Errorf("missing %s", k)
				}
			}
		})
	}
}
