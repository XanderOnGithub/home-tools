package games

import (
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// Store holds the server list in memory and writes changes through to
// JSON files under dir (ADR 0002). Safe for concurrent use.
type Store struct {
	dir string

	mu      sync.RWMutex
	servers map[string]Server // by ID
}

// Open loads every server under dir/servers. A missing folder means none
// configured yet; any invalid file fails the whole open.
func Open(dir string) (*Store, error) {
	list, err := jsonfile.LoadDir(filepath.Join(dir, "servers"), func(s Server) string { return s.ID })
	if err != nil {
		return nil, err
	}
	s := &Store{dir: dir, servers: make(map[string]Server, len(list))}
	for _, srv := range list {
		s.servers[srv.ID] = srv
	}
	return s, nil
}

// Servers returns every server sorted by name, archived included.
func (s *Store) Servers() []Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Server, 0, len(s.servers))
	for _, srv := range s.servers {
		out = append(out, srv)
	}
	slices.SortFunc(out, func(a, b Server) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// Server returns the server with id, and whether it exists.
func (s *Store) Server(id string) (Server, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	srv, ok := s.servers[id]
	return srv, ok
}

// SaveServer validates srv and writes it to disk, then to memory.
func (s *Store) SaveServer(srv Server) error {
	if err := srv.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := jsonfile.Write(filepath.Join(s.dir, "servers", srv.ID+".json"), srv); err != nil {
		return err
	}
	s.servers[srv.ID] = srv
	return nil
}
