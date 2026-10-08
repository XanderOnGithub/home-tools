package users

import (
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// Store holds all profiles in memory and writes changes through to JSON
// files under dir. Safe for concurrent use; always use it via *Store.
//
// Tools call User(id) while holding their own store's lock. That's safe:
// this store never calls back into a tool, so locks are always taken in
// the same order (tool → users) and can't deadlock.
type Store struct {
	dir string

	mu    sync.RWMutex
	users map[string]User // by ID
}

// Open loads every profile under dir. A missing dir means no profiles
// yet; any invalid file fails the whole open.
func Open(dir string) (*Store, error) {
	list, err := jsonfile.LoadDir(dir, func(u User) string { return u.ID })
	if err != nil {
		return nil, err
	}
	s := &Store{dir: dir, users: make(map[string]User, len(list))}
	for _, u := range list {
		s.users[u.ID] = u
	}
	return s, nil
}

// Users returns every profile sorted by name, archived included (history
// still needs their names); the picker hides archived ones.
func (s *Store) Users() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	slices.SortFunc(out, func(a, b User) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// User returns the profile with id, and whether it exists.
func (s *Store) User(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

// Save validates u and writes it to <dir>/<id>.json, then to memory.
// Creating a profile is just its first save; archiving is a save with
// Archived set. Nothing is deleted.
func (s *Store) Save(u User) error {
	if err := u.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := jsonfile.Write(filepath.Join(s.dir, u.ID+".json"), u); err != nil {
		return err
	}
	s.users[u.ID] = u
	return nil
}
