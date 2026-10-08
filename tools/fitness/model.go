package fitness

// Needed for time.Time in Session and Entry.
import "time"

// ---- Catalog ----

// Exercise is a shared catalog entry. Sessions reference it by ID, so
// exercises are archived rather than deleted to keep history resolvable.
type Exercise struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Activation is a sparse map of how hard each muscle works, in (0, 1].
	// Muscles not listed read as 0 (Go's map zero value).
	Activation map[Muscle]float64 `json:"activation"`

	// Bodyweight is true if the exercise doesn't require external weight, e.g. pushups.
	Bodyweight bool `json:"bodyweight"`

	// Metrics lists which Set fields this exercise records.
	Metrics []Metric `json:"metrics"`

	Archived bool `json:"archived,omitempty"`
}

// Metric names a measurement an exercise tracks; each maps to a Set field.
type Metric string

const (
	MetricReps     Metric = "reps"
	MetricWeight   Metric = "weight"
	MetricDistance Metric = "distance"
	MetricDuration Metric = "duration"
)

// Muscle values double as element IDs in the body diagram SVG.
type Muscle string

const (
	MuscleChest      Muscle = "chest"
	MuscleShoulders  Muscle = "shoulders"
	MuscleBiceps     Muscle = "biceps"
	MuscleTriceps    Muscle = "triceps"
	MuscleForearms   Muscle = "forearms"
	MuscleAbs        Muscle = "abs"
	MuscleObliques   Muscle = "obliques"
	MuscleLats       Muscle = "lats"
	MuscleTraps      Muscle = "traps"
	MuscleLowerBack  Muscle = "lower_back"
	MuscleGlutes     Muscle = "glutes"
	MuscleQuads      Muscle = "quads"
	MuscleHamstrings Muscle = "hamstrings"
	MuscleCalves     Muscle = "calves"
)

// ---- Planning ----

// Routine is a shared, reusable workout template, stored as
// routines/<id>.json. It suggests what to do; it never constrains a session.
// Starting a session from a routine pre-fills the session's entries.
type Routine struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CreatedBy string            `json:"created_by"` // user ID
	Exercises []RoutineExercise `json:"exercises"`  // in suggested order
}

// RoutineExercise is one step of a routine. SuggestedSets is a hint shown
// in the UI (0 = no suggestion); sessions may log any number of sets.
type RoutineExercise struct {
	ExerciseID    string `json:"exercise_id"`
	SuggestedSets int    `json:"suggested_sets,omitempty"`
}

// ---- Logging ----

// Session is one workout, stored as users/<user_id>/sessions/<id>.json.
// ID is assigned once from the start time and never recomputed.
// A zero EndedAt means the session is still in progress.
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	RoutineID    string    `json:"routine_id,omitempty"`
	StartedAt    time.Time `json:"started_at"`
	EndedAt      time.Time `json:"ended_at,omitzero"`
	BodyWeightKg float64   `json:"body_weight_kg,omitempty"`
	Entries      []Entry   `json:"entries"`
}

// Entry is one exercise performed within a session. Sets may be empty
// while the exercise has been picked but nothing logged yet.
type Entry struct {
	ExerciseID string `json:"exercise_id"`
	Sets       []Set  `json:"sets"`
}

// Set is a single recorded instance of an exercise, e.g. 10 reps at 100kg.
// The exercise's Metrics decide which fields apply; unused fields stay zero
// and are omitted from JSON.
type Set struct {
	Reps        int     `json:"reps,omitempty"`
	WeightKg    float64 `json:"weight_kg,omitempty"`
	DurationSec int     `json:"duration_sec,omitempty"`
	DistanceM   float64 `json:"distance_m,omitempty"`
}
