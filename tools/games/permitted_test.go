package games

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPermittedListHandlers(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "adminlist", permittedListName)
	if err := os.MkdirAll(filepath.Dir(list), 0o755); err != nil {
		t.Fatal(err)
	}
	// Windows line endings, a comment, and the "Steam_" form must survive.
	if err := os.WriteFile(list, []byte("// List permitted players ID  ONE per line\r\nSteam_76561198000000001\r\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []Server{
		{ID: "valheim", Name: "Valheim", Game: GameValheim, Container: "vh", PermittedList: list},
		{ID: "fresh", Name: "Fresh", Game: GameValheim, Container: "vh2", PermittedList: filepath.Join(dir, "new", permittedListName)},
	} {
		if err := store.SaveServer(s); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	Register(mux, store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	tests := []struct {
		method, path string
		wantStatus   int
		wantBody     string
	}{
		{"GET", "/api/servers/valheim/whitelist", 200, `{"players":["76561198000000001"]}`},
		{"PUT", "/api/servers/valheim/whitelist/76561198000000001", 200, "already on"},
		{"PUT", "/api/servers/valheim/whitelist/76561198000000002", 200, `"restart_needed":true`},
		{"GET", "/api/servers/valheim/whitelist", 200, `{"players":["76561198000000001","76561198000000002"]}`},
		{"DELETE", "/api/servers/valheim/whitelist/76561198000000001", 200, "Removed"},
		{"DELETE", "/api/servers/valheim/whitelist/76561198000000001", 200, "isn't on"},
		{"PUT", "/api/servers/valheim/whitelist/12345", 400, "SteamID64"},
		{"PUT", "/api/servers/valheim/whitelist/Steve", 400, "SteamID64"},
		{"PUT", "/api/servers/fresh/whitelist/76561198000000003", 200, "Added"}, // no file yet
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
		if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), tt.wantBody) {
			t.Errorf("%s %s = %d %s, want %d containing %q", tt.method, tt.path, rec.Code, rec.Body, tt.wantStatus, tt.wantBody)
		}
	}

	got, err := os.ReadFile(list)
	if err != nil {
		t.Fatal(err)
	}
	if want := "// List permitted players ID  ONE per line\n76561198000000002\n"; string(got) != want {
		t.Errorf("file = %q, want %q", got, want)
	}
	if info, _ := os.Stat(list); info.Mode().Perm() != 0o640 {
		t.Errorf("mode = %v, want the original 0640 kept", info.Mode().Perm())
	}
}

func TestPermittedListValidate(t *testing.T) {
	ok := Server{ID: "valheim", Name: "Valheim", Game: GameValheim, Container: "vh"}
	tests := []struct {
		name, path string
		game       Game
		wantErr    bool
	}{
		{"valid", "/valheim/adminlist/permittedlist.txt", GameValheim, false},
		{"relative", "valheim/permittedlist.txt", GameValheim, true},
		{"other file", "/etc/passwd", GameValheim, true},
		{"unclean", "/valheim/../etc/permittedlist.txt", GameValheim, true},
		{"minecraft", "/valheim/permittedlist.txt", GameMinecraft, true},
	}
	for _, tt := range tests {
		s := ok
		s.Game, s.PermittedList = tt.game, tt.path
		if err := s.Validate(); (err != nil) != tt.wantErr {
			t.Errorf("%s: Validate() = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestBusy(t *testing.T) {
	var b busy
	if _, ok := b.begin("mc", ActionRestart); !ok {
		t.Fatal("first begin refused")
	}
	if running, ok := b.begin("mc", ActionStop); ok || running != ActionRestart {
		t.Errorf("second begin = %q, %v; want restart, false", running, ok)
	}
	if _, ok := b.begin("vh", ActionStop); !ok {
		t.Error("another server is blocked")
	}
	if b.action("mc") != ActionRestart {
		t.Errorf("action = %q", b.action("mc"))
	}
	b.end("mc")
	if _, ok := b.begin("mc", ActionStart); !ok {
		t.Error("begin after end refused")
	}
}
