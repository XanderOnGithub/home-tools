package games

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestArt(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	servers := []Server{
		{ID: "summer", Name: "Summer", Game: GameMinecraft, Container: "mc"},
		{ID: "creative", Name: "Creative", Game: GameMinecraft, Container: "mc2"},
		{ID: "vh", Name: "Valheim", Game: GameValheim, Container: "vh"},
	}
	for _, s := range servers {
		if err := store.SaveServer(s); err != nil {
			t.Fatal(err)
		}
	}
	art := filepath.Join(dir, "art")
	write := func(name, body string) {
		t.Helper()
		if err := os.MkdirAll(art, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(art, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("minecraft.jpg", "game cover")    // every Minecraft server
	write("summer.png", "own cover")        // …but Summer has its own
	write("minecraft-icon.webp", "mc icon") // icons are separate
	write("valheim.gif", "wrong extension") // ignored
	// A folder named like a picture isn't one.
	if err := os.Mkdir(filepath.Join(art, "vh.jpg"), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		id, kind string
		code     int
		body     string
		mime     string
	}{
		{"summer", "cover", http.StatusOK, "own cover", "image/png"},
		{"creative", "cover", http.StatusOK, "game cover", "image/jpeg"},
		{"creative", "icon", http.StatusOK, "mc icon", "image/webp"},
		{"vh", "cover", http.StatusNotFound, "", ""},
		{"vh", "icon", http.StatusNotFound, "", ""},
		{"summer", "banner", http.StatusNotFound, "", ""}, // unknown kind
		{"nope", "cover", http.StatusNotFound, "", ""},    // unknown server
	}
	mux := http.NewServeMux()
	Register(mux, store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, tt := range tests {
		rec := do(mux, "GET", "/api/servers/"+tt.id+"/art/"+tt.kind)
		if rec.Code != tt.code {
			t.Errorf("%s %s: status %d, want %d", tt.id, tt.kind, rec.Code, tt.code)
			continue
		}
		if tt.code != http.StatusOK {
			continue
		}
		if got := rec.Body.String(); got != tt.body {
			t.Errorf("%s %s: body %q, want %q", tt.id, tt.kind, got, tt.body)
		}
		if got := rec.Header().Get("Content-Type"); got != tt.mime {
			t.Errorf("%s %s: type %q, want %q", tt.id, tt.kind, got, tt.mime)
		}
	}

	// The list says which pictures each server has, so the UI only asks
	// for ones that exist.
	var list []serverView
	if err := json.Unmarshal(do(mux, "GET", "/api/servers").Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	want := map[string][]ArtKind{"summer": {ArtCover, ArtIcon}, "creative": {ArtCover, ArtIcon}, "vh": nil}
	for _, v := range list {
		if !slices.Equal(v.Art, want[v.ID]) {
			t.Errorf("%s art = %v, want %v", v.ID, v.Art, want[v.ID])
		}
	}
}
