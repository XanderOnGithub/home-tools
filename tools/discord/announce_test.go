package discord

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

const channel = "123456789012345678"

func TestPostValidate(t *testing.T) {
	poll := func(f func(*PostPoll)) *PostPoll {
		p := &PostPoll{PollQuestion: PollQuestion{Question: "Game night?", Answers: []string{"Fri", "Sat"}}, DurationHours: 24}
		if f != nil {
			f(p)
		}
		return p
	}
	tests := []struct {
		name string
		post Post
		ok   bool
	}{
		{"message", Post{ChannelID: channel, Message: "Server's back up"}, true},
		{"poll only", Post{ChannelID: channel, Poll: poll(nil)}, true},
		{"both", Post{ChannelID: channel, Message: "Vote!", Everyone: true, Poll: poll(nil)}, true},
		{"no channel", Post{Message: "hi"}, false},
		{"bad channel", Post{ChannelID: "general", Message: "hi"}, false},
		{"nothing to send", Post{ChannelID: channel, Message: "   "}, false},
		{"ping alone is nothing", Post{ChannelID: channel, Everyone: true}, false},
		{"too long", Post{ChannelID: channel, Message: strings.Repeat("a", 2001)}, false},
		{"too long with the ping", Post{ChannelID: channel, Everyone: true, Message: strings.Repeat("a", 1995)}, false},
		{"2000 is fine", Post{ChannelID: channel, Message: strings.Repeat("é", 2000)}, true},
		{"one answer", Post{ChannelID: channel, Poll: poll(func(p *PostPoll) { p.Answers = []string{"Fri"} })}, false},
		{"same answer twice", Post{ChannelID: channel, Poll: poll(func(p *PostPoll) { p.Answers = []string{"Fri", "fri"} })}, false},
		{"no question", Post{ChannelID: channel, Poll: poll(func(p *PostPoll) { p.Question = "" })}, false},
		{"no duration", Post{ChannelID: channel, Poll: poll(func(p *PostPoll) { p.DurationHours = 0 })}, false},
		{"over a week", Post{ChannelID: channel, Poll: poll(func(p *PostPoll) { p.DurationHours = 169 })}, false},
	}
	for _, tt := range tests {
		if err := tt.post.Validate(); (err == nil) != tt.ok {
			t.Errorf("%s: err = %v, want ok = %v", tt.name, err, tt.ok)
		}
	}
}

func TestPostMessage(t *testing.T) {
	// Plain text never pings, even with a pasted @here.
	m := Post{ChannelID: channel, Message: "  @here hi  "}.message()
	if m.Content != "@here hi" || m.AllowedMentions != noPings || m.Poll != nil {
		t.Errorf("plain = %+v", m)
	}
	// Everyone: prefixed, and only @everyone is allowed to ping.
	m = Post{ChannelID: channel, Message: "Game night", Everyone: true}.message()
	if m.Content != "@everyone Game night" {
		t.Errorf("everyone content = %q", m.Content)
	}
	if got := m.AllowedMentions.Parse; len(got) != 1 || got[0] != discordgo.AllowedMentionTypeEveryone {
		t.Errorf("everyone mentions = %v", got)
	}
	// A poll keeps its answers in order, its choice mode and duration.
	m = Post{ChannelID: channel, Poll: &PostPoll{
		PollQuestion:  PollQuestion{Question: "Which night?", Answers: []string{"Fri", "Sat", "Sun"}, Multi: true},
		DurationHours: 48,
	}}.message()
	if m.Content != "" || m.Poll == nil || m.Poll.Question.Text != "Which night?" || !m.Poll.AllowMultiselect || m.Poll.Duration != 48 {
		t.Fatalf("poll = %+v", m.Poll)
	}
	if len(m.Poll.Answers) != 3 || m.Poll.Answers[2].Media.Text != "Sun" {
		t.Errorf("answers = %+v", m.Poll.Answers)
	}
}

func TestSendHandler(t *testing.T) {
	b, _ := newTestBot(t)
	mux := http.NewServeMux()
	Register(mux, b.store, b, quiet)
	send := func(body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/send", strings.NewReader(body)))
		return rec
	}
	if rec := send(`{"channel_id":"` + channel + `"}`); rec.Code != 400 {
		t.Errorf("empty post = %d %s", rec.Code, rec.Body)
	}
	if rec := send(`{"channel_id":"` + channel + `","message":"hi","colour":"red"}`); rec.Code != 400 {
		t.Errorf("unknown field = %d %s", rec.Code, rec.Body)
	}
	// Valid, but the test bot isn't connected.
	rec := send(`{"channel_id":"` + channel + `","message":"hi"}`)
	if rec.Code != 409 || !strings.Contains(rec.Body.String(), "isn't connected") {
		t.Errorf("offline = %d %s", rec.Code, rec.Body)
	}
}
