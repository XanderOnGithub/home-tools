package discord

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestMergeMembers(t *testing.T) {
	m := func(id, user, global, nick string, bot bool) *discordgo.Member {
		return &discordgo.Member{Nick: nick, User: &discordgo.User{ID: id, Username: user, GlobalName: global, Bot: bot}}
	}
	lists := [][]*discordgo.Member{
		{m("1", "mayabuilds", "Maya", "", false), m("2", "samtheman", "", "", false), m("9", "homebot", "", "", true)},
		{m("1", "mayabuilds", "Maya", "Captain Maya", false), m("3", "jake", "Jake", "cousin jake", false), nil},
	}
	got := mergeMembers(lists, 10)
	want := []Member{
		{UserID: "3", Name: "cousin jake", Username: "jake"},    // nickname wins
		{UserID: "1", Name: "Maya", Username: "mayabuilds"},     // first server's name, once
		{UserID: "2", Name: "samtheman", Username: "samtheman"}, // no display name: the username
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
	if got := mergeMembers(lists, 2); len(got) != 2 {
		t.Errorf("limit 2: got %d", len(got))
	}
}

func TestFaces(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	fetches := 0
	f := faces{
		now: func() time.Time { return now },
		urlF: func(_ context.Context, id string) (string, error) {
			switch id {
			case "1":
				return "https://cdn.discordapp.com/avatars/1/abc.png?size=64", nil
			case "evil":
				return "https://example.com/x.png", nil
			}
			return "", errors.New("unknown user")
		},
		get: func(_ context.Context, url string) ([]byte, string, error) {
			fetches++
			return []byte("png"), "image/png", nil
		},
	}
	ctx := context.Background()
	if img, mime, err := f.avatar(ctx, "1"); err != nil || string(img) != "png" || mime != "image/png" {
		t.Fatalf("first = %q %q %v", img, mime, err)
	}
	f.avatar(ctx, "1")
	if fetches != 1 {
		t.Errorf("cached: fetched %d times, want 1", fetches)
	}
	now = now.Add(25 * time.Hour)
	f.avatar(ctx, "1")
	if fetches != 2 {
		t.Errorf("after a day: fetched %d times, want 2", fetches)
	}
	// Failures are remembered for an hour, and only Discord's CDN is used.
	if _, _, err := f.avatar(ctx, "2"); !errors.Is(err, errNoFace) {
		t.Errorf("unknown = %v", err)
	}
	if _, _, err := f.avatar(ctx, "evil"); !errors.Is(err, errNoFace) || fetches != 2 {
		t.Errorf("other host: err %v, fetches %d", err, fetches)
	}
}

func TestMemberHandlers(t *testing.T) {
	b, _ := newTestBot(t)
	mux := http.NewServeMux()
	Register(mux, b.store, b, quiet)
	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		return rec
	}
	if rec := get("/api/members?q="); rec.Code != 400 {
		t.Errorf("empty search = %d", rec.Code)
	}
	if rec := get("/api/members?q=" + strings.Repeat("a", 33)); rec.Code != 400 {
		t.Errorf("long search = %d", rec.Code)
	}
	if rec := get("/api/members?q=jo"); rec.Code != 409 || !strings.Contains(rec.Body.String(), "isn't connected") {
		t.Errorf("offline search = %d %s", rec.Code, rec.Body)
	}
	if rec := get("/api/members/abc/avatar"); rec.Code != 400 {
		t.Errorf("bad id = %d", rec.Code)
	}
	if rec := get("/api/members/123456789012345678/avatar"); rec.Code != 404 {
		t.Errorf("offline avatar = %d", rec.Code)
	}
}
