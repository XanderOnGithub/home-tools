package fitness

import (
	"testing"
	"time"
)

func TestSetValidate(t *testing.T) {
	bench := Exercise{ID: "bench_press", Metrics: []Metric{MetricReps, MetricWeight}}
	treadmill := Exercise{ID: "treadmill", Metrics: []Metric{MetricDuration}}
	pullUp := Exercise{ID: "pull_up", Bodyweight: true, Metrics: []Metric{MetricReps, MetricWeight}}

	tests := []struct {
		name    string
		ex      Exercise
		set     Set
		wantErr bool
	}{
		{"bench ok", bench, Set{Reps: 8, WeightKg: 60}, false},
		{"bench zero reps", bench, Set{Reps: 0, WeightKg: 60}, true},
		{"bench negative weight", bench, Set{Reps: 8, WeightKg: -60}, true},
		{"bench untracked duration", bench, Set{Reps: 8, WeightKg: 60, DurationSec: 20}, true},
		{"bench missing weight", bench, Set{Reps: 8}, true},

		{"treadmill ok", treadmill, Set{DurationSec: 1200}, false},
		{"treadmill negative weight", treadmill, Set{DurationSec: 1200, WeightKg: -5}, true},
		{"treadmill untracked weight and reps", treadmill, Set{DurationSec: 1200, WeightKg: 5, Reps: 20}, true},
		{"treadmill missing duration", treadmill, Set{WeightKg: 5, Reps: 20}, true},

		{"pull-up bodyweight only", pullUp, Set{Reps: 10}, false},
		{"pull-up added weight", pullUp, Set{Reps: 8, WeightKg: 10}, false},
		{"pull-up zero reps", pullUp, Set{WeightKg: 10}, true},
		{"pull-up negative weight", pullUp, Set{Reps: 8, WeightKg: -10}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.set.Validate(tt.ex)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExerciseValidate(t *testing.T) {
	squat := Exercise{
		ID:         "squat",
		Name:       "Squat",
		Activation: map[Muscle]float64{MuscleQuadriceps: 1, MuscleGlutes: 0.5},
		Metrics:    []Metric{MetricReps, MetricWeight},
	}
	// with returns a copy of squat with one change applied, so each case
	// differs from the valid exercise in exactly one way.
	with := func(change func(*Exercise)) Exercise {
		e := squat
		change(&e)
		return e
	}

	tests := []struct {
		name    string
		ex      Exercise
		wantErr bool
	}{
		{"valid", squat, false},
		{"missing ID", with(func(e *Exercise) { e.ID = "" }), true},
		{"missing name", with(func(e *Exercise) { e.Name = "" }), true},
		{"missing metrics", with(func(e *Exercise) { e.Metrics = nil }), true},
		{"missing activation", with(func(e *Exercise) { e.Activation = nil }), true},
		{"activation exactly 1", with(func(e *Exercise) { e.Activation = map[Muscle]float64{MuscleQuadriceps: 1} }), false},
		{"activation zero", with(func(e *Exercise) { e.Activation = map[Muscle]float64{MuscleQuadriceps: 0} }), true},
		{"activation over 1", with(func(e *Exercise) { e.Activation = map[Muscle]float64{MuscleQuadriceps: 1.01} }), true},
		{"activation negative", with(func(e *Exercise) { e.Activation = map[Muscle]float64{MuscleQuadriceps: -0.5} }), true},
		{"unknown metric", with(func(e *Exercise) { e.Metrics = []Metric{"rep"} }), true},
		{"unknown muscle", with(func(e *Exercise) { e.Activation = map[Muscle]float64{"quadz": 1} }), true},
		{"unknown equipment", with(func(e *Exercise) { e.Equipment = []Equipment{"barbel"} }), true},
		{"unknown category", with(func(e *Exercise) { e.Category = "strenght" }), true},
		{"unknown level", with(func(e *Exercise) { e.Level = "pro" }), true},
		{"known category and level", with(func(e *Exercise) { e.Category, e.Level = CategoryStrength, LevelBeginner }), false},
		{"known equipment", with(func(e *Exercise) { e.Equipment = []Equipment{EquipmentBarbell, EquipmentBench} }), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ex.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoutineValidate(t *testing.T) {
	push := Routine{
		ID:        "push_day",
		Name:      "Push Day",
		CreatedBy: "xander",
		Exercises: []RoutineExercise{
			{ExerciseID: "bench_press", SuggestedSets: 4},
			{ExerciseID: "push_up"},
		},
	}
	// with copies push and applies one change. Exercises is a slice, so a
	// change must assign a new slice rather than edit push's elements.
	with := func(change func(*Routine)) Routine {
		r := push
		change(&r)
		return r
	}

	tests := []struct {
		name    string
		r       Routine
		wantErr bool
	}{
		{"valid", push, false},
		{"missing ID", with(func(r *Routine) { r.ID = "" }), true},
		{"missing name", with(func(r *Routine) { r.Name = "" }), true},
		{"missing created_by", with(func(r *Routine) { r.CreatedBy = "" }), true},
		{"no exercises", with(func(r *Routine) { r.Exercises = nil }), true},
		{"exercise missing ID", with(func(r *Routine) {
			r.Exercises = []RoutineExercise{{SuggestedSets: 3}}
		}), true},
		{"negative suggested sets", with(func(r *Routine) {
			r.Exercises = []RoutineExercise{{ExerciseID: "bench_press", SuggestedSets: -1}}
		}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserValidate(t *testing.T) {
	xander := User{
		ID:          "xander",
		Name:        "Xander",
		Birthday:    time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC),
		HeightM:     1.8,
		Units:       UnitsMetric,
		AvatarColor: "#4f8cff",
	}
	with := func(change func(*User)) User {
		u := xander
		change(&u)
		return u
	}

	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{"ok", xander, false},
		{"optional fields empty", with(func(u *User) { u.Birthday = time.Time{}; u.HeightM = 0 }), false},
		{"emoji avatar", with(func(u *User) { u.AvatarEmoji = "🏋️" }), false},
		{"uppercase hex", with(func(u *User) { u.AvatarColor = "#4F8CFF" }), false},
		{"bad id", with(func(u *User) { u.ID = "../root" }), true},
		{"blank name", with(func(u *User) { u.Name = "   " }), true},
		{"missing units", with(func(u *User) { u.Units = "" }), true},
		{"unknown units", with(func(u *User) { u.Units = "furlongs" }), true},
		{"negative height", with(func(u *User) { u.HeightM = -1 }), true},
		{"future birthday", with(func(u *User) { u.Birthday = time.Now().AddDate(1, 0, 0) }), true},
		{"missing color", with(func(u *User) { u.AvatarColor = "" }), true},
		{"color without #", with(func(u *User) { u.AvatarColor = "4f8cff0" }), true},
		{"short color", with(func(u *User) { u.AvatarColor = "#fff" }), true},
		{"non-hex color", with(func(u *User) { u.AvatarColor = "#zzzzzz" }), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
