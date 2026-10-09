package games

import (
	"slices"
	"testing"
	"time"
)

// Made-up player names only (never real players').
func TestActivityParse(t *testing.T) {
	at := func(min int) time.Time { return time.Date(2026, 10, 8, 18, min, 0, 0, time.UTC) }
	type line struct {
		min  int
		text string
	}
	tests := []struct {
		name       string
		game       Game
		lines      []line
		wantOnline []string
		wantEvents []string // "player kind", newest first
	}{
		{"minecraft join and leave", GameMinecraft, []line{
			{1, "[18:01:00] [Server thread/INFO]: Steve joined the game"},
			{2, "[18:02:00 INFO]: Alex joined the game"},
			{3, "[18:03:00] [Server thread/INFO]: Steve left the game"},
			{4, "[18:04:00] [Server thread/INFO]: <Alex> Steve joined the game"}, // chat, not a join
		}, []string{"Alex"}, []string{"Steve leave", "Alex join", "Steve join"}},
		{"minecraft renamed player", GameMinecraft, []line{
			{1, "[18:01:00] [Server thread/INFO]: Steve (formerly known as Alex) joined the game"},
		}, []string{"Steve"}, []string{"Steve join"}},
		{"valheim spawn, death, respawn, leave", GameValheim, []line{
			{1, "10/08/2026 18:01:00: Got character ZDOID from Ragnhild : 42:1"},
			{2, "10/08/2026 18:02:00: Got character ZDOID from Erik the Red : 77:1"},
			{3, "10/08/2026 18:03:00: Got character ZDOID from Ragnhild : 0:0"},  // died
			{4, "10/08/2026 18:04:00: Got character ZDOID from Ragnhild : 42:9"}, // respawned
			{5, "10/08/2026 18:05:00: Destroying abandoned non persistent zdo 77:3 owner 77"},
			{5, "10/08/2026 18:05:00: Destroying abandoned non persistent zdo 77:4 owner 77"},
		}, []string{"Ragnhild"}, []string{"Erik the Red leave", "Erik the Red join", "Ragnhild join"}},
		{"valheim nobody connected clears missed leaves", GameValheim, []line{
			{1, "10/08/2026 18:01:00: Got character ZDOID from Ragnhild : 42:1"},
			{9, "10/08/2026 18:09:00: Connections 0 ZDOS:1234  sent:0 recv:0"},
		}, nil, []string{"Ragnhild leave", "Ragnhild join"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &serverActivity{online: map[string]time.Time{}, peers: map[string]string{}}
			for _, l := range tt.lines {
				sa.parse(tt.game, at(l.min), l.text)
			}
			if got := sa.snapshot(); !slices.Equal(got, tt.wantOnline) && len(got)+len(tt.wantOnline) > 0 {
				t.Errorf("online = %v, want %v", got, tt.wantOnline)
			}
			var got []string
			for _, e := range sa.recent() {
				got = append(got, e.Player+" "+e.Kind)
			}
			if !slices.Equal(got, tt.wantEvents) {
				t.Errorf("events = %v, want %v", got, tt.wantEvents)
			}
		})
	}
}

func TestActivityKeepsNewest(t *testing.T) {
	sa := &serverActivity{online: map[string]time.Time{}}
	base := time.Date(2026, 10, 8, 18, 0, 0, 0, time.UTC)
	for i := range maxEvents + 10 {
		sa.join(base.Add(time.Duration(i)*time.Minute), "Steve")
		sa.leave(base.Add(time.Duration(i)*time.Minute+time.Second), "Steve")
	}
	ev := sa.recent()
	if len(ev) != maxEvents || !ev[0].At.After(ev[1].At) {
		t.Errorf("got %d events, newest first = %v", len(ev), ev[0].At.After(ev[1].At))
	}
}

func TestActivityUpdate(t *testing.T) {
	docker, _ := fakeDocker(t)
	srv := Server{ID: "minecraft", Game: GameMinecraft, Container: "mc"}
	run := State{Running: true, StartedAt: time.Date(2026, 10, 8, 18, 0, 0, 0, time.UTC)}
	var a activity

	// The fake answers the same line every time: read once, it counts once.
	for range 2 {
		online, events, err := a.update(t.Context(), docker, srv, run)
		if err != nil || !slices.Equal(online, []string{"Steve"}) || len(events) != 1 {
			t.Fatalf("update = %v %v %v", online, events, err)
		}
	}

	// A new run (restart): nobody's online yet, but the history stays.
	run.StartedAt = run.StartedAt.Add(time.Hour) // after the fake's line
	online, events, err := a.update(t.Context(), docker, srv, run)
	if err != nil || len(online) != 0 || len(events) != 1 {
		t.Errorf("after restart = %v %v %v", online, events, err)
	}
}
