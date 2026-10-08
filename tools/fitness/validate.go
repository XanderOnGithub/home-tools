package fitness

import (
	"errors"
	"fmt"
	"slices"
)

// ErrInvalid marks data that breaks a model rule. Callers check it with
// errors.Is (e.g. to answer HTTP 400 instead of 500).
var ErrInvalid = errors.New("invalid")

// validID reports whether id is safe to use as a file or folder name:
// non-empty, only letters, digits, '-' and '_'. This blocks path tricks
// like "../" since IDs become paths on disk.
func validID(id string) bool {
	if id == "" {
		return false
	}
	for _, c := range id {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_'
		if !ok {
			return false
		}
	}
	return true
}

// Validate checks s against the metrics ex tracks:
//   - no value may be negative
//   - a tracked metric must be > 0, except weight on bodyweight exercises
//     (0 means bodyweight only)
//   - an untracked metric must be 0
func (s Set) Validate(ex Exercise) error {
	if s.Reps < 0 || s.WeightKg < 0 || s.DurationSec < 0 || s.DistanceM < 0 {
		return fmt.Errorf("%w set: negative value", ErrInvalid)
	}

	fields := [...]struct {
		metric Metric
		filled bool
	}{
		{MetricReps, s.Reps > 0},
		{MetricWeight, s.WeightKg > 0},
		{MetricDuration, s.DurationSec > 0},
		{MetricDistance, s.DistanceM > 0},
	}

	for _, f := range fields {
		// Linear scan: an exercise tracks at most 4 metrics, so this beats
		// building a map (no allocation, no hashing).
		tracked := slices.Contains(ex.Metrics, f.metric)
		optional := f.metric == MetricWeight && ex.Bodyweight

		switch {
		case f.filled && !tracked:
			return fmt.Errorf("%w set: %s not tracked by %s", ErrInvalid, f.metric, ex.ID)
		case !f.filled && tracked && !optional:
			return fmt.Errorf("%w set: %s required by %s", ErrInvalid, f.metric, ex.ID)
		}
	}
	return nil
}

// Validate checks e's own rules:
//   - ID and Name are required
//   - at least one metric is tracked
//   - at least one muscle activation, each in (0, 1]
//   - metrics, muscles, equipment, category and level are known values
//     (category and level may be empty)
func (e Exercise) Validate() error {
	if !validID(e.ID) {
		return fmt.Errorf("%w exercise: bad ID %q", ErrInvalid, e.ID)
	}
	if e.Name == "" {
		return fmt.Errorf("%w exercise %s: missing name", ErrInvalid, e.ID)
	}
	if len(e.Metrics) == 0 {
		return fmt.Errorf("%w exercise %s: missing metrics", ErrInvalid, e.ID)
	}
	for _, m := range e.Metrics {
		if !slices.Contains(AllMetrics, m) {
			return fmt.Errorf("%w exercise %s: unknown metric %q", ErrInvalid, e.ID, m)
		}
	}
	for _, eq := range e.Equipment {
		if !slices.Contains(AllEquipment, eq) {
			return fmt.Errorf("%w exercise %s: unknown equipment %q", ErrInvalid, e.ID, eq)
		}
	}
	if e.Category != "" && !slices.Contains(AllCategories, e.Category) {
		return fmt.Errorf("%w exercise %s: unknown category %q", ErrInvalid, e.ID, e.Category)
	}
	if e.Level != "" && !slices.Contains(AllLevels, e.Level) {
		return fmt.Errorf("%w exercise %s: unknown level %q", ErrInvalid, e.ID, e.Level)
	}
	if len(e.Activation) == 0 {
		return fmt.Errorf("%w exercise %s: missing muscle activation", ErrInvalid, e.ID)
	}
	for m, v := range e.Activation {
		if !slices.Contains(AllMuscles, m) {
			return fmt.Errorf("%w exercise %s: unknown muscle %q", ErrInvalid, e.ID, m)
		}
		if v <= 0 || v > 1 {
			return fmt.Errorf("%w exercise %s: activation %s=%g not in (0, 1]", ErrInvalid, e.ID, m, v)
		}
	}
	return nil
}

// Validate checks r's own rules:
//   - ID, Name and CreatedBy are required
//   - at least one exercise, each with an ExerciseID
//   - SuggestedSets is not negative (0 means no suggestion)
//
// Whether each ExerciseID exists in the catalog is checked by the store,
// which has the catalog; Validate only sees the routine itself.
func (r Routine) Validate() error {
	if !validID(r.ID) {
		return fmt.Errorf("%w routine: bad ID %q", ErrInvalid, r.ID)
	}
	if r.Name == "" {
		return fmt.Errorf("%w routine %s: missing name", ErrInvalid, r.ID)
	}
	if r.CreatedBy == "" {
		return fmt.Errorf("%w routine %s: missing created_by", ErrInvalid, r.ID)
	}
	if len(r.Exercises) == 0 {
		return fmt.Errorf("%w routine %s: no exercises", ErrInvalid, r.ID)
	}
	for i, ex := range r.Exercises {
		if ex.ExerciseID == "" {
			return fmt.Errorf("%w routine %s: exercise %d missing exercise_id", ErrInvalid, r.ID, i)
		}
		if ex.SuggestedSets < 0 {
			return fmt.Errorf("%w routine %s: exercise %d has negative suggested_sets", ErrInvalid, r.ID, i)
		}
	}
	return nil
}

// Validate checks s's own rules:
//   - ID and UserID are valid IDs (they become paths on disk)
//   - StartedAt is set; EndedAt, if set, is not before it
//   - BodyWeightKg is not negative
//   - every entry has an ExerciseID
//
// Checking entries against the exercise catalog (and each Set against its
// exercise) is done by the store, which has the catalog.
func (s Session) Validate() error {
	if !validID(s.ID) {
		return fmt.Errorf("%w session: bad ID %q", ErrInvalid, s.ID)
	}
	if !validID(s.UserID) {
		return fmt.Errorf("%w session %s: bad user_id %q", ErrInvalid, s.ID, s.UserID)
	}
	if s.StartedAt.IsZero() {
		return fmt.Errorf("%w session %s: missing started_at", ErrInvalid, s.ID)
	}
	if !s.EndedAt.IsZero() && s.EndedAt.Before(s.StartedAt) {
		return fmt.Errorf("%w session %s: ended before it started", ErrInvalid, s.ID)
	}
	if s.BodyWeightKg < 0 {
		return fmt.Errorf("%w session %s: negative body_weight_kg", ErrInvalid, s.ID)
	}
	for i, e := range s.Entries {
		if e.ExerciseID == "" {
			return fmt.Errorf("%w session %s: entry %d missing exercise_id", ErrInvalid, s.ID, i)
		}
	}
	return nil
}
