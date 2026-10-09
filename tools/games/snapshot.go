package games

import (
	"context"
	"slices"
	"sync"
	"time"
)

// freshFor is how long a snapshot counts as current. Older ones are still
// served instantly, but trigger a refresh in the background.
const freshFor = 4 * time.Second

// snapshot keeps the last looked-up view of every server, so the list
// answers at once instead of waiting on Docker and every game (a game that
// doesn't answer costs its whole query timeout). Stale-while-revalidate:
//
//   - fresh: served as is
//   - stale: served as is, and one refresh starts in the background, so
//     the next look (the UI polls) is current
//   - none yet (startup, or after a config change): the request waits for
//     the first refresh
//
// Nothing runs while nobody's looking: refreshes only start from requests.
type snapshot struct {
	mu      sync.Mutex
	views   []serverView // nil = none yet
	takenAt time.Time
	running chan struct{} // non-nil while a refresh runs; closed when it ends
}

// get returns the current views, refreshing per the rules above. build
// does the actual lookups.
func (s *snapshot) get(ctx context.Context, build func() []serverView) []serverView {
	s.mu.Lock()
	if s.views != nil {
		if time.Since(s.takenAt) > freshFor && s.running == nil {
			s.start(build)
		}
		out := slices.Clone(s.views)
		s.mu.Unlock()
		return out
	}
	if s.running == nil {
		s.start(build)
	}
	done := s.running
	s.mu.Unlock()

	select {
	case <-done:
	case <-ctx.Done():
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.views)
}

// start runs build in the background. Caller holds s.mu.
func (s *snapshot) start(build func() []serverView) {
	done := make(chan struct{})
	s.running = done
	go func() {
		views := build()
		s.mu.Lock()
		s.views, s.takenAt, s.running = views, time.Now(), nil
		s.mu.Unlock()
		close(done)
	}()
}

// put replaces one server's view (after an action, which already looked
// it up), keeping the rest.
func (s *snapshot) put(v serverView) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.views {
		if s.views[i].ID == v.ID {
			s.views[i] = v
		}
	}
}

// reset drops the snapshot (the server list itself changed), so the next
// request waits for a full lookup.
func (s *snapshot) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.views = nil
}

// setBusy marks one server's view as having action a running ("" = none),
// so every viewer sees it at once instead of after the next refresh.
func (s *snapshot) setBusy(id string, a Action) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.views {
		if s.views[i].ID == id {
			s.views[i].Busy = a
		}
	}
}
