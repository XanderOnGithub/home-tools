package discord

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testPoll() PollFeature {
	p := DefaultConfig().Features.Poll
	p.Enabled, p.ChannelID = true, "123456789012345678"
	p.Questions = []PollQuestion{
		{Question: "Best pizza topping?", Answers: []string{"Pepperoni", "Pineapple"}},
		{Question: "Morning or night?", Answers: []string{"Morning", "Night"}},
		{Question: "Next game?", Answers: []string{"Valheim", "Minecraft", "Something new"}},
	}
	return p
}

func TestPollValidate(t *testing.T) {
	with := func(f func(*PollFeature)) PollFeature { p := testPoll(); f(&p); return p }
	tests := []struct {
		name    string
		p       PollFeature
		wantErr bool
	}{
		{"valid", testPoll(), false},
		{"off without channel or questions", with(func(p *PollFeature) { p.Enabled, p.ChannelID, p.Questions = false, "", nil }), false},
		{"on without channel", with(func(p *PollFeature) { p.ChannelID = "" }), true},
		{"on without questions", with(func(p *PollFeature) { p.Questions = nil }), true},
		{"every 8 days", with(func(p *PollFeature) { p.EveryDays = 8 }), true},
		{"bad time", with(func(p *PollFeature) { p.PostAt = "6pm" }), true},
		{"24:00", with(func(p *PollFeature) { p.PostAt = "24:00" }), true},
		{"open 0 h", with(func(p *PollFeature) { p.DurationHours = 0 }), true},
		{"one answer", with(func(p *PollFeature) { p.Questions[0].Answers = []string{"Yes"} }), true},
		{"answer too long", with(func(p *PollFeature) { p.Questions[0].Answers[0] = strings.Repeat("a", 56) }), true},
		{"answer twice", with(func(p *PollFeature) { p.Questions[0].Answers = []string{"Yes", "yes"} }), true},
		{"question twice", with(func(p *PollFeature) { p.Questions[1].Question = "best pizza topping?" }), true},
		{"empty question", with(func(p *PollFeature) { p.Questions[0].Question = " " }), true},
	}
	for _, tt := range tests {
		if err := tt.p.validate(); (err != nil) != tt.wantErr {
			t.Errorf("%s: validate() = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestNextPollSlot(t *testing.T) {
	p := testPoll() // every 2 days at 18:00
	loc := time.FixedZone("test", -4*3600)
	// 2026-10-10 is day 20736 (even): a poll day; the 9th is not.
	tests := []struct {
		name string
		now  time.Time
		last string
		want string // the slot's time, local
	}{
		{"before today's slot", time.Date(2026, 10, 10, 9, 0, 0, 0, loc), "", "2026-10-10 18:00"},
		{"overdue today (bot was down)", time.Date(2026, 10, 10, 21, 0, 0, 0, loc), "2026-10-08", "2026-10-10 18:00"},
		{"today's done", time.Date(2026, 10, 10, 21, 0, 0, 0, loc), "2026-10-10", "2026-10-12 18:00"},
		{"not a poll day", time.Date(2026, 10, 9, 12, 0, 0, 0, loc), "2026-10-08", "2026-10-10 18:00"},
		{"posted early (now button)", time.Date(2026, 10, 9, 12, 0, 0, 0, loc), "2026-10-10", "2026-10-12 18:00"},
	}
	for _, tt := range tests {
		got := nextPollSlot(tt.now, p, tt.last).at.Format("2006-01-02 15:04")
		if got != tt.want {
			t.Errorf("%s: next = %s, want %s", tt.name, got, tt.want)
		}
	}

	// Every question before any repeats, and never the same twice in a row.
	p.EveryDays = 1
	var prev string
	seen := map[string]int{}
	for i := range 30 {
		day := time.Date(2026, 10, 10+i, 12, 0, 0, 0, loc)
		q := questionFor(p, nextPollSlot(day, p, day.AddDate(0, 0, -1).Format(time.DateOnly))).Question
		if q == prev {
			t.Fatalf("day %d repeats %q", i, q)
		}
		seen[q]++
		prev = q
	}
	for _, q := range p.Questions {
		if seen[q.Question] != 10 {
			t.Errorf("%q asked %d times in 30 days, want 10", q.Question, seen[q.Question])
		}
	}
}

func TestPollPostNowAndLoop(t *testing.T) {
	b, _ := newTestBot(t)
	if err := b.PostPollNow(); err == nil || !strings.Contains(err.Error(), "turn the poll on") {
		t.Errorf("poll off: err = %v", err)
	}
	cfg := b.store.Config()
	cfg.Features.Poll = testPoll()
	if _, err := b.store.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if err := b.PostPollNow(); err != errOffline {
		t.Errorf("offline: err = %v", err)
	}
	if b.NextPoll().IsZero() {
		t.Error("NextPoll is zero with the poll on")
	}
	// Offline, the loop keeps waiting instead of spinning, and stops with ctx.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	b.runPolls(ctx)
	if b.store.LastPoll() != "" {
		t.Error("a poll was recorded while offline")
	}
}

func TestFeatureCommands(t *testing.T) {
	names := func(f Features) string {
		var out []string
		for _, c := range commands(f) {
			out = append(out, c.Name)
		}
		return strings.Join(out, ",")
	}
	if got := names(Features{}); got != "sens,restart,whitelist" {
		t.Errorf("all off: %s", got)
	}
	if got := names(Features{Blob: BlobFeature{Enabled: true}}); got != "sens,restart,whitelist,blob" {
		t.Errorf("blob on: %s", got)
	}
}

func TestBlobCommand(t *testing.T) {
	b, _ := newTestBot(t)
	blob := invocation{command: "blob", opts: map[string]any{"name": "Jim", "color": "blue"}}

	if r := run(b, blob); !strings.Contains(r.last(), "don't know") {
		t.Errorf("off: %q", r.last())
	}
	cfg := b.store.Config()
	cfg.Features.Blob.Enabled = true
	if _, err := b.store.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	r := run(b, blob)
	f := r.sent[0].file
	if f == nil || f.name != "jim.png" || r.sent[0].private {
		t.Fatalf("blob = %+v", r.sent[0])
	}
	img, err := png.Decode(bytes.NewReader(f.data))
	if err != nil || img.Bounds().Dx() != blobPNGPx {
		t.Fatalf("png: %v, %v", img.Bounds(), err)
	}
	if _, _, _, a := img.At(0, 0).RGBA(); a != 0 {
		t.Error("corner isn't transparent: the blob needs margin and no background")
	}

	blob.opts["animated"] = true
	if r := run(b, blob); r.sent[0].file == nil || r.sent[0].file.contentType != "image/gif" {
		t.Errorf("animated = %+v", r.sent[0])
	}
	blob.opts["color"] = "pink"
	if r := run(b, blob); !r.sent[0].private {
		t.Error("bad color should answer privately")
	}
	if got := fileName("Zoë the 2nd!"); got != "zoë-the-2nd-" {
		t.Errorf("fileName = %q", got)
	}
}

func TestOldConfigGetsDefaults(t *testing.T) {
	dir := t.TempDir()
	// A config from before features existed.
	old := `{"revision": 3, "status_boards": [], "verified": [], "names": ["Jim"]}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	p := s.Config().Features.Poll
	if p.Enabled || p.EveryDays != 2 || p.PostAt != "18:00" || p.DurationHours != 24 || p.Questions == nil {
		t.Errorf("defaults = %+v", p)
	}
}
