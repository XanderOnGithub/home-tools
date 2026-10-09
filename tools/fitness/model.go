package fitness

// Needed for time.Time in Session and Entry.
import "time"

// ---- Catalog ----

// Exercise is a shared catalog entry. Sessions reference it by ID, so
// exercises are archived rather than deleted to keep history resolvable.
// (Every model is archived, never deleted: see Archived fields.)
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

	// Equipment lists everything needed; empty means none. The UI filters
	// on it ("what can I do with what I have").
	Equipment []Equipment `json:"equipment,omitempty"`

	Category     Category `json:"category,omitempty"`
	Level        Level    `json:"level,omitempty"`
	Instructions []string `json:"instructions,omitempty"` // ordered steps
	// Images are paths relative to the fitness images folder, e.g.
	// "Air_Bike/0.jpg" (start position) and ".../1.jpg" (end position).
	Images []string `json:"images,omitempty"`

	// Source records where an imported exercise came from (attribution);
	// empty for exercises made by hand.
	Source string `json:"source,omitempty"`

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

// AllMetrics lists every valid Metric (for validation and UI options).
var AllMetrics = []Metric{MetricReps, MetricWeight, MetricDistance, MetricDuration}

// Muscle values match free-exercise-db's muscle names (spaces → '_') so
// imports are lossless. The body diagram uses different region names; the
// frontend maps Muscle → diagram regions.
type Muscle string

const (
	MuscleAbdominals Muscle = "abdominals"
	MuscleAbductors  Muscle = "abductors"
	MuscleAdductors  Muscle = "adductors"
	MuscleBiceps     Muscle = "biceps"
	MuscleCalves     Muscle = "calves"
	MuscleChest      Muscle = "chest"
	MuscleForearms   Muscle = "forearms"
	MuscleGlutes     Muscle = "glutes"
	MuscleHamstrings Muscle = "hamstrings"
	MuscleLats       Muscle = "lats"
	MuscleLowerBack  Muscle = "lower_back"
	MuscleMiddleBack Muscle = "middle_back"
	MuscleNeck       Muscle = "neck"
	MuscleQuadriceps Muscle = "quadriceps"
	MuscleShoulders  Muscle = "shoulders"
	MuscleTraps      Muscle = "traps"
	MuscleTriceps    Muscle = "triceps"
)

// AllMuscles lists every valid Muscle (for validation and UI options).
var AllMuscles = []Muscle{
	MuscleAbdominals, MuscleAbductors, MuscleAdductors, MuscleBiceps,
	MuscleCalves, MuscleChest, MuscleForearms, MuscleGlutes, MuscleHamstrings,
	MuscleLats, MuscleLowerBack, MuscleMiddleBack, MuscleNeck,
	MuscleQuadriceps, MuscleShoulders, MuscleTraps, MuscleTriceps,
}

// Equipment is something an exercise needs.
type Equipment string

const (
	EquipmentBarbell      Equipment = "barbell"
	EquipmentDumbbell     Equipment = "dumbbell"
	EquipmentKettlebell   Equipment = "kettlebell"
	EquipmentCable        Equipment = "cable"
	EquipmentMachine      Equipment = "machine"
	EquipmentBench        Equipment = "bench"
	EquipmentPullUpBar    Equipment = "pull_up_bar"
	EquipmentBand         Equipment = "band"
	EquipmentTreadmill    Equipment = "treadmill"
	EquipmentEZBar        Equipment = "ez_bar"
	EquipmentMedicineBall Equipment = "medicine_ball"
	EquipmentExerciseBall Equipment = "exercise_ball"
	EquipmentFoamRoller   Equipment = "foam_roller"
)

// AllEquipment lists every valid Equipment (for validation and UI filters).
var AllEquipment = []Equipment{
	EquipmentBarbell, EquipmentDumbbell, EquipmentKettlebell, EquipmentCable,
	EquipmentMachine, EquipmentBench, EquipmentPullUpBar, EquipmentBand,
	EquipmentTreadmill, EquipmentEZBar, EquipmentMedicineBall,
	EquipmentExerciseBall, EquipmentFoamRoller,
}

