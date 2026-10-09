package games

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestParseWhitelist(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"There are 2 whitelisted player(s): alice, bob", []string{"alice", "bob"}},
		{"There are 1 whitelisted players: alice", []string{"alice"}},
		{"There are no whitelisted players", []string{}},
	}
	for _, tt := range tests {
		got, err := parseWhitelist(tt.in)
		if err != nil || !slices.Equal(got, tt.want) {
			t.Errorf("%q = %v, %v; want %v", tt.in, got, err, tt.want)
		}
	}
	if _, err := parseWhitelist("Unknown command"); err == nil {
		t.Error("parseWhitelist accepted an unknown answer")
	}
}

func TestConsoleHandlers(t *testing.T) {
	addr, cmds := fakeRCONReplies(t, "secret", map[string]string{
		"say hi":                 "",
		"whitelist list":         "There are 1 whitelisted player(s): alice",
		"whitelist add bob":      "§eAdded bob to the whitelist",
		"whitelist remove alice": "Removed alice from the whitelist",
	})
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []Server{
		{ID: "minecraft", Name: "Minecraft", Game: GameMinecraft, Container: "mc", Query: addr, RCONPassword: "secret"},
		{ID: "norcon", Name: "No RCON", Game: GameMinecraft, Container: "mc"},
		{ID: "valheim", Name: "Valheim", Game: GameValheim, Container: "vh"},
	} {
		if err := store.SaveServer(s); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	Register(mux, store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	tests := []struct {
		method, path, body string
		wantStatus         int
		wantBody           string
	}{
		{"POST", "/api/servers/minecraft/console", `{"command":"/say hi"}`, 200, `"output":""`},
		{"GET", "/api/servers/minecraft/whitelist", "", 200, `{"players":["alice"]}`},
		{"PUT", "/api/servers/minecraft/whitelist/bob", "", 200, `"Added bob to the whitelist"`},
		{"DELETE", "/api/servers/minecraft/whitelist/alice", "", 200, `"Removed alice from the whitelist"`},
		{"POST", "/api/servers/minecraft/console", `{"command":"  "}`, 400, "empty"},
		{"POST", "/api/servers/minecraft/console", `{"command":"say a\nop bob"}`, 400, "line breaks"},
		{"POST", "/api/servers/minecraft/console", `{"cmd":"say hi"}`, 400, ""},
		{"PUT", "/api/servers/minecraft/whitelist/bob%20op", "", 400, "username"},
		{"POST", "/api/servers/norcon/console", `{"command":"say hi"}`, 409, "RCON"},
		{"POST", "/api/servers/valheim/console", `{"command":"say hi"}`, 404, ""},
		{"PUT", "/api/servers/valheim/whitelist/bob", "", 404, ""},
		{"POST", "/api/servers/nope/console", `{"command":"say hi"}`, 404, ""},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body)))
			if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("= %d %s, want %d containing %q", rec.Code, rec.Body, tt.wantStatus, tt.wantBody)
			}
		})
	}
	want := []string{"say hi", "whitelist list", "whitelist add bob", "whitelist remove alice"}
	if got := cmds.all(); !slices.Equal(got, want) {
		t.Errorf("commands sent = %q, want %q", got, want)
	}

	// The list tells the UI where the console works.
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/servers", nil))
	var views []serverView
	if err := json.Unmarshal(rec.Body.Bytes(), &views); err != nil {
		t.Fatal(err)
	}
	for _, v := range views {
		if v.Console != (v.ID == "minecraft") {
			t.Errorf("%s console = %v", v.ID, v.Console)
		}
	}
}
