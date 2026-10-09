package discord

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

const (
	alice = "100000000000000001" // verified
	bob   = "100000000000000002" // not verified
)

// fakeGames is the games API in memory: a running Minecraft server and a
// Valheim one, both with whitelists.
type fakeGames struct {
	mu        sync.Mutex
	restarts  []string
	restartFn func(id string) error
	edits     []string
}

func (f *fakeGames) Servers(context.Context) ([]GameServer, error) {
	mc := GameServer{ID: "minecraft", Name: "Minecraft", Game: "minecraft", Whitelist: true}
	vh := GameServer{ID: "valheim", Name: "Valheim", Game: "valheim", Whitelist: true}
	old := GameServer{ID: "old", Name: "Old", Game: "minecraft", Archived: true}
	return []GameServer{mc, vh, old}, nil
}

func (f *fakeGames) Restart(_ context.Context, id string) error {
	f.mu.Lock()
	f.restarts = append(f.restarts, id)
	fn := f.restartFn
	f.mu.Unlock()
	if fn != nil {
		return fn(id)
	}
	return nil
}

func (f *fakeGames) Whitelist(context.Context, string) ([]string, error) {
	return []string{"Steve", "Alex"}, nil
}

func (f *fakeGames) EditWhitelist(_ context.Context, id, player string, add bool) (WhitelistResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.edits = append(f.edits, id+" "+player)
	if id == "valheim" {
		return WhitelistResult{Output: "Added " + player + " to the permitted list.", RestartNeeded: true}, nil
	}
	return WhitelistResult{Output: "added " + player + " to the whitelist"}, nil
}

// fakeResponder records what the bot answered.
type fakeResponder struct {
	sent, edited []reply
	deferred     bool
	private      bool
}

func (r *fakeResponder) send(rep reply) error { r.sent = append(r.sent, rep); return nil }
func (r *fakeResponder) later(p bool) error   { r.deferred, r.private = true, p; return nil }
func (r *fakeResponder) edit(rep reply) error { r.edited = append(r.edited, rep); return nil }

// last is the final text the person sees.
func (r *fakeResponder) last() string {
	if n := len(r.edited); n > 0 {
		return r.edited[n-1].content
	}
	if n := len(r.sent); n > 0 {
		return r.sent[n-1].content
	}
	return ""
}

func newTestBot(t *testing.T) (*Bot, *fakeGames) {
	t.Helper()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg := store.Config()
	cfg.Verified = []VerifiedUser{{UserID: alice, Name: "Alice"}}
	if _, err := store.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	g := &fakeGames{}
	b := NewBot("", store, g, quiet)
	return b, g
}

func run(b *Bot, in invocation) *fakeResponder {
	r := &fakeResponder{}
	if err := b.handle(context.Background(), in, r); err != nil {
		panic(err)
	}
	return r
}

func TestRestart(t *testing.T) {
	b, g := newTestBot(t)
	now := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)
	b.now = func() time.Time { return now }
	restart := func(user, server string) *fakeResponder {
		return run(b, invocation{command: "restart", opts: map[string]any{"server": server}, userID: user, userName: "Someone"})
	}

	// Not verified: refused privately, and the request is remembered once.
	r := restart(bob, "minecraft")
	r = restart(bob, "minecraft")
	if !r.sent[0].private || !strings.Contains(r.last(), "Only verified") {
		t.Errorf("unverified: %+v", r.sent)
	}
	if reqs := b.store.Requests(); len(reqs) != 1 || reqs[0].UserID != bob || reqs[0].Command != "restart" {
		t.Errorf("requests = %+v", reqs)
	}

	// Verified: announced publicly, then edited with the result.
	r = restart(alice, "Minecraft") // by name works too
	if len(r.sent) != 1 || r.sent[0].private || !strings.Contains(r.last(), "✅") {
		t.Errorf("verified restart: sent %+v, edited %+v", r.sent, r.edited)
	}
	// Right after: cooldown, whoever asks.
	if r = restart(alice, "minecraft"); !strings.Contains(r.last(), "moments ago") {
		t.Errorf("cooldown: %q", r.last())
	}
	// Another server isn't affected.
	if r = restart(alice, "valheim"); !strings.Contains(r.last(), "✅") {
		t.Errorf("other server: %q", r.last())
	}
	// After the cooldown, it works again; the games tool's 409 is passed on.
	now = now.Add(restartCooldown)
	g.restartFn = func(string) error {
		return &GamesError{Status: 409, Message: "Minecraft is already restarting; try again in a minute"}
	}
	if r = restart(alice, "minecraft"); !strings.Contains(r.last(), "already restarting") {
		t.Errorf("409: %q", r.last())
	}
	g.restartFn = func(string) error { return errGamesDown }
	now = now.Add(restartCooldown)
	if r = restart(alice, "minecraft"); !strings.Contains(r.last(), "lost touch") {
		t.Errorf("games down: %q", r.last())
	}
	if r = restart(alice, "old"); !strings.Contains(r.last(), "no server") {
		t.Errorf("archived server: %q", r.last())
	}
	if got := strings.Join(g.restarts, ","); got != "minecraft,valheim,minecraft,minecraft" {
		t.Errorf("restarts = %s", got)
	}
}