// Category is the kind of training an exercise is (from free-exercise-db).
type Category string

const (
	CategoryStrength             Category = "strength"
	CategoryPowerlifting         Category = "powerlifting"
	CategoryOlympicWeightlifting Category = "olympic_weightlifting"
	CategoryStrongman            Category = "strongman"
	CategoryPlyometrics          Category = "plyometrics"
	CategoryStretching           Category = "stretching"
	CategoryCardio               Category = "cardio"
)

// AllCategories lists every valid Category (for validation and UI filters).
var AllCategories = []Category{
	CategoryStrength, CategoryPowerlifting, CategoryOlympicWeightlifting,
	CategoryStrongman, CategoryPlyometrics, CategoryStretching, CategoryCardio,
}

// Level is how much experience an exercise assumes.
type Level string

const (
	LevelBeginner     Level = "beginner"
	LevelIntermediate Level = "intermediate"
	LevelExpert       Level = "expert"
)

// AllLevels lists every valid Level (for validation and UI filters).
var AllLevels = []Level{LevelBeginner, LevelIntermediate, LevelExpert}

// ---- Planning ----

// Plan is a shared, reusable workout template, stored as
// plans/<id>.json. It suggests what to do; it never constrains a session.
// Starting a session from a plan pre-fills the session's entries.
type Plan struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	CreatedBy string         `json:"created_by"` // user ID
	Exercises []PlanExercise `json:"exercises"`  // in suggested order
	Archived  bool           `json:"archived,omitempty"`
}

// PlanExercise is one step of a plan. SuggestedSets is a hint shown
// in the UI (0 = no suggestion); sessions may log any number of sets.
type PlanExercise struct {
	ExerciseID    string `json:"exercise_id"`
	SuggestedSets int    `json:"suggested_sets,omitempty"`
}

// ---- Logging ----

// Session is one workout, stored as users/<user_id>/sessions/<id>.json.
// ID is assigned once from the start time and never recomputed.
// A zero EndedAt means the session is still in progress.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PlanID    string    `json:"plan_id,omitempty"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitzero"`
	Entries   []Entry   `json:"entries"`
	Archived  bool      `json:"archived,omitempty"`
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

// ---- Per-user settings and body tracking ----

// Profile is a user's fitness setup ("workout profile"), stored as
// users/<user_id>/fitness.json. The shared profile (name, color, units)
// lives in internal/users; this holds only what fitness needs. No file yet
// means the user hasn't done onboarding.
type Profile struct {
	UserID  string  `json:"user_id"`
	Goal    Goal    `json:"goal,omitempty"` // optional; not asked in onboarding
	HeightM float64 `json:"height_m,omitempty"`
	// Schedule is the person's routine: weekday → plan ID; missing = rest day.
	Schedule map[Weekday]string `json:"schedule,omitempty"`
	// WeightPromptSkipped is the ISO week ("2026-W41") whose weight
	// check-in was skipped, so home doesn't ask again that week.
	WeightPromptSkipped string `json:"weight_prompt_skipped,omitempty"`
}

// Goal is what a user trains for; it shapes advice (reps, rest), not data.
type Goal string

const (
	GoalStrength  Goal = "strength"
	GoalMuscle    Goal = "muscle"
	GoalEndurance Goal = "endurance"
	GoalGeneral   Goal = "general"
)

// AllGoals lists every valid Goal (for validation and UI options).
var AllGoals = []Goal{GoalStrength, GoalMuscle, GoalEndurance, GoalGeneral}

// Weekday keys a Schedule. Lowercase English names, so the JSON reads
// naturally: {"monday": "upper", "tuesday": "lower"}.
type Weekday string

// AllWeekdays lists the days Monday first (the UI's week order).
var AllWeekdays = []Weekday{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

// WeightEntry is one body-weight measurement. A user's entries are stored
// together, oldest first, in users/<user_id>/weights.json; at most one
// per day (saving the same date again replaces it).
type WeightEntry struct {
	Date     string  `json:"date"` // "YYYY-MM-DD"
	WeightKg float64 `json:"weight_kg"`
}
