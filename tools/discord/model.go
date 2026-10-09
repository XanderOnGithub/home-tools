// Package discord is the household's Discord bot and its settings
// (decision #42): sensitivity conversions, live game server statuses,
// whitelisting and restarts for verified people, and a blob persona with
// a new name and color every day. The bot dials out to Discord's gateway,
// so nothing is exposed; its settings page is LAN-only like every tool.
package discord

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// Config is everything the settings page edits, stored as config.json.
// The file is the source of truth (hand-editable); the bot reads it on
// every use, so changes apply without a restart.
type Config struct {
	// Revision counts saves. A save must carry the revision it was based
	// on, so a stale tab can't silently undo someone else's change
	// (optimistic concurrency, like an HTTP ETag).
	Revision     int            `json:"revision"`
	StatusBoards []StatusBoard  `json:"status_boards"`
	Verified     []VerifiedUser `json:"verified"`
	Names        []string       `json:"names"` // the persona's daily names
	Features     Features       `json:"features"`
}

// Features are the optional parts of the bot, each with its own switch
// (decision #44). Off = its commands disappear from Discord and its jobs
// stop, at once, without a restart. A typed struct rather than a map of
// flags: each feature has settings of its own, validated like everything
// else. New features start off (the zero value).
type Features struct {
	Poll PollFeature `json:"poll"`
	Blob BlobFeature `json:"blob"`
}

// PollFeature posts a poll from Questions every EveryDays days at PostAt
// (local time) in ChannelID, as a native Discord poll.
type PollFeature struct {
	Enabled       bool           `json:"enabled"`
	ChannelID     string         `json:"channel_id"`
	EveryDays     int            `json:"every_days"`     // 1–7
	PostAt        string         `json:"post_at"`        // "18:00", local time
	DurationHours int            `json:"duration_hours"` // how long votes are open, 1–168
	Questions     []PollQuestion `json:"questions"`
}

// PollQuestion is one poll: a question and its answers (Discord's limits:
// 300 and 55 characters, 2–10 answers).
type PollQuestion struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
	Multi    bool     `json:"multi,omitempty"` // people may pick several
}

// BlobFeature is /blob: anyone can make a blob of their own.
type BlobFeature struct {
	Enabled bool `json:"enabled"`
}

// StatusBoard puts a game server's live status in a Discord channel: one
// message the bot keeps editing.
type StatusBoard struct {
	ServerID  string `json:"server_id"` // games tool server ID
	ChannelID string `json:"channel_id"`
}

// VerifiedUser may restart servers and edit whitelists. Name is only a
// label for the settings page; the ID is what counts.
type VerifiedUser struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

// Limits: generous for a household, small enough that every lookup can
// stay a plain scan or map.
const (
	maxBoards   = 50
	maxVerified = 200
	maxNames    = 500
	maxNameLen  = 32 // Discord's nickname limit
	maxLabelLen = 100

	maxQuestions   = 500
	maxQuestionLen = 300 // Discord's poll limits
	maxAnswerLen   = 55
	maxAnswers     = 10
)

// postAt is a local time of day, "HH:MM" (24 h).
var postAt = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// ErrInvalid marks data that breaks a rule (HTTP 400).
var ErrInvalid = errors.New("invalid")

// ErrConflict means a save was based on an old revision (HTTP 409).
var ErrConflict = errors.New("changed elsewhere")

// snowflake is a Discord ID: a 64-bit number, 17–20 digits today.
var snowflake = regexp.MustCompile(`^[0-9]{17,20}$`)

// DefaultNames are plain, friendly names: the bot is a blob called Jim.
var DefaultNames = []string{
	"Jim", "Larry", "Jerry", "Gary", "Steve", "Kevin", "Dave", "Phil", "Doug",
	"Barry", "Carl", "Frank", "Greg", "Ted", "Ron", "Bob", "Hank", "Walt",
	"Stan", "Earl", "Norm", "Lou", "Ed", "Pam", "Linda", "Barb", "Sue",
	"Deb", "Carol", "Jan", "Kim", "Pat", "Gus", "Bert", "Ernie", "Vern",
}

// DefaultConfig is what a fresh install starts with: no boards, nobody
// verified, the default names.
func DefaultConfig() Config {
	c := Config{StatusBoards: []StatusBoard{}, Verified: []VerifiedUser{}, Names: append([]string(nil), DefaultNames...)}
	c.fillDefaults()
	return c
}

// fillDefaults sets settings that a config from before they existed (or
// a hand-written one) leaves empty. Everything else keeps its zero value.
func (c *Config) fillDefaults() {
	p := &c.Features.Poll
	if p.EveryDays == 0 {
		p.EveryDays = 2 // every other day: often enough to talk, rare enough not to annoy
	}
	if p.PostAt == "" {
		p.PostAt = "18:00"
	}
	if p.DurationHours == 0 {
		p.DurationHours = 24
	}
	if p.Questions == nil {
		p.Questions = []PollQuestion{}
	}
}

