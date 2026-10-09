package hostroute

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	tool := func(name string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(name)) })
	}
	byHost := New("")
	byHost.Handle("fitness", tool("fitness"))
	byHost.Handle("games", tool("games"))
	dev := New("games")
	dev.Handle("fitness", tool("fitness"))
	dev.Handle("games", tool("games"))

	tests := []struct {
		name   string
		router *Router
		host   string
		want   string // body; "" = 404
	}{
		{"fitness subdomain", byHost, "fitness.example.com", "fitness"},
		{"games with port", byHost, "games.example.com:443", "games"},
		{"case-insensitive", byHost, "Fitness.Example.com", "fitness"},
		{"unknown subdomain", byHost, "recipes.example.com", ""},
		{"bare localhost", byHost, "localhost:8080", ""},
		{"LAN IP", byHost, "192.168.3.3:8080", ""},
		{"IPv6", byHost, "[::1]:8080", ""},
		{"dev: localhost", dev, "localhost:8080", "games"},
		{"dev: any host", dev, "fitness.example.com", "games"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.host
			rec := httptest.NewRecorder()
			tt.router.ServeHTTP(rec, req)
			if tt.want == "" {
				if rec.Code != http.StatusNotFound {
					t.Errorf("status = %d, want 404", rec.Code)
				}
				return
			}
			if rec.Body.String() != tt.want {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.want)
			}
		})
	}
}
