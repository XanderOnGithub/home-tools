package fitness

import (
	"errors"
	"slices"
	"testing"
)

func TestParseFreeExerciseDB(t *testing.T) {
	data := []byte(`[
	{"id":"Barbell_Bench_Press_-_Medium_Grip","name":"Barbell Bench Press - Medium Grip",
	 "force":"push","level":"beginner","mechanic":"compound","equipment":"barbell",
	 "primaryMuscles":["chest"],"secondaryMuscles":["shoulders","triceps"],
	 "instructions":["Lie back.","Press."],"category":"strength",
	 "images":["Barbell_Bench_Press_-_Medium_Grip/0.jpg","Barbell_Bench_Press_-_Medium_Grip/1.jpg"]},
	{"id":"Pullups","name":"Pullups","force":"pull","level":"beginner","mechanic":"compound",
	 "equipment":"body only","primaryMuscles":["lats"],"secondaryMuscles":["biceps","middle back"],
	 "instructions":[],"category":"strength","images":[]},
	{"id":"Hamstring_Stretch","name":"Hamstring Stretch","force":null,"level":"beginner",
	 "mechanic":null,"equipment":null,"primaryMuscles":["hamstrings","lower back"],
	 "secondaryMuscles":["lower back"],"instructions":[],"category":"stretching","images":[]}
	]`)

	got, err := ParseFreeExerciseDB(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d exercises, want 3", len(got))
	}

	bench, pull, stretch := got[0], got[1], got[2]
	if bench.Activation[MuscleChest] != 1 || bench.Activation[MuscleTriceps] != 0.5 {
		t.Errorf("bench activation = %v", bench.Activation)
	}
	if bench.Bodyweight || !slices.Equal(bench.Equipment, []Equipment{EquipmentBarbell}) {
		t.Errorf("bench: bodyweight=%v equipment=%v", bench.Bodyweight, bench.Equipment)
	}
	if !slices.Equal(bench.Metrics, []Metric{MetricReps, MetricWeight}) || bench.Category != CategoryStrength {
		t.Errorf("bench: metrics=%v category=%q", bench.Metrics, bench.Category)
	}
	if !pull.Bodyweight || len(pull.Equipment) != 0 || pull.Activation[MuscleMiddleBack] != 0.5 {
		t.Errorf("pullups: bodyweight=%v equipment=%v activation=%v", pull.Bodyweight, pull.Equipment, pull.Activation)
	}
	if !slices.Equal(stretch.Metrics, []Metric{MetricDuration}) {
		t.Errorf("stretch metrics = %v, want [duration]", stretch.Metrics)
	}
	if stretch.Activation[MuscleLowerBack] != 1 {
		t.Errorf("muscle in primary and secondary should be 1.0, got %v", stretch.Activation[MuscleLowerBack])
	}
}

func TestParseFreeExerciseDBHolds(t *testing.T) {
	// "strength" in the dataset, but a plank is held, not repeated.
	data := []byte(`[{"id":"Plank","name":"Plank","level":"beginner","equipment":"body only",
	 "primaryMuscles":["abdominals"],"secondaryMuscles":[],"instructions":[],"category":"strength","images":[]}]`)
	got, err := ParseFreeExerciseDB(data)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got[0].Metrics, []Metric{MetricDuration}) {
		t.Errorf("plank metrics = %v, want [duration]", got[0].Metrics)
	}
}

func TestParseFreeExerciseDBRejectsUnknown(t *testing.T) {
	tests := map[string]string{
		"equipment": `[{"id":"x","name":"X","level":"beginner","equipment":"spaceship","primaryMuscles":["chest"],"category":"strength"}]`,
		"muscle":    `[{"id":"x","name":"X","level":"beginner","equipment":"barbell","primaryMuscles":["wings"],"category":"strength"}]`,
		"category":  `[{"id":"x","name":"X","level":"beginner","equipment":"barbell","primaryMuscles":["chest"],"category":"yoga"}]`,
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseFreeExerciseDB([]byte(data)); !errors.Is(err, ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
		})
	}
}