// Validate checks c's own rules. Whether a board's server exists is not
// checked: the games tool may be down, and the page shows unknown ones.
func (c Config) Validate() error {
	if len(c.StatusBoards) > maxBoards || len(c.Verified) > maxVerified || len(c.Names) > maxNames {
		return fmt.Errorf("%w: too many entries (at most %d boards, %d people, %d names)", ErrInvalid, maxBoards, maxVerified, maxNames)
	}
	// Sets catch duplicates in one pass: O(n) instead of comparing pairs.
	boards := make(map[StatusBoard]struct{}, len(c.StatusBoards))
	for _, b := range c.StatusBoards {
		switch {
		case !jsonfile.ValidID(b.ServerID):
			return fmt.Errorf("%w status board: bad server ID %q", ErrInvalid, b.ServerID)
		case !snowflake.MatchString(b.ChannelID):
			return fmt.Errorf("%w status board: bad channel ID %q", ErrInvalid, b.ChannelID)
		}
		if _, dup := boards[b]; dup {
			return fmt.Errorf("%w: %s is already shown in that channel", ErrInvalid, b.ServerID)
		}
		boards[b] = struct{}{}
	}
	people := make(map[string]struct{}, len(c.Verified))
	for _, u := range c.Verified {
		switch {
		case !snowflake.MatchString(u.UserID):
			return fmt.Errorf("%w: %q isn't a Discord user ID (17–20 digits)", ErrInvalid, u.UserID)
		case strings.TrimSpace(u.Name) == "" || utf8.RuneCountInString(u.Name) > maxLabelLen:
			return fmt.Errorf("%w verified user %s: name must be 1–%d characters", ErrInvalid, u.UserID, maxLabelLen)
		}
		if _, dup := people[u.UserID]; dup {
			return fmt.Errorf("%w: %s is verified twice", ErrInvalid, u.Name)
		}
		people[u.UserID] = struct{}{}
	}
	if len(c.Names) == 0 {
		return fmt.Errorf("%w: the bot needs at least one name", ErrInvalid)
	}
	names := make(map[string]struct{}, len(c.Names))
	for _, n := range c.Names {
		if n != strings.TrimSpace(n) || n == "" || utf8.RuneCountInString(n) > maxNameLen {
			return fmt.Errorf("%w name %q: 1–%d characters, no spaces around it", ErrInvalid, n, maxNameLen)
		}
		key := strings.ToLower(n)
		if _, dup := names[key]; dup {
			return fmt.Errorf("%w: %q is listed twice", ErrInvalid, n)
		}
		names[key] = struct{}{}
	}
	return c.Features.Poll.validate()
}

// validate checks the poll's settings. Turning it on needs a channel and
// at least one question; the questions themselves must always fit
// Discord's limits, so a saved list can always be posted.
func (p PollFeature) validate() error {
	switch {
	case p.EveryDays < 1 || p.EveryDays > 7:
		return fmt.Errorf("%w poll: every 1–7 days, got %d", ErrInvalid, p.EveryDays)
	case !postAt.MatchString(p.PostAt):
		return fmt.Errorf("%w poll: time must be HH:MM, got %q", ErrInvalid, p.PostAt)
	case p.DurationHours < 1 || p.DurationHours > 168:
		return fmt.Errorf("%w poll: open for 1–168 hours, got %d", ErrInvalid, p.DurationHours)
	case p.ChannelID != "" && !snowflake.MatchString(p.ChannelID):
		return fmt.Errorf("%w poll: bad channel ID %q", ErrInvalid, p.ChannelID)
	case p.Enabled && p.ChannelID == "":
		return fmt.Errorf("%w: pick a channel for the poll before turning it on", ErrInvalid)
	case p.Enabled && len(p.Questions) == 0:
		return fmt.Errorf("%w: add a question before turning the poll on", ErrInvalid)
	case len(p.Questions) > maxQuestions:
		return fmt.Errorf("%w: at most %d poll questions", ErrInvalid, maxQuestions)
	}
	seen := make(map[string]struct{}, len(p.Questions))
	for _, q := range p.Questions {
		text := strings.TrimSpace(q.Question)
		switch {
		case text == "" || utf8.RuneCountInString(q.Question) > maxQuestionLen:
			return fmt.Errorf("%w poll question %q: 1–%d characters", ErrInvalid, q.Question, maxQuestionLen)
		case len(q.Answers) < 2 || len(q.Answers) > maxAnswers:
			return fmt.Errorf("%w poll %q: 2–%d answers", ErrInvalid, q.Question, maxAnswers)
		}
		key := strings.ToLower(text)
		if _, dup := seen[key]; dup {
			return fmt.Errorf("%w: the poll %q is listed twice", ErrInvalid, q.Question)
		}
		seen[key] = struct{}{}
		answers := make(map[string]struct{}, len(q.Answers))
		for _, a := range q.Answers {
			if strings.TrimSpace(a) == "" || utf8.RuneCountInString(a) > maxAnswerLen {
				return fmt.Errorf("%w poll %q: answers must be 1–%d characters, got %q", ErrInvalid, q.Question, maxAnswerLen, a)
			}
			if _, dup := answers[strings.ToLower(a)]; dup {
				return fmt.Errorf("%w poll %q: the answer %q is there twice", ErrInvalid, q.Question, a)
			}
			answers[strings.ToLower(a)] = struct{}{}
		}
	}
	return nil
}
