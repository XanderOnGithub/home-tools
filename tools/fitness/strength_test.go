package fitness

import (
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEstimatedMax(t *testing.T) {
	tests := []struct {
		set  Set
		want float64
	}{
		{Set{Reps: 1, WeightKg: 100}, 100}, // a single is the max
		{Set{Reps: 5, WeightKg: 90}, 105},  // 90 × (1 + 5/30)
		{Set{Reps: 12, WeightKg: 60}, 84},  // 60 × 1.4: the last rep count that counts
		{Set{Reps: 13, WeightKg: 60}, 0},   // too many reps to estimate from
		{Set{Reps: 8}, 0},                  // no weight
		{Set{WeightKg: 60, DurationSec: 30}, 0},
	}
	for _, tt := range tests {
		if got := estimatedMax(tt.set); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("estimatedMax(%+v) = %v, want %v", tt.set, got, tt.want)
		}
	}
}

func TestStrength(t *testing.T) {
	now := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)
	day := 24 * time.Hour
	catalog := map[string]Exercise{
		"bench":  {ID: "bench", Category: CategoryStrength, Metrics: []Metric{MetricReps, MetricWeight}, Activation: map[Muscle]float64{MuscleChest: 1, MuscleTriceps: 0.5}},
		"fly":    {ID: "fly", Category: CategoryStrength, Metrics: []Metric{MetricReps, MetricWeight}, Activation: map[Muscle]float64{MuscleChest: 1}},
		"pushup": {ID: "pushup", Category: CategoryStrength, Bodyweight: true, Metrics: []Metric{MetricReps, MetricWeight}, Activation: map[Muscle]float64{MuscleChest: 1, MuscleTriceps: 1}},
		"run":    {ID: "run", Category: CategoryCardio, Metrics: []Metric{MetricDuration, MetricDistance}, Activation: map[Muscle]float64{MuscleQuadriceps: 1}},
		"plank":  {ID: "plank", Category: CategoryStrength, Metrics: []Metric{MetricDuration}, Activation: map[Muscle]float64{MuscleAbdominals: 1}},
	}
	exercise := func(id string) (Exercise, bool) { ex, ok := catalog[id]; return ex, ok }
	workout := func(ago time.Duration, entries ...Entry) Session {
		at := now.Add(-ago)
		return Session{StartedAt: at, EndedAt: at.Add(time.Hour), Entries: entries}
	}
	entry := func(id string, sets ...Set) Entry { return Entry{ExerciseID: id, Sets: sets} }

	sessions := []Session{
		// Bench: best 100 kg single 5 weeks ago (previous window), now 5 × 90 = 105: +5%.
		workout(35*day, entry("bench", Set{Reps: 1, WeightKg: 100}, Set{Reps: 15, WeightKg: 60})),
		// Fly: first done 20 days ago at 10 × 15 (20 kg max), now 10 × 18 (24): +20% since start.
		workout(20*day, entry("fly", Set{Reps: 10, WeightKg: 15})),
		// Push-ups: 20 reps 6 weeks ago, 25 now; a weighted set doesn't count as reps.
		workout(42*day, entry("pushup", Set{Reps: 20})),
		workout(2*day,
			entry("bench", Set{Reps: 5, WeightKg: 90}),
			entry("fly", Set{Reps: 10, WeightKg: 18}),
			entry("pushup", Set{Reps: 25}, Set{Reps: 40, WeightKg: 10}),
			entry("run", Set{DurationSec: 1800, DistanceM: 5000}),
			entry("plank", Set{DurationSec: 60}),
		),
		// Unfinished and discarded workouts don't count.
		{StartedAt: now.Add(-time.Hour), Entries: []Entry{entry("bench", Set{Reps: 1, WeightKg: 200})}},
		{StartedAt: now.Add(-2 * time.Hour), EndedAt: now, Archived: true, Entries: []Entry{entry("bench", Set{Reps: 1, WeightKg: 200})}},
	}
	got := strength(sessions, exercise, now)

	wantEx := map[string]ExerciseStrength{
		"bench":  {Measure: MeasureMaxKg, Best: 105, Recent: 105, Change: 0.05, Basis: BasisFourWeeks},
		"fly":    {Measure: MeasureMaxKg, Best: 24, Recent: 24, Change: 0.2, Basis: BasisSinceStart},
		"pushup": {Measure: MeasureReps, Best: 25, Recent: 25, Change: 0.25, Basis: BasisFourWeeks},
	}
	if len(got.Exercises) != len(wantEx) {
		t.Fatalf("exercises = %+v, want bench, fly and pushup only (no cardio, no holds)", got.Exercises)
	}
	for _, es := range got.Exercises {
		w := wantEx[es.ExerciseID]
		w.ExerciseID = es.ExerciseID
		if math.Abs(es.Best-w.Best) > 1e-9 || math.Abs(es.Recent-w.Recent) > 1e-9 || es.Change != w.Change || es.Basis != w.Basis || es.Measure != w.Measure {
			t.Errorf("%s = %+v, want %+v", es.ExerciseID, es, w)
		}
	}

	// Chest: bench and push-ups have 4-week trends, so the newer fly's
	// "since start" is left out: (5% + 25%) / 2 = 15%. Triceps: push-ups
	// only (bench works it as a helper, activation 0.5).
	wantMuscles := map[Muscle]MuscleStrength{
		MuscleChest:   {Change: 0.15, Basis: BasisFourWeeks, Exercises: []string{"bench", "pushup"}},
		MuscleTriceps: {Change: 0.25, Basis: BasisFourWeeks, Exercises: []string{"pushup"}},
	}
	if len(got.Muscles) != len(wantMuscles) {
		t.Fatalf("muscles = %+v", got.Muscles)
	}
	for _, ms := range got.Muscles {
		w := wantMuscles[ms.Muscle]
		if ms.Change != w.Change || ms.Basis != w.Basis || len(ms.Exercises) != len(w.Exercises) {
			t.Errorf("%s = %+v, want %+v", ms.Muscle, ms, w)
		}
	}

	// Someone with one workout: numbers, but nothing to compare with yet.
	first := strength([]Session{workout(time.Hour, entry("bench", Set{Reps: 5, WeightKg: 60}))}, exercise, now)
	if len(first.Muscles) != 0 || first.Exercises[0].Basis != "" {
		t.Errorf("first workout = %+v", first)
	}
}

func TestStrengthHandler(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, progressStore(t), slog.New(slog.NewTextHandler(io.Discard, nil)))
	for path, want := range map[string]int{"/api/users/xander/strength": 200, "/api/users/nobody/strength": 404} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != want {
			t.Errorf("%s = %d, want %d", path, rec.Code, want)
		}
		if want == 200 && !strings.Contains(rec.Body.String(), `"muscles":[`) {
			t.Errorf("%s body = %s", path, rec.Body)
		}
	}
}
