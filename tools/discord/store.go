package discord

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// maxRequests bounds the access-request list: the newest few are all
// anyone needs, and a bound keeps every operation on it O(1) in practice.
const maxRequests = 20

// AccessRequest is someone who tried a verified-only command. The
// settings page lists them with a Verify button, so nobody has to look up
// a Discord user ID by hand.
type AccessRequest struct {
	UserID  string    `json:"user_id"`
	Name    string    `json:"name"`    // their display name when they asked
	Command string    `json:"command"` // what they tried, e.g. "restart"
	At      time.Time `json:"at"`
}

// state is what the bot writes for itself (state.json), never edited by
// people: which message is which status board, and who asked for access.
type state struct {
	// StatusMessages maps a board key ("<server>:<channel>") to the
	// message the bot edits, so a restart of Home Tools keeps editing the
	// same message instead of posting a new one.
	StatusMessages map[string]string `json:"status_messages"`
	Requests       []AccessRequest   `json:"requests"` // newest first, ≤ maxRequests
}

// Store holds the config and the bot's state in memory and writes changes
// through to JSON files under dir (ADR 0002). Safe for concurrent use.
type Store struct {
	dir string

	mu       sync.RWMutex
	cfg      Config
	verified map[string]struct{} // user IDs in cfg.Verified: O(1) checks per command
	st       state
}

// Open loads dir/config.json and dir/state.json. Missing files mean a
// fresh install (DefaultConfig, empty state); invalid ones fail the open.
func Open(dir string) (*Store, error) {
	cfg, err := jsonfile.Read[Config](filepath.Join(dir, "config.json"))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		cfg = DefaultConfig()
	case err != nil:
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("load %s: %w", filepath.Join(dir, "config.json"), err)
	}
	st, err := jsonfile.Read[state](filepath.Join(dir, "state.json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if st.StatusMessages == nil {
		st.StatusMessages = make(map[string]string)
	}
	s := &Store{dir: dir, st: st}
	s.setConfig(cfg)
	return s, nil
}

// setConfig swaps in cfg and rebuilds the lookup set. Caller holds s.mu
// (or is Open, before anyone else can see s).
func (s *Store) setConfig(cfg Config) {
	s.cfg = cfg
	s.verified = make(map[string]struct{}, len(cfg.Verified))
	for _, u := range cfg.Verified {
		s.verified[u.UserID] = struct{}{}
	}
}

// Config returns a copy of the config: callers may change it freely.
func (s *Store) Config() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c := s.cfg
	c.StatusBoards = slices.Clone(c.StatusBoards)
	c.Verified = slices.Clone(c.Verified)
	c.Names = slices.Clone(c.Names)
	return c
}

// SaveConfig validates cfg and writes it, if cfg.Revision is the current
// one (else ErrConflict). It returns the saved config, with the new
// revision. People who are now verified drop off the request list.
func (s *Store) SaveConfig(cfg Config) (Config, error) {
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg.Revision != s.cfg.Revision {
		return Config{}, fmt.Errorf("%w: someone saved these settings since you loaded them; reload and try again", ErrConflict)
	}
	cfg.Revision++
	if err := jsonfile.Write(filepath.Join(s.dir, "config.json"), cfg); err != nil {
		return Config{}, err
	}
	s.setConfig(cfg)

	kept := slices.DeleteFunc(slices.Clone(s.st.Requests), func(r AccessRequest) bool {
		_, ok := s.verified[r.UserID]
		return ok
	})
	if len(kept) != len(s.st.Requests) {
		if err := s.writeState(func(st *state) { st.Requests = kept }); err != nil {
			return cfg, err // the config itself is saved
		}
	}
	return cfg, nil
}

// IsVerified reports whether userID may use verified-only commands. O(1).
func (s *Store) IsVerified(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.verified[userID]
	return ok
}

// StatusMessage returns the message ID of the board with key, or "".
func (s *Store) StatusMessage(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.st.StatusMessages[key]
}

// StatusMessages returns a copy of every board key → message ID.
func (s *Store) StatusMessages() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.st.StatusMessages))
	for k, v := range s.st.StatusMessages {
		out[k] = v
	}
	return out
}

// SetStatusMessage remembers the message of board key ("" forgets it).
func (s *Store) SetStatusMessage(key, messageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.st.StatusMessages[key] == messageID {
		return nil // nothing changed: skip the disk write
	}
	return s.writeState(func(st *state) {
		if messageID == "" {
			delete(st.StatusMessages, key)
		} else {
			st.StatusMessages[key] = messageID
		}
	})
}

// AddRequest records that someone unverified asked for access. A repeat
// moves them to the front instead of adding a second entry; the list
// keeps the newest maxRequests (n ≤ 20, so the scan is effectively O(1)).
func (s *Store) AddRequest(r AccessRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.verified[r.UserID]; ok {
		return nil
	}
	reqs := slices.DeleteFunc(slices.Clone(s.st.Requests), func(o AccessRequest) bool { return o.UserID == r.UserID })
	reqs = slices.Insert(reqs, 0, r)
	if len(reqs) > maxRequests {
		reqs = reqs[:maxRequests]
	}
	return s.writeState(func(st *state) { st.Requests = reqs })
}

// Requests returns the access requests, newest first (never nil: it's
// sent as a JSON list).
func (s *Store) Requests() []AccessRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]AccessRequest{}, s.st.Requests...)
}

// DismissRequest removes userID's request, if any.
func (s *Store) DismissRequest(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	reqs := slices.DeleteFunc(slices.Clone(s.st.Requests), func(o AccessRequest) bool { return o.UserID == userID })
	if len(reqs) == len(s.st.Requests) {
		return nil
	}
	return s.writeState(func(st *state) { st.Requests = reqs })
}

// writeState applies change to a copy of the state, writes the copy, and
// only then keeps it: memory never runs ahead of the disk (decision #15:
// the lock is held across the write). Caller holds s.mu.
func (s *Store) writeState(change func(*state)) error {
	next := state{StatusMessages: make(map[string]string, len(s.st.StatusMessages)), Requests: s.st.Requests}
	for k, v := range s.st.StatusMessages {
		next.StatusMessages[k] = v
	}
	change(&next)
	if err := jsonfile.Write(filepath.Join(s.dir, "state.json"), next); err != nil {
		return err
	}
	s.st = next
	return nil
}
