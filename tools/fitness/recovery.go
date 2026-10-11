package fitness

import (
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

// Muscle recovery (decision #49): an estimate of how worked each muscle
// is right now, from the user's recent workouts. Nothing is stored; it's
// a pure function of the sessions, the catalog and the time.
//
// The model, kept deliberately simple (it's a guide, not a measurement;
// the research behind it: docs/research/muscle-recovery.md):
//   - Every logged set adds fatigue to each muscle it works, as much as the
//     exercise's activation (1 for a primary muscle, 0.5 for a secondary
//     one), at the time the workout started. Stretches add none. We don't
//     log effort, so every set counts as a hard one (near failure).
//   - Sets of an exercise not done in the 14 days before count 1.5×
//     (noveltyFactor): new or returning movements do more damage, and the
//     protection of having done it fades over 1–2 weeks (repeated bout
//     effect). Gentler than the research's 5-day floor for a first
//     exposure, which would tell every beginner to wait 5 days.
//   - Fatigue halves every recoveryHalfLife × the muscle's multiplier
//     (muscleSpeed: legs a little faster, arms and hamstrings a little
//     slower; design defaults, size matters little). With the thresholds
//     below, 3 hard sets are ready again after about 1.5 days, 6 after
//     2.5 and 10 after 3.3: the research's 48–96 h for hard to very hard
//     sessions.
//   - Rest from fatigue ≥ restFatigue, Recovering from ≥ readyFatigue,
//     Ready below that.
//
// O(all sets of the user) per call: novelty needs the whole history.

const (
	recoveryHalfLife = 24 * time.Hour
	recoveryWindow   = 7 * 24 * time.Hour // older sets have decayed to < 1% of a set
	restFatigue      = 3.0                // about 3 hard sets, recently
	readyFatigue     = 1.0                // less than one hard set's worth left
	noveltyFactor    = 1.5                // sets of a new (or returning) exercise
	noveltyGap       = 14 * 24 * time.Hour
)

// muscleSpeed scales recoveryHalfLife per muscle (missing = 1): the
// research doc's defaults (§3): quads, calves and glutes recover a little
// faster, hamstrings and arms a little slower. Kept small on purpose.
var muscleSpeed = map[Muscle]float64{
	MuscleQuadriceps: 0.9,
	MuscleCalves:     0.9,
	MuscleGlutes:     0.9,
	MuscleHamstrings: 1.1,
	MuscleBiceps:     1.1,
	MuscleTriceps:    1.1,
}

// halfLife is how fast m's fatigue fades.
func halfLife(m Muscle) time.Duration {
	if f, ok := muscleSpeed[m]; ok {
		return time.Duration(float64(recoveryHalfLife) * f)
	}
	return recoveryHalfLife
}

// novelExercises marks, per session index, the exercises that hadn't been
// done (with logged sets) in the noveltyGap before it. Discarded workouts
// don't count. sessions are oldest → newest; O(all sets).
func novelExercises(sessions []Session) []map[string]bool {
	last := map[string]time.Time{} // exercise ID → latest start with sets
	out := make([]map[string]bool, len(sessions))
	for i, sess := range sessions {
		if sess.Archived {
			continue
		}
		novel := map[string]bool{}
		for _, e := range sess.Entries {
			if len(e.Sets) == 0 {
				continue
			}
			if at, ok := last[e.ExerciseID]; !ok || sess.StartedAt.Sub(at) > noveltyGap {
				novel[e.ExerciseID] = true
			}
		}
		for _, e := range sess.Entries { // after the check: an exercise twice in one session is novel twice
			if len(e.Sets) > 0 {
				last[e.ExerciseID] = sess.StartedAt
			}
		}
		out[i] = novel
	}
	return out
}

// RecoveryState is how ready a muscle is to be trained again.
type RecoveryState string

const (
	StateReady      RecoveryState = "ready"
	StateRecovering RecoveryState = "recovering" // fine to train lightly
	StateRest       RecoveryState = "rest"       // worked hard recently
)

// MuscleRecovery is one muscle's estimate.
type MuscleRecovery struct {
	Muscle Muscle        `json:"muscle"`
	State  RecoveryState `json:"state"`
	// ReadyAt is when it's ready again; zero (omitted) when it already is.
	ReadyAt time.Time `json:"ready_at,omitzero"`
	// LastWorked is the start of the latest workout that worked it in the
	// window; zero (omitted) if none did.
	LastWorked time.Time `json:"last_worked,omitzero"`
	// Sets7d is this week's volume: sets × activation over the last 7
	// days (a primary set counts 1, a secondary one 0.5).
	Sets7d float64 `json:"sets_7d"`
	// Exercises worked it in the window, most recent first (IDs).
	Exercises []string `json:"exercises"`
}

// Recovery returns every muscle's estimate for userID at now, in
// AllMuscles order. In-progress workouts count: sets logged a minute ago
// already tire the muscle.
func (s *Store) Recovery(userID string, now time.Time) []MuscleRecovery {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return recovery(s.sessions[userID], func(id string) (Exercise, bool) {
		ex, ok := s.exercises[id]
		return ex, ok
	}, now)
}

// recovery is the model above. sessions are oldest → newest.
func recovery(sessions []Session, exercise func(id string) (Exercise, bool), now time.Time) []MuscleRecovery {
	type acc struct {
		fatigue    float64
		sets       float64
		lastWorked time.Time
		exercises  []string // newest first after the loop's reversal
	}
	by := make(map[Muscle]*acc, len(AllMuscles))
	for _, m := range AllMuscles {
		by[m] = &acc{}
	}

	novel := novelExercises(sessions)
	for i := len(sessions) - 1; i >= 0; i-- { // newest first: stop at the window's edge
		sess := sessions[i]
		age := now.Sub(sess.StartedAt)
		if age > recoveryWindow {
			break
		}
		if sess.Archived || age < 0 { // discarded, or a clock in the future
			continue
		}
		for _, e := range sess.Entries {
			ex, ok := exercise(e.ExerciseID)
			if !ok || len(e.Sets) == 0 || ex.Category == CategoryStretching {
				continue
			}
			n := float64(len(e.Sets))
			load := n
			if novel[i][ex.ID] {
				load *= noveltyFactor
			}
			for m, a := range ex.Activation {
				st := by[m]
				if st == nil {
					continue // not a muscle we know (validation prevents this)
				}
				st.fatigue += load * a * math.Exp2(-float64(age)/float64(halfLife(m)))
				st.sets += n * a
				if st.lastWorked.IsZero() {
					st.lastWorked = sess.StartedAt // newest first: the first one is the latest
				}
				if !slices.Contains(st.exercises, ex.ID) {
					st.exercises = append(st.exercises, ex.ID)
				}
			}
		}
	}

	out := make([]MuscleRecovery, 0, len(AllMuscles))
	for _, m := range AllMuscles {
		st := by[m]
		r := MuscleRecovery{Muscle: m, State: StateReady, LastWorked: st.lastWorked, Sets7d: math.Round(st.sets*10) / 10, Exercises: st.exercises}
		if r.Exercises == nil {
			r.Exercises = []string{}
		}
		if st.fatigue >= readyFatigue {
			r.State = StateRecovering
			if st.fatigue >= restFatigue {
				r.State = StateRest
			}
			// fatigue · 2^(-t/h) = readyFatigue  ⇒  t = h · log2(fatigue / readyFatigue)
			wait := time.Duration(float64(halfLife(m)) * math.Log2(st.fatigue/readyFatigue))
			r.ReadyAt = now.Add(wait).Truncate(time.Minute)
		}
		out = append(out, r)
	}
	return out
}

// getRecovery: GET /api/users/{user}/recovery.
func (h *handlers) getRecovery(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	if _, ok := h.store.users.User(user); !ok {
		httpx.WriteError(w, http.StatusNotFound, "user "+user+" not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.store.Recovery(user, time.Now()))
}
