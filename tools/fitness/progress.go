package fitness

import (
	"cmp"
	"net/http"
	"slices"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

// Progress reads (decision #39): all-time, per exercise, from finished,
// non-archived sessions. Small answers instead of every session ever;
// charts and records are computed in the frontend.

// LoggedExercise is one exercise a user has logged sets for.
type LoggedExercise struct {
	ExerciseID string    `json:"exercise_id"`
	LastDone   time.Time `json:"last_done"` // started_at of the latest workout with it
	Workouts   int       `json:"workouts"`  // how many workouts included it
}

// ExerciseWorkout is what one workout logged for one exercise.
type ExerciseWorkout struct {
	SessionID string    `json:"session_id"`
	StartedAt time.Time `json:"started_at"`
	Sets      []Set     `json:"sets"`
}

// counts reports whether a session belongs in progress: finished and kept.
// In-progress workouts would show half-done numbers.
func counts(sess Session) bool {
	return !sess.Archived && !sess.EndedAt.IsZero()
}

// ExerciseLog lists every exercise userID has logged sets for, most
// recently done first. O(total sets of the user).
func (s *Store) ExerciseLog(userID string) []LoggedExercise {
	s.mu.RLock()
	defer s.mu.RUnlock()

	byID := make(map[string]*LoggedExercise)
	for _, sess := range s.sessions[userID] { // oldest → newest
		if !counts(sess) {
			continue
		}
		seen := make(map[string]bool, len(sess.Entries)) // an exercise twice in one workout counts once
		for _, e := range sess.Entries {
			if len(e.Sets) == 0 || seen[e.ExerciseID] {
				continue
			}
			seen[e.ExerciseID] = true
			le := byID[e.ExerciseID]
			if le == nil {
				le = &LoggedExercise{ExerciseID: e.ExerciseID}
				byID[e.ExerciseID] = le
			}
			le.Workouts++
			le.LastDone = sess.StartedAt // sessions are in order, so the last write is the latest
		}
	}

	out := make([]LoggedExercise, 0, len(byID))
	for _, le := range byID {
		out = append(out, *le)
	}
	// Newest first; ties (same workout) by ID so the order is stable.
	slices.SortFunc(out, func(a, b LoggedExercise) int {
		return cmp.Or(b.LastDone.Compare(a.LastDone), cmp.Compare(a.ExerciseID, b.ExerciseID))
	})
	return out
}

// ExerciseHistory returns every workout in which userID logged sets for
// exerciseID, newest first. An exercise done twice in one workout comes
// back as one entry with all its sets. The sets are copies, so callers
// can't modify the store's memory.
func (s *Store) ExerciseHistory(userID, exerciseID string) []ExerciseWorkout {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := []ExerciseWorkout{} // [] not null in JSON when there's nothing
	list := s.sessions[userID]
	for i := len(list) - 1; i >= 0; i-- {
		sess := list[i]
		if !counts(sess) {
			continue
		}
		var sets []Set
		for _, e := range sess.Entries {
			if e.ExerciseID == exerciseID {
				sets = append(sets, e.Sets...) // append copies the values
			}
		}
		if len(sets) > 0 {
			out = append(out, ExerciseWorkout{SessionID: sess.ID, StartedAt: sess.StartedAt, Sets: sets})
		}
	}
	return out
}

// getExerciseLog: GET /api/users/{user}/exercise-log.
func (h *handlers) getExerciseLog(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	if _, ok := h.store.users.User(user); !ok {
		httpx.WriteError(w, http.StatusNotFound, "user "+user+" not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.store.ExerciseLog(user))
}

// getExerciseHistory: GET /api/users/{user}/exercise-log/{exercise}.
func (h *handlers) getExerciseHistory(w http.ResponseWriter, r *http.Request) {
	user, id := r.PathValue("user"), r.PathValue("exercise")
	if _, ok := h.store.users.User(user); !ok {
		httpx.WriteError(w, http.StatusNotFound, "user "+user+" not found")
		return
	}
	if _, ok := h.store.Exercise(id); !ok {
		httpx.WriteError(w, http.StatusNotFound, "exercise "+id+" not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.store.ExerciseHistory(user, id))
}
