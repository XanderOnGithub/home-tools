package fitness

import (
	"testing"
	"time"
)

func TestRecovery(t *testing.T) {
	now := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)
	catalog := map[string]Exercise{
		"bench":   {ID: "bench", Category: CategoryStrength, Activation: map[Muscle]float64{MuscleChest: 1, MuscleTriceps: 0.5, MuscleShoulders: 0.5}},
		"squat":   {ID: "squat", Category: CategoryStrength, Activation: map[Muscle]float64{MuscleQuadriceps: 1, MuscleGlutes: 0.5}},
		"stretch": {ID: "stretch", Category: CategoryStretching, Activation: map[Muscle]float64{MuscleChest: 1}},
	}
	exercise := func(id string) (Exercise, bool) { ex, ok := catalog[id]; return ex, ok }
	workout := func(ago time.Duration, entries ...Entry) Session {
		return Session{StartedAt: now.Add(-ago), Entries: entries}
	}
	sets := func(id string, n int) Entry { return Entry{ExerciseID: id, Sets: make([]Set, n)} }
	// Done 10 days ago: too old to tire anything, recent enough that the
	// exercises aren't new (no novelty factor), so the cases below show
	// volume and decay alone.
	familiar := workout(10*24*time.Hour, sets("bench", 1), sets("squat", 1))
	known := func(s ...Session) []Session { return append([]Session{familiar}, s...) }

	type want struct {
		state   RecoveryState
		readyIn time.Duration // 0 = ready now
		sets7d  float64
	}
	tests := []struct {
		name     string
		sessions []Session // oldest → newest
		want     map[Muscle]want
	}{
		{"nothing logged", nil, map[Muscle]want{MuscleChest: {StateReady, 0, 0}}},
		{
			// chest 6 → ready after 24 h · log2(6) ≈ 62 h; triceps 3 (half
			// credit), 3 is exactly "rest", and arms are slower (26.4 h
			// half-life): 26.4 h · log2(3) ≈ 41.8 h.
			"6 sets of bench just now", known(workout(0, sets("bench", 6))),
			map[Muscle]want{
				MuscleChest:      {StateRest, 62*time.Hour + 2*time.Minute, 6},
				MuscleTriceps:    {StateRest, 41*time.Hour + 50*time.Minute, 3},
				MuscleQuadriceps: {StateReady, 0, 0},
			},
		},
		{
			// 48 h later: chest 6/4 = 1.5 → recovering for 24 h · log2(1.5) ≈ 14 h;
			// triceps 3 · 2^(-48/26.4) ≈ 0.85 → ready.
			"6 sets of bench two days ago", known(workout(48*time.Hour, sets("bench", 6))),
			map[Muscle]want{
				MuscleChest:   {StateRecovering, 14*time.Hour + 2*time.Minute, 6},
				MuscleTriceps: {StateReady, 0, 3},
			},
		},
		{
			// Quads fade faster (21.6 h): 2 · 2^(-24/21.6) ≈ 0.93 left of a
			// day ago + 2 now = 2.93 → recovering, ready in 21.6 h · log2(2.93) ≈ 33.5 h.
			"fatigue adds up across workouts", known(workout(24*time.Hour, sets("squat", 2)), workout(0, sets("squat", 2))),
			map[Muscle]want{MuscleQuadriceps: {StateRecovering, 33*time.Hour + 28*time.Minute, 4}},
		},
		{
			// A first time counts 1.5×: 3 sets of a new exercise = 4.5 →
			// ready in 24 h · log2(4.5) ≈ 52 h, not the 38 h of a familiar one.
			"a new exercise counts more", []Session{workout(0, sets("bench", 3))},
			map[Muscle]want{MuscleChest: {StateRest, 52*time.Hour + 5*time.Minute, 3}},
		},
		{
			// 20 days off = new again; the familiar session doesn't help.
			"back after a layoff", []Session{workout(20*24*time.Hour, sets("bench", 3)), workout(0, sets("bench", 3))},
			map[Muscle]want{MuscleChest: {StateRest, 52*time.Hour + 5*time.Minute, 3}},
		},
		{
			// 10 days since: still familiar, 3 sets = exactly 3 → 38 h.
			"done within two weeks", known(workout(0, sets("bench", 3))),
			map[Muscle]want{MuscleChest: {StateRest, 38*time.Hour + 2*time.Minute, 3}},
		},
		{
			"stretches, discarded workouts and old ones don't count", []Session{
				workout(8*24*time.Hour, sets("bench", 10)),
				{StartedAt: now.Add(-time.Hour), Entries: []Entry{sets("bench", 10)}, Archived: true},
				workout(0, sets("stretch", 5), sets("unknown", 3)),
			},
			map[Muscle]want{MuscleChest: {StateReady, 0, 0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := recovery(tt.sessions, exercise, now)
			if len(got) != len(AllMuscles) {
				t.Fatalf("%d muscles, want all %d", len(got), len(AllMuscles))
			}
			for _, r := range got {
				w, ok := tt.want[r.Muscle]
				if !ok {
					continue
				}
				if r.State != w.state || r.Sets7d != w.sets7d {
					t.Errorf("%s: state %s, sets %v; want %s, %v", r.Muscle, r.State, r.Sets7d, w.state, w.sets7d)
				}
				var readyIn time.Duration
				if !r.ReadyAt.IsZero() {
					readyIn = r.ReadyAt.Sub(now)
				}
				if d := readyIn - w.readyIn; d < -2*time.Minute || d > 2*time.Minute {
					t.Errorf("%s: ready in %v, want about %v", r.Muscle, readyIn, w.readyIn)
				}
				if w.state != StateReady && r.LastWorked.IsZero() {
					t.Errorf("%s: worked but no last_worked", r.Muscle)
				}
			}
		})
	}
}
