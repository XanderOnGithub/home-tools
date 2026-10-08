package fitness

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// Store holds all fitness data in memory and writes changes through to
// JSON files under dir. Safe for concurrent use; always use it via *Store
// (a copied mutex protects nothing).
type Store struct {
	dir string

	mu        sync.RWMutex
	exercises map[string]Exercise  // by ID
	routines  map[string]Routine   // by ID
	sessions  map[string][]Session // by user ID, sorted oldest → newest
}

// Open loads all fitness data under dir into memory. Missing folders mean
// no data yet; any unreadable or invalid file fails the whole open.
func Open(dir string) (*Store, error) {
	// Load Exercises
	exercises, err := loadDir(filepath.Join(dir, "exercises"), func(e Exercise) string { return e.ID })
	if err != nil {
		return nil, err
	}

	// Load Routines
	routines, err := loadDir(filepath.Join(dir, "routines"), func(r Routine) string { return r.ID })
	if err != nil {
		return nil, err
	}

	// List users folder
	userDirs, err := os.ReadDir(filepath.Join(dir, "users"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	// Build the store
	s := &Store{
		dir:       dir,
		exercises: make(map[string]Exercise, len(exercises)),
		routines:  make(map[string]Routine, len(routines)),
		sessions:  make(map[string][]Session, len(userDirs)),
	}

	// Fill the maps
	for _, ex := range exercises {
		s.exercises[ex.ID] = ex
	}
	for _, r := range routines {
		if err := s.checkRoutineRefs(r); err != nil {
			return nil, fmt.Errorf("load routine %s: %w", r.ID, err)
		}
		s.routines[r.ID] = r
	}

	// Load each user's sessions; ReadDir order keeps them sorted.
	for _, u := range userDirs {
		if !u.IsDir() {
			continue
		}
		sessDir := filepath.Join(dir, "users", u.Name(), "sessions")
		sessions, err := loadDir(sessDir, func(s Session) string { return s.ID })
		if err != nil {
			return nil, err
		}
		// Saves write to users/<UserID>/, so a mismatch would move the
		// session to another user's folder on its next edit.
		for _, sess := range sessions {
			path := filepath.Join(sessDir, sess.ID+".json")
			if sess.UserID != u.Name() {
				return nil, fmt.Errorf("load %s: user_id %q does not match folder %q", path, sess.UserID, u.Name())
			}
			if err := s.checkSessionRefs(sess); err != nil {
				return nil, fmt.Errorf("load %s: %w", path, err)
			}
		}
		s.sessions[u.Name()] = sessions
	}

	return s, nil
}

// SessionIDLayout formats a session's start time as its ID and filename.
// Always UTC and fixed-width, so sorting IDs as strings sorts them by time.
const SessionIDLayout = "2006-01-02T15-04-05Z"

// NewSessionID returns the ID for a session starting at t.
func NewSessionID(t time.Time) string {
	return t.UTC().Format(SessionIDLayout)
}

// SaveSession validates sess and writes it to disk, then to memory.
// An empty ID means a new session; its ID is derived from StartedAt.
// It returns the saved session (with its ID filled in).
func (s *Store) SaveSession(sess Session) (Session, error) {
	if sess.ID == "" {
		sess.ID = NewSessionID(sess.StartedAt)
	}
	if err := sess.Validate(); err != nil {
		return Session{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkSessionRefs(sess); err != nil {
		return Session{}, err
	}

	path := filepath.Join(s.dir, "users", sess.UserID, "sessions", sess.ID+".json")
	if err := jsonfile.Write(path, sess); err != nil {
		return Session{}, err
	}

	// Disk succeeded; now memory. Scan backwards from the newest session:
	// edits almost always target the latest one, so this usually stops
	// after one step. The same scan finds the insert point for a session
	// logged after the fact, keeping the slice sorted.
	list := s.sessions[sess.UserID]
	i := len(list)
	for i > 0 && list[i-1].ID >= sess.ID {
		i--
	}
	if i < len(list) && list[i].ID == sess.ID {
		list[i] = sess // update in place
	} else {
		list = slices.Insert(list, i, sess) // i == len(list) is a plain append
	}
	s.sessions[sess.UserID] = list
	return sess, nil
}

// RecentSessions returns up to n of userID's non-archived sessions, newest
// first. The result is a new slice, so callers can't modify the store's memory.
func (s *Store) RecentSessions(userID string, n int) []Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := s.sessions[userID]
	out := make([]Session, 0, min(n, len(list)))
	for i := len(list) - 1; i >= 0 && len(out) < n; i-- {
		if !list[i].Archived {
			out = append(out, list[i])
		}
	}
	return out
}

// Exercises returns the whole catalog sorted by name, archived included:
// history still needs archived names; the UI hides them from pickers.
func (s *Store) Exercises() []Exercise {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return sortedByName(s.exercises, func(e Exercise) string { return e.Name })
}

// Exercise returns the exercise with id, and whether it exists.
func (s *Store) Exercise(id string) (Exercise, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ex, ok := s.exercises[id]
	return ex, ok
}

// SaveExercise validates ex and writes it to disk, then to memory.
// Archiving is a save with Archived set; nothing is ever deleted.
func (s *Store) SaveExercise(ex Exercise) error {
	if err := ex.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := jsonfile.Write(filepath.Join(s.dir, "exercises", ex.ID+".json"), ex); err != nil {
		return err
	}
	s.exercises[ex.ID] = ex
	return nil
}

// Routines returns all routines sorted by name, archived included.
func (s *Store) Routines() []Routine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return sortedByName(s.routines, func(r Routine) string { return r.Name })
}

// SaveRoutine validates r, checks its exercises exist, and writes it to
// disk, then to memory. Archiving is a save with Archived set.
func (s *Store) SaveRoutine(r Routine) error {
	if err := r.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkRoutineRefs(r); err != nil {
		return err
	}
	if err := jsonfile.Write(filepath.Join(s.dir, "routines", r.ID+".json"), r); err != nil {
		return err
	}
	s.routines[r.ID] = r
	return nil
}

// checkSessionRefs checks every entry against the catalog: the exercise
// must exist and each set must fit the metrics it tracks. Callers hold
// s.mu (or own s exclusively, as Open does).
func (s *Store) checkSessionRefs(sess Session) error {
	for _, e := range sess.Entries {
		ex, ok := s.exercises[e.ExerciseID]
		if !ok {
			return fmt.Errorf("%w session %s: unknown exercise %q", ErrInvalid, sess.ID, e.ExerciseID)
		}
		for _, set := range e.Sets {
			if err := set.Validate(ex); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkRoutineRefs checks that every exercise r lists exists in the
// catalog. Same locking rule as checkSessionRefs.
func (s *Store) checkRoutineRefs(r Routine) error {
	for _, re := range r.Exercises {
		if _, ok := s.exercises[re.ExerciseID]; !ok {
			return fmt.Errorf("%w routine %s: unknown exercise %q", ErrInvalid, r.ID, re.ExerciseID)
		}
	}
	return nil
}

// sortedByName returns m's values sorted by name(v). Maps have no order,
// so list endpoints sort to give the UI a stable result. O(n log n) per
// call; fine for a catalog of hundreds.
func sortedByName[T any](m map[string]T, name func(T) string) []T {
	out := make([]T, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	slices.SortFunc(out, func(a, b T) int { return strings.Compare(name(a), name(b)) })
	return out
}

// validator is any model that can check itself. A Go constraint can
// require methods but not fields, which is why loadDir also takes id.
type validator interface {
	Validate() error
}

// loadDir decodes every *.json file in dir, in filename order (os.ReadDir
// sorts by name, so timestamp-named sessions come back oldest first).
// Dotfiles are skipped: they are temp files from jsonfile.Write.
// A missing dir means no data yet, not an error.
//
// Each item must pass its own Validate (the files are hand-editable, so
// never trust them more than an API request). Catalog references are the
// store's job, checked in Open once everything is loaded.
//
// Each item's id(v) must equal its filename (minus .json): saves write to
// <id>.json, so a mismatch (e.g. a hand-copied file) would make the next
// save overwrite a different file. Fail loudly rather than lose data.
func loadDir[T validator](dir string, id func(T) string) ([]T, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items := make([]T, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasPrefix(name, ".") || filepath.Ext(name) != ".json" {
			continue
		}
		path := filepath.Join(dir, name)
		v, err := jsonfile.Read[T](path)
		if err != nil {
			return nil, err
		}
		if want := strings.TrimSuffix(name, ".json"); id(v) != want {
			return nil, fmt.Errorf("load %s: id %q does not match filename", path, id(v))
		}
		if err := v.Validate(); err != nil {
			return nil, fmt.Errorf("load %s: %w", path, err)
		}
		items = append(items, v)
	}
	return items, nil
}
