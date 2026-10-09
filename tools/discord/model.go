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
)

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
	return Config{StatusBoards: []StatusBoard{}, Verified: []VerifiedUser{}, Names: append([]string(nil), DefaultNames...)}
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
	return nil
}
