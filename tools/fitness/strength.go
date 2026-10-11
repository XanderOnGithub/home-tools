package fitness

import (
	"cmp"
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

// Strength trends (decision #50): are you getting stronger, per exercise
// and per muscle, compared with yourself. Nothing is stored; it's derived
// from finished workouts.
//
//   - A weighted exercise's measure is its estimated one-rep max (Epley:
//     weight × (1 + reps/30)), from sets of 1–12 reps; above that the
//     formula is unreliable. A bodyweight exercise's measure is the most
//     reps in a set without added weight. Cardio, stretches and timed
//     holds have no strength measure.
//   - An exercise's change compares its best in the last 4 weeks with its
//     best in the 4 weeks before. Without those (someone new), it compares
//     with the first workout that had it, if that was at least a week
//     before the latest: "since you started".
//   - A muscle's change is the average change of the exercises that work
//     it as a main mover (activation 1), 4-week ones preferred.
//
// O(sets of the user) per call.

const (
	strengthWindow = 28 * 24 * time.Hour
	maxRepsForMax  = 12 // Epley drifts above this
	minSinceStart  = 7 * 24 * time.Hour
)

// Basis says what a change compares with.
type Basis string

const (
	BasisFourWeeks  Basis = "4_weeks"     // the last 4 weeks vs the 4 before
	BasisSinceStart Basis = "since_start" // the last 4 weeks vs the first time
)

// Measure says what an exercise's numbers are.
type Measure string

const (
	MeasureMaxKg Measure = "max_kg" // estimated one-rep max, kg
	MeasureReps  Measure = "reps"   // most reps in a set (bodyweight)
)

// ExerciseStrength is one exercise's numbers.
type ExerciseStrength struct {
	ExerciseID string  `json:"exercise_id"`
	Measure    Measure `json:"measure"`
	Best       float64 `json:"best"`   // all time
	Recent     float64 `json:"recent"` // best in the last 4 weeks; 0 = not done then
	Change     float64 `json:"change"` // 0.08 = 8% better; only with a Basis
	Basis      Basis   `json:"basis,omitempty"`
}

// MuscleStrength is one muscle's trend; muscles without one are left out.
type MuscleStrength struct {
	Muscle    Muscle   `json:"muscle"`
	Change    float64  `json:"change"`
	Basis     Basis    `json:"basis"`
	Exercises []string `json:"exercises"` // the exercises it's based on
}

// Strength is the answer of GET /api/users/{user}/strength.
type Strength struct {
	Muscles   []MuscleStrength   `json:"muscles"`   // AllMuscles order
	Exercises []ExerciseStrength `json:"exercises"` // every exercise with a measure, by ID
}

// Strength returns userID's trends at now.
func (s *Store) Strength(userID string, now time.Time) Strength {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return strength(s.sessions[userID], func(id string) (Exercise, bool) {
		ex, ok := s.exercises[id]
		return ex, ok
	}, now)
}

// estimatedMax is a set's estimated one-rep max in kg (Epley), or 0 when
// the set can't give one (no weight, no reps, or more than 12 reps).
func estimatedMax(s Set) float64 {
	switch {
	case s.WeightKg <= 0 || s.Reps <= 0 || s.Reps > maxRepsForMax:
		return 0
	case s.Reps == 1:
		return s.WeightKg // a single is the max itself (Epley would add 3%)
	}
	return s.WeightKg * (1 + float64(s.Reps)/30)
}

// measureOf says how ex is measured, or "" if it has no strength measure.
func measureOf(ex Exercise) Measure {
	if ex.Category == CategoryCardio || ex.Category == CategoryStretching || !slices.Contains(ex.Metrics, MetricReps) {
		return ""
	}
	if ex.Bodyweight || !slices.Contains(ex.Metrics, MetricWeight) {
		return MeasureReps
	}
	return MeasureMaxKg
}

// score is a set's value under measure (0 = it doesn't count).
func score(m Measure, s Set) float64 {
	if m == MeasureReps {
		if s.WeightKg > 0 {
			return 0 // added weight: not comparable with bodyweight reps
		}
		return float64(s.Reps)
	}
	return estimatedMax(s)
}

// strength is the model above. sessions are oldest → newest.
func strength(sessions []Session, exercise func(id string) (Exercise, bool), now time.Time) Strength {
	type point struct {
		at    time.Time
		value float64
	}
	byExercise := map[string][]point{} // one point per workout: its best set
	measures := map[string]Measure{}
	for _, sess := range sessions {
		if !counts(sess) || sess.StartedAt.After(now) {
			continue
		}
		bests := map[string]float64{} // an exercise twice in one workout: one point
		for _, e := range sess.Entries {
			ex, ok := exercise(e.ExerciseID)
			if !ok {
				continue
			}
			m := measureOf(ex)
			if m == "" {
				continue
			}
			measures[ex.ID] = m
			for _, set := range e.Sets {
				bests[ex.ID] = max(bests[ex.ID], score(m, set))
			}
		}
		for id, v := range bests {
			if v > 0 {
				byExercise[id] = append(byExercise[id], point{sess.StartedAt, v})
			}
		}
	}

	recentFrom, previousFrom := now.Add(-strengthWindow), now.Add(-2*strengthWindow)
	out := Strength{Muscles: []MuscleStrength{}, Exercises: []ExerciseStrength{}}
	for id, pts := range byExercise { // pts oldest → newest
		es := ExerciseStrength{ExerciseID: id, Measure: measures[id]}
		var previous float64
		for _, p := range pts {
			es.Best = max(es.Best, p.value)
			switch {
			case !p.at.Before(recentFrom):
				es.Recent = max(es.Recent, p.value)
			case !p.at.Before(previousFrom):
				previous = max(previous, p.value)
			}
		}
		first, latest := pts[0], pts[len(pts)-1]
		switch {
		case es.Recent > 0 && previous > 0:
			es.Change, es.Basis = es.Recent/previous-1, BasisFourWeeks
		case es.Recent > 0 && latest.at.Sub(first.at) >= minSinceStart:
			es.Change, es.Basis = es.Recent/first.value-1, BasisSinceStart
		}
		es.Change = math.Round(es.Change*1000) / 1000
		out.Exercises = append(out.Exercises, es)
	}
	slices.SortFunc(out.Exercises, func(a, b ExerciseStrength) int { return cmp.Compare(a.ExerciseID, b.ExerciseID) })

	for _, m := range AllMuscles {
		var four, since []ExerciseStrength
		for _, es := range out.Exercises { // sorted: the Exercises list comes out stable
			ex, _ := exercise(es.ExerciseID)
			if ex.Activation[m] < 1 {
				continue
			}
			switch es.Basis {
			case BasisFourWeeks:
				four = append(four, es)
			case BasisSinceStart:
				since = append(since, es)
			}
		}
		used, basis := four, BasisFourWeeks
		if len(used) == 0 {
			used, basis = since, BasisSinceStart
		}
		if len(used) == 0 {
			continue
		}
		ms := MuscleStrength{Muscle: m, Basis: basis, Exercises: make([]string, 0, len(used))}
		for _, es := range used {
			ms.Change += es.Change / float64(len(used))
			ms.Exercises = append(ms.Exercises, es.ExerciseID)
		}
		ms.Change = math.Round(ms.Change*1000) / 1000
		out.Muscles = append(out.Muscles, ms)
	}
	return out
}

// getStrength: GET /api/users/{user}/strength.
func (h *handlers) getStrength(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	if _, ok := h.store.users.User(user); !ok {
		httpx.WriteError(w, http.StatusNotFound, "user "+user+" not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.store.Strength(user, time.Now()))
}
