package fitness

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// ErrInvalid marks data that breaks a model rule. Callers check it with
// errors.Is (e.g. to answer HTTP 400 instead of 500).
var ErrInvalid = errors.New("invalid")

// ErrConflict marks a create that would overwrite existing data
// (HTTP 409).
var ErrConflict = errors.New("conflict")

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
	if !jsonfile.ValidID(e.ID) {
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

// maxRestSec caps a plan's rest hint; anything longer is a typo
// (9000 for 90), not a rest.
const maxRestSec = 3600

// Validate checks r's own rules:
//   - ID, Name and CreatedBy are required
//   - at least one exercise, each with an ExerciseID
//   - SuggestedSets is not negative (0 means no suggestion)
//   - RestSec is 0 (default rest) to maxRestSec
//
// Whether each ExerciseID exists in the catalog is checked by the store,
// which has the catalog; Validate only sees the plan itself.
func (r Plan) Validate() error {
	if !jsonfile.ValidID(r.ID) {
		return fmt.Errorf("%w plan: bad ID %q", ErrInvalid, r.ID)
	}
	if r.Name == "" {
		return fmt.Errorf("%w plan %s: missing name", ErrInvalid, r.ID)
	}
	if r.CreatedBy == "" {
		return fmt.Errorf("%w plan %s: missing created_by", ErrInvalid, r.ID)
	}
	if len(r.Exercises) == 0 {
		return fmt.Errorf("%w plan %s: no exercises", ErrInvalid, r.ID)
	}
	for i, ex := range r.Exercises {
		if ex.ExerciseID == "" {
			return fmt.Errorf("%w plan %s: exercise %d missing exercise_id", ErrInvalid, r.ID, i)
		}
		if ex.SuggestedSets < 0 {
			return fmt.Errorf("%w plan %s: exercise %d has negative suggested_sets", ErrInvalid, r.ID, i)
		}
		if ex.RestSec < 0 || ex.RestSec > maxRestSec {
			return fmt.Errorf("%w plan %s: exercise %d rest_sec must be 0 to %d", ErrInvalid, r.ID, i, maxRestSec)
		}
	}
	return nil
}

// Validate checks s's own rules:
//   - ID and UserID are valid IDs (they become paths on disk)
//   - StartedAt is set; EndedAt, if set, is not before it
//   - every entry has an ExerciseID
//
// Checking entries against the exercise catalog (and each Set against its
// exercise) is done by the store, which has the catalog.
func (s Session) Validate() error {
	if !jsonfile.ValidID(s.ID) {
		return fmt.Errorf("%w session: bad ID %q", ErrInvalid, s.ID)
	}
	if !jsonfile.ValidID(s.UserID) {
		return fmt.Errorf("%w session %s: bad user_id %q", ErrInvalid, s.ID, s.UserID)
	}
	if s.StartedAt.IsZero() {
		return fmt.Errorf("%w session %s: missing started_at", ErrInvalid, s.ID)
	}
	if !s.EndedAt.IsZero() && s.EndedAt.Before(s.StartedAt) {
		return fmt.Errorf("%w session %s: ended before it started", ErrInvalid, s.ID)
	}
	for i, e := range s.Entries {
		if e.ExerciseID == "" {
			return fmt.Errorf("%w session %s: entry %d missing exercise_id", ErrInvalid, s.ID, i)
		}
	}
	return nil
}

// Validate checks p's own rules:
//   - UserID is a valid ID
//   - Goal, if set, is a known value
//   - HeightM is 0 (not set) or a plausible human height (0.5 to 2.75 m)
//   - Schedule keys are weekdays and values are valid plan IDs
//   - WeightPromptSkipped, if set, is an ISO week like "2026-W41"
//
// Whether scheduled plans exist is checked by the store.
func (p Profile) Validate() error {
	if !jsonfile.ValidID(p.UserID) {
		return fmt.Errorf("%w fitness profile: bad user_id %q", ErrInvalid, p.UserID)
	}
	if p.Goal != "" && !slices.Contains(AllGoals, p.Goal) {
		return fmt.Errorf("%w fitness profile %s: unknown goal %q", ErrInvalid, p.UserID, p.Goal)
	}
	if p.HeightM != 0 && (p.HeightM < 0.5 || p.HeightM > 2.75) {
		return fmt.Errorf("%w fitness profile %s: height_m %g is not between 0.5 and 2.75", ErrInvalid, p.UserID, p.HeightM)
	}
	for day, planID := range p.Schedule {
		if !slices.Contains(AllWeekdays, day) {
			return fmt.Errorf("%w fitness profile %s: unknown weekday %q", ErrInvalid, p.UserID, day)
		}
		if !jsonfile.ValidID(planID) {
			return fmt.Errorf("%w fitness profile %s: %s has bad plan ID %q", ErrInvalid, p.UserID, day, planID)
		}
	}
	if p.WeightPromptSkipped != "" && !validISOWeek(p.WeightPromptSkipped) {
		return fmt.Errorf("%w fitness profile %s: weight_prompt_skipped %q is not like 2026-W41", ErrInvalid, p.UserID, p.WeightPromptSkipped)
	}
	return nil
}

// validISOWeek reports whether w looks like "2026-W41" (week 01 to 53).
func validISOWeek(w string) bool {
	var year, week int
	n, err := fmt.Sscanf(w, "%4d-W%2d", &year, &week)
	return err == nil && n == 2 && len(w) == 8 && week >= 1 && week <= 53
}

// Validate checks w's rules:
//   - Date is a real "YYYY-MM-DD" date, not in the future
//   - WeightKg is plausible for a person (20 to 400 kg)
func (w WeightEntry) Validate() error {
	day, err := time.Parse(time.DateOnly, w.Date)
	if err != nil {
		return fmt.Errorf("%w weight: date %q is not YYYY-MM-DD", ErrInvalid, w.Date)
	}
	// Compare dates, not instants: "today" is still allowed late in the
	// day in any time zone, so allow one day of slack.
	if day.After(time.Now().AddDate(0, 0, 1)) {
		return fmt.Errorf("%w weight: date %s is in the future", ErrInvalid, w.Date)
	}
	if w.WeightKg < 20 || w.WeightKg > 400 {
		return fmt.Errorf("%w weight %s: %g kg is not between 20 and 400", ErrInvalid, w.Date, w.WeightKg)
	}
	return nil
}
