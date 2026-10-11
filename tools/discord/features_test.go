package discord

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	mrand "math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testPoll() PollFeature {
	p := DefaultConfig().Features.Poll
	p.Enabled, p.ChannelID = true, "123456789012345678"
	p.Packs = []PollPack{
		{Name: "Food", Enabled: true, Questions: []PollQuestion{
			{Question: "Best pizza topping?", Answers: []string{"Pepperoni", "Pineapple"}},
			{Question: "Morning or night?", Answers: []string{"Morning", "Night"}},
		}},
		{Name: "Games", Enabled: true, Questions: []PollQuestion{
			{Question: "Next game?", Answers: []string{"Valheim", "Minecraft", "Something new"}},
		}},
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
		{"off without channel or packs", with(func(p *PollFeature) { p.Enabled, p.ChannelID, p.Packs = false, "", []PollPack{} }), false},
		{"on without channel", with(func(p *PollFeature) { p.ChannelID = "" }), true},
		{"on without packs", with(func(p *PollFeature) { p.Packs = []PollPack{} }), true},
		{"on with every pack off", with(func(p *PollFeature) { p.Packs[0].Enabled, p.Packs[1].Enabled = false, false }), true},
		{"on with one pack off", with(func(p *PollFeature) { p.Packs[0].Enabled = false }), false},
		{"pack without a name", with(func(p *PollFeature) { p.Packs[0].Name = " " }), true},
		{"pack name twice", with(func(p *PollFeature) { p.Packs[1].Name = "food" }), true},
		{"every 8 days", with(func(p *PollFeature) { p.EveryDays = 8 }), true},
		{"bad time", with(func(p *PollFeature) { p.PostAt = "6pm" }), true},
		{"24:00", with(func(p *PollFeature) { p.PostAt = "24:00" }), true},
		{"open 0 h", with(func(p *PollFeature) { p.DurationHours = 0 }), true},
		{"one answer", with(func(p *PollFeature) { p.Packs[0].Questions[0].Answers = []string{"Yes"} }), true},
		{"answer too long", with(func(p *PollFeature) { p.Packs[0].Questions[0].Answers[0] = strings.Repeat("a", 56) }), true},
		{"answer twice", with(func(p *PollFeature) { p.Packs[0].Questions[0].Answers = []string{"Yes", "yes"} }), true},
		{"same question, other answers", with(func(p *PollFeature) { p.Packs[0].Questions[1].Question = "Best pizza topping?" }), false},
		{"same poll twice", with(func(p *PollFeature) {
			p.Packs[0].Questions[1] = PollQuestion{Question: "best pizza topping?", Answers: []string{"pepperoni", "pineapple"}}
		}), true},
		{"same poll in two packs", with(func(p *PollFeature) { p.Packs[1].Questions = append(p.Packs[1].Questions, p.Packs[0].Questions[0]) }), true},
		{"empty question", with(func(p *PollFeature) { p.Packs[0].Questions[0].Question = " " }), true},
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

}

func TestNextQuestion(t *testing.T) {
	p := testPoll() // 3 questions on, in two packs
	p.Packs = append(p.Packs, PollPack{Name: "Off", Questions: []PollQuestion{{Question: "Never?", Answers: []string{"A", "B"}}}})
	rng := mrand.New(mrand.NewPCG(1, 2))

	// Over many rounds: each question once per round, never the same twice
	// in a row (not even across rounds), never one from a pack that's off.
	asked, last := map[string]string(nil), ""
	for round := range 50 {
		got := map[string]bool{}
		for i := range 3 {
			q, next := nextQuestion(p, asked, last, fmt.Sprintf("r%d-%d", round, i), rng.IntN)
			if q.Question == "Never?" {
				t.Fatal("asked a question from a pack that's off")
			}
			if q.key() == last {
				t.Fatalf("round %d: %q twice in a row", round, q.Question)
			}
			if got[q.key()] {
				t.Fatalf("round %d: %q twice in one round", round, q.Question)
			}
			got[q.key()] = true
			asked, last = next, q.key()
		}
	}

	// The record keeps packs that are off, and drops questions that are gone.
	off := p.Packs[2].Questions[0].key()
	asked = map[string]string{off: "2026-01-01", "gone\x1fA\x1fB": "2026-01-01"}
	_, next := nextQuestion(p, asked, "", "2026-10-11", rng.IntN)
	if next[off] != "2026-01-01" || len(next) != 2 {
		t.Errorf("record = %v", next)
	}

	// A single question repeats (there's nothing else).
	one := PollFeature{Packs: []PollPack{{Name: "One", Enabled: true, Questions: p.Packs[1].Questions}}}
	q, asked := nextQuestion(one, nil, "", "d1", rng.IntN)
	if again, _ := nextQuestion(one, asked, q.key(), "d2", rng.IntN); again.key() != q.key() {
		t.Errorf("single question: %q", again.Question)
	}
}

func TestDefaultPollPacks(t *testing.T) {
	p := DefaultConfig().Features.Poll
	p.Enabled, p.ChannelID = true, "123456789012345678"
	if err := p.validate(); err != nil {
		t.Fatalf("default packs don't fit Discord's limits: %v", err)
	}
	if len(p.Packs) == 0 || len(p.questions()) < 20 {
		t.Errorf("packs = %d, questions on = %d", len(p.Packs), len(p.questions()))
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
	if p.Enabled || p.EveryDays != 2 || p.PostAt != "18:00" || p.DurationHours != 24 || len(p.Packs) != len(DefaultPollPacks()) {
		t.Errorf("defaults = %+v", p)
	}

	// A config from before packs: its questions become a pack of their own.
	old = `{"revision": 4, "status_boards": [], "verified": [], "names": ["Jim"],
		"features": {"poll": {"every_days": 3, "post_at": "20:00", "duration_hours": 24,
			"questions": [{"question": "Tea or coffee?", "answers": ["Tea", "Coffee"]}]}}}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}
	if s, err = Open(dir); err != nil {
		t.Fatal(err)
	}
	p = s.Config().Features.Poll
	mine := p.Packs[len(p.Packs)-1]
	if p.Questions != nil || mine.Name != "My questions" || !mine.Enabled || mine.Questions[0].Question != "Tea or coffee?" || p.EveryDays != 3 {
		t.Errorf("migrated = %+v", p)
	}
	// Saving writes packs only, and an emptied pack list stays empty.
	cfg := s.Config()
	cfg.Features.Poll.Packs = []PollPack{}
	if _, err := s.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if s, err = Open(dir); err != nil {
		t.Fatal(err)
	}
	if n := len(s.Config().Features.Poll.Packs); n != 0 {
		t.Errorf("removed packs came back: %d", n)
	}
}

func TestAskedPolls(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetPollPosted("2026-10-11", "a", map[string]string{"a": "2026-10-11", "b": "2026-10-09"}); err != nil {
		t.Fatal(err)
	}
	if err := s.ForgetAsked([]string{"b", "nope"}); err != nil {
		t.Fatal(err)
	}
	if s, err = Open(dir); err != nil { // it's on disk
		t.Fatal(err)
	}
	asked, last := s.AskedPolls()
	if last != "a" || len(asked) != 1 || asked["a"] != "2026-10-11" || s.LastPoll() != "2026-10-11" {
		t.Errorf("asked = %v, last = %q, last poll = %q", asked, last, s.LastPoll())
	}
}
