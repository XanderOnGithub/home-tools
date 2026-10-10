package discord

import (
	"testing"
	"time"
)

func TestPickStatus(t *testing.T) {
	s := StatusFeature{Enabled: true, LiveGames: true, Phrases: []string{"I'm blobbing it", "{name} is blobbing it", "Blob o'clock"}}
	const h = 500000

	// Someone's playing and last hour wasn't a game hour: the game wins.
	if text, game := pickStatus(s, h, 0, "Watching Steve play Minecraft", "Jim"); !game || text != "Watching Steve play Minecraft" {
		t.Errorf("game hour = %q, %v", text, game)
	}
	// Last hour was a game hour: a phrase, even though someone's still on.
	if _, game := pickStatus(s, h, h-1, "Watching Steve play Minecraft", "Jim"); game {
		t.Error("game status two hours running")
	}
	// Live games off: always a phrase.
	off := s
	off.LiveGames = false
	if _, game := pickStatus(off, h, 0, "Watching Steve play Minecraft", "Jim"); game {
		t.Error("game status with live games off")
	}
	// {name} is filled in; phrases rotate without repeating back to back.
	prev := ""
	seen := map[string]bool{}
	for i := range int64(9) {
		text, _ := pickStatus(s, h+i, 0, "", "Jim")
		if text == prev {
			t.Fatalf("hour %d repeats %q", i, text)
		}
		if text == "{name} is blobbing it" {
			t.Fatal("{name} not filled in")
		}
		seen[text] = true
		prev = text
	}
	if !seen["Jim is blobbing it"] || len(seen) != 3 {
		t.Errorf("seen = %v", seen)
	}
}

func TestGameStatus(t *testing.T) {
	server := func(name string, online int, names ...string) GameServer {
		s := GameServer{Name: name}
		s.State = &struct {
			Running   bool      `json:"running"`
			Status    string    `json:"status"`
			StartedAt time.Time `json:"started_at"`
		}{Running: true}
		s.Players = &struct {
			Online int      `json:"online"`
			Max    int      `json:"max"`
			Names  []string `json:"names"`
		}{Online: online, Names: names}
		return s
	}
	tests := []struct {
		servers []GameServer
		want    string
	}{
		{nil, ""},
		{[]GameServer{server("Minecraft", 0)}, ""},
		{[]GameServer{server("Minecraft", 1, "Steve")}, "Watching Steve play Minecraft"},
		{[]GameServer{server("Valheim", 1)}, "Watching a blob play Valheim"},
		{[]GameServer{server("Minecraft", 1, "Steve"), server("Valheim", 3)}, "Watching 3 blobs play Valheim"},
		{[]GameServer{{Name: "Stopped"}}, ""},
	}
	for _, tt := range tests {
		if got := gameStatus(tt.servers); got != tt.want {
			t.Errorf("gameStatus = %q, want %q", got, tt.want)
		}
	}
}

func TestHourNumber(t *testing.T) {
	ny := time.FixedZone("NY", -4*3600)
	a := time.Date(2026, 10, 9, 23, 59, 0, 0, ny)
	if hourNumber(a.Add(2*time.Minute)) != hourNumber(a)+1 {
		t.Error("midnight doesn't start the next hour")
	}
	if hourNumber(time.Date(2026, 10, 9, 20, 30, 0, 0, ny))%24 != 20 {
		t.Error("hours don't follow the local clock")
	}
}

func TestStatusDefaultsAndRules(t *testing.T) {
	c := DefaultConfig()
	if s := c.Features.Status; s.Enabled || !s.LiveGames || len(s.Phrases) != len(DefaultPhrases) {
		t.Errorf("defaults = %+v", s)
	}
	for _, p := range DefaultPhrases {
		if len([]rune(p)) > maxPhraseLen {
			t.Errorf("default %q is too long", p)
		}
	}
	if err := c.Validate(); err != nil {
		t.Errorf("defaults invalid: %v", err)
	}
	// Someone emptied the list on purpose: it stays empty.
	c.Features.Status.Phrases = []string{}
	c.fillDefaults()
	if len(c.Features.Status.Phrases) != 0 {
		t.Error("an emptied list was refilled")
	}
	c.Features.Status.Enabled = true
	if c.Validate() == nil {
		t.Error("on with no phrases passed")
	}
	c.Features.Status.Phrases = []string{"Blob", "blob"}
	if c.Validate() == nil {
		t.Error("duplicate phrases passed")
	}
}