func TestWhitelist(t *testing.T) {
	b, g := newTestBot(t)
	wl := func(sub, server, player string) *fakeResponder {
		return run(b, invocation{command: "whitelist", sub: sub, opts: map[string]any{"server": server, "player": player}, userID: alice, userName: "Alice"})
	}
	tests := []struct {
		sub, server, player string
		want                string
	}{
		{"add", "minecraft", "Steve", "Added Steve to the whitelist."},
		{"add", "minecraft", "no spaces", "isn't a Minecraft username"},
		{"add", "valheim", "76561198000000001", "/restart"},
		{"add", "valheim", "https://steamcommunity.com/profiles/76561198000000002/", "76561198000000002"},
		{"add", "valheim", "https://steamcommunity.com/id/coolname", "custom Steam link"},
		{"add", "valheim", "Steve", "SteamID64"},
		{"list", "minecraft", "", "Steve, Alex"},
	}
	for _, tt := range tests {
		if r := wl(tt.sub, tt.server, tt.player); !strings.Contains(r.last(), tt.want) {
			t.Errorf("%s %s %q = %q, want %q", tt.sub, tt.server, tt.player, r.last(), tt.want)
		}
	}
	if got := strings.Join(g.edits, ","); got != "minecraft Steve,valheim 76561198000000001,valheim 76561198000000002" {
		t.Errorf("edits = %s", got)
	}
	if r := wl("list", "minecraft", ""); !r.private {
		t.Error("a list should only show to the asker")
	}
}

func TestSensCommand(t *testing.T) {
	b, _ := newTestBot(t)
	r := run(b, invocation{command: "sens", opts: map[string]any{"from_game": "overwatch2", "value": 5.0, "to_game": "valorant", "from_dpi": int64(800), "to_dpi": int64(1600)}})
	e := r.sent[0].embed
	if e == nil || e.Fields[1].Value != "`0.236`" || e.Footer == nil || r.sent[0].private {
		t.Fatalf("sens = %+v", r.sent)
	}
	r = run(b, invocation{command: "sens", opts: map[string]any{"from_game": "doom", "value": 5.0, "to_game": "valorant"}})
	if !r.sent[0].private {
		t.Errorf("bad input should answer privately: %+v", r.sent)
	}
}

func TestComplete(t *testing.T) {
	b, _ := newTestBot(t)
	got := b.complete(context.Background(), invocation{command: "restart", focused: "server", opts: map[string]any{"server": "val"}})
	if len(got) != 1 || got[0].value != "valheim" {
		t.Errorf("complete = %+v", got)
	}
	if got := b.complete(context.Background(), invocation{command: "restart", focused: "server", opts: map[string]any{}}); len(got) != 2 {
		t.Errorf("empty input: %+v (archived must be hidden)", got)
	}
}

func TestBoardEmbed(t *testing.T) {
	running := GameServer{ID: "mc", Name: "Minecraft"}
	running.State = &struct {
		Running   bool      `json:"running"`
		Status    string    `json:"status"`
		StartedAt time.Time `json:"started_at"`
	}{Running: true, StartedAt: time.Unix(1791500000, 0)}
	running.Players = &struct {
		Online int      `json:"online"`
		Max    int      `json:"max"`
		Names  []string `json:"names"`
	}{2, 20, []string{"Steve", "__init__"}}

	e := boardEmbed("mc", running, true, nil)
	if e.Color != boardGreen || !strings.Contains(e.Description, "<t:1791500000:R>") || e.Fields[0].Value != "2 of 20 online\nSteve, \\_\\_init\\_\\_" {
		t.Errorf("running = %+v %+v", e, e.Fields[0])
	}
	busy := running
	busy.Busy = "restart"
	if e := boardEmbed("mc", busy, true, nil); e.Color != boardAmber || !strings.Contains(e.Description, "Restarting") {
		t.Errorf("busy = %+v", e)
	}
	if e := boardEmbed("mc", GameServer{}, false, nil); e.Color != boardRed || e.Title != "mc" {
		t.Errorf("missing = %+v", e)
	}
	if e := boardEmbed("mc", GameServer{}, false, errors.New("down")); !strings.Contains(e.Description, "can't reach") {
		t.Errorf("games down = %+v", e)
	}
	// Same content → same hash; any change → a new one (that's what triggers an edit).
	if hashEmbed(boardEmbed("mc", running, true, nil)) != hashEmbed(boardEmbed("mc", running, true, nil)) {
		t.Error("hash isn't stable")
	}
	if hashEmbed(boardEmbed("mc", running, true, nil)) == hashEmbed(boardEmbed("mc", busy, true, nil)) {
		t.Error("hash didn't change with the content")
	}
}

