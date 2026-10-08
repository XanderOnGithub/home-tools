package fitness

import (
	"errors"
	"fmt"
	"slices"
)

// ErrInvalid marks data that breaks a model rule. Callers check it with
// errors.Is (e.g. to answer HTTP 400 instead of 500).
var ErrInvalid = errors.New("invalid")

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
func (e Exercise) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("%w exercise: missing ID", ErrInvalid)
	}
	if e.Name == "" {
		return fmt.Errorf("%w exercise %s: missing name", ErrInvalid, e.ID)
	}
	if len(e.Metrics) == 0 {
		return fmt.Errorf("%w exercise %s: missing metrics", ErrInvalid, e.ID)
	}
	if len(e.Activation) == 0 {
		return fmt.Errorf("%w exercise %s: missing muscle activation", ErrInvalid, e.ID)
	}
	for m, v := range e.Activation {
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
	if r.ID == "" {
		return fmt.Errorf("%w routine: missing ID", ErrInvalid)
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