func TestConfigValidate(t *testing.T) {
	ok := DefaultConfig()
	with := func(f func(*Config)) Config { c := DefaultConfig(); f(&c); return c }
	tests := []struct {
		name    string
		c       Config
		wantErr bool
	}{
		{"default", ok, false},
		{"board", with(func(c *Config) { c.StatusBoards = []StatusBoard{{"minecraft", "123456789012345678"}} }), false},
		{"duplicate board", with(func(c *Config) {
			c.StatusBoards = []StatusBoard{{"minecraft", "123456789012345678"}, {"minecraft", "123456789012345678"}}
		}), true},
		{"bad channel", with(func(c *Config) { c.StatusBoards = []StatusBoard{{"minecraft", "general"}} }), true},
		{"bad server", with(func(c *Config) { c.StatusBoards = []StatusBoard{{"../x", "123456789012345678"}} }), true},
		{"bad user", with(func(c *Config) { c.Verified = []VerifiedUser{{"alice", "Alice"}} }), true},
		{"no label", with(func(c *Config) { c.Verified = []VerifiedUser{{alice, " "}} }), true},
		{"twice verified", with(func(c *Config) { c.Verified = []VerifiedUser{{alice, "A"}, {alice, "B"}} }), true},
		{"no names", with(func(c *Config) { c.Names = nil }), true},
		{"name case dup", with(func(c *Config) { c.Names = []string{"Jim", "jim"} }), true},
		{"name spaces", with(func(c *Config) { c.Names = []string{" Jim"} }), true},
		{"name too long", with(func(c *Config) { c.Names = []string{strings.Repeat("a", 33)} }), true},
	}
	for _, tt := range tests {
		if err := tt.c.Validate(); (err != nil) != tt.wantErr {
			t.Errorf("%s: Validate() = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestStore(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := s.Config()
	if cfg.Revision != 0 || len(cfg.Names) != len(DefaultNames) {
		t.Fatalf("fresh config = %+v", cfg)
	}
	// Requests: a repeat moves to the front; the list is bounded.
	for i := range maxRequests + 5 {
		id := "2000000000000000" + string(rune('a'+i%26)) + string(rune('a'+i/26))
		if err := s.AddRequest(AccessRequest{UserID: id, Name: "n"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.AddRequest(AccessRequest{UserID: bob, Name: "Bob"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddRequest(AccessRequest{UserID: bob, Name: "Bob"}); err != nil {
		t.Fatal(err)
	}
	if reqs := s.Requests(); len(reqs) != maxRequests || reqs[0].UserID != bob || reqs[1].UserID == bob {
		t.Errorf("requests: %d, first %s, second %s", len(reqs), reqs[0].UserID, reqs[1].UserID)
	}

	// Verifying someone drops their request; a stale revision is refused.
	cfg.Verified = []VerifiedUser{{UserID: bob, Name: "Bob"}}
	saved, err := s.SaveConfig(cfg)
	if err != nil || saved.Revision != 1 {
		t.Fatalf("save = %+v, %v", saved, err)
	}
	if !s.IsVerified(bob) || s.IsVerified(alice) {
		t.Error("IsVerified wrong")
	}
	for _, r := range s.Requests() {
		if r.UserID == bob {
			t.Error("verified user still requested")
		}
	}
	if _, err := s.SaveConfig(cfg); !errors.Is(err, ErrConflict) {
		t.Errorf("stale save: err = %v, want ErrConflict", err)
	}
	if err := s.SetStatusMessage("minecraft:1", "999"); err != nil {
		t.Fatal(err)
	}

	// Everything survives a reopen.
	re, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !re.IsVerified(bob) || re.Config().Revision != 1 || re.StatusMessage("minecraft:1") != "999" || len(re.Requests()) != maxRequests-1 {
		t.Errorf("after reopen: %+v %v", re.Config(), re.StatusMessages())
	}

	// A hand-edited file with a typo fails loudly.
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"nmes": []}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(dir); err == nil {
		t.Error("typo'd config loaded")
	}
}

func TestHandlers(t *testing.T) {
	b, _ := newTestBot(t)
	mux := http.NewServeMux()
	Register(mux, b.store, b, quiet)
	do := func(method, path, body string, header ...string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if len(header) == 2 {
			req.Header.Set(header[0], header[1])
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	if rec := do("GET", "/api/config", ""); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"revision":1`) {
		t.Errorf("GET config = %d %s", rec.Code, rec.Body)
	}
	body := `{"revision":1,"status_boards":[{"server_id":"minecraft","channel_id":"123456789012345678"}],"verified":[],"names":["Jim"]}`
	if rec := do("PUT", "/api/config", body); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"revision":2`) {
		t.Errorf("PUT config = %d %s", rec.Code, rec.Body)
	}
	if rec := do("PUT", "/api/config", body); rec.Code != 409 {
		t.Errorf("stale PUT = %d %s", rec.Code, rec.Body)
	}
	if rec := do("PUT", "/api/config", `{"revision":2,"names":[]}`); rec.Code != 400 {
		t.Errorf("invalid PUT = %d %s", rec.Code, rec.Body)
	}
	if rec := do("PUT", "/api/config", `{"revision":2,"nmes":[]}`); rec.Code != 400 {
		t.Errorf("typo PUT = %d %s", rec.Code, rec.Body)
	}

	rec := do("GET", "/api/bot", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"connected":false`) ||
		!strings.Contains(rec.Body.String(), "DISCORD_TOKEN") || !strings.Contains(rec.Body.String(), `"name":"Jim"`) ||
		!strings.Contains(rec.Body.String(), `"id":"minecraft"`) || strings.Contains(rec.Body.String(), `"old"`) {
		t.Errorf("GET bot = %d %s", rec.Code, rec.Body)
	}

	rec = do("GET", "/api/persona.gif", "")
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/gif" || rec.Header().Get("ETag") == "" {
		t.Fatalf("GET gif = %d %v", rec.Code, rec.Header())
	}
	if again := do("GET", "/api/persona.gif", "", "If-None-Match", rec.Header().Get("ETag")); again.Code != 304 {
		t.Errorf("revalidate = %d", again.Code)
	}

	if err := b.store.AddRequest(AccessRequest{UserID: bob, Name: "Bob"}); err != nil {
		t.Fatal(err)
	}
	if rec := do("DELETE", "/api/requests/"+bob, ""); rec.Code != 204 || len(b.store.Requests()) != 0 {
		t.Errorf("DELETE request = %d, %d left", rec.Code, len(b.store.Requests()))
	}
}

func TestGamesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/servers", func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "games.internal" {
			t.Errorf("Host = %q: hostroute needs a games.* host", r.Host)
		}
		w.Write([]byte(`[{"id":"mc","name":"Minecraft","game":"minecraft","state":{"running":true,"status":"running","started_at":"2026-10-09T18:00:00Z"},"players":{"online":1,"max":20,"names":["Steve"]},"busy":"restart"}]`))
	})
	mux.HandleFunc("POST /api/servers/{id}/restart", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"Minecraft is already restarting"}`))
	})
	mux.HandleFunc("PUT /api/servers/{id}/whitelist/{player}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"output":"Added ` + r.PathValue("player") + `","restart_needed":true}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	g := NewGamesHTTP(srv.URL + "/")
	ctx := context.Background()

	list, err := g.Servers(ctx)
	if err != nil || len(list) != 1 || list[0].Players.Names[0] != "Steve" || !list[0].State.Running || list[0].Busy != "restart" {
		t.Errorf("Servers = %+v, %v", list, err)
	}
	var ge *GamesError
	if err := g.Restart(ctx, "mc"); !errors.As(err, &ge) || ge.Status != 409 || ge.Message != "Minecraft is already restarting" {
		t.Errorf("Restart = %v", err)
	}
	if res, err := g.EditWhitelist(ctx, "vh", "76561198000000001", true); err != nil || !res.RestartNeeded {
		t.Errorf("EditWhitelist = %+v, %v", res, err)
	}
	if _, err := g.Whitelist(ctx, "mc"); !errors.As(err, &ge) || ge.Status != 404 {
		t.Errorf("unknown route = %v", err)
	}
	srv.Close()
	if _, err := g.Servers(ctx); !errors.Is(err, errGamesDown) {
		t.Errorf("down = %v", err)
	}
}
