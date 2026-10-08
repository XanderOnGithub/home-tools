package fitness

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/users"
)

// testUsers returns a shared profile store (in its own temp dir) holding
// one profile per id.
func testUsers(t *testing.T, ids ...string) *users.Store {
	t.Helper()
	us, err := users.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if err := us.Save(users.User{ID: id, Name: id, Color: users.ColorGreen, Units: users.UnitsMetric}); err != nil {
			t.Fatal(err)
		}
	}
	return us
}

func TestOpenRejectsBadData(t *testing.T) {
	squat := `{"id":"squat","name":"Squat","activation":{"quadriceps":1},"metrics":["reps","weight"]}`
	tests := []struct {
		name  string
		files map[string]string
	}{
		{"invalid exercise", map[string]string{
			"exercises/squat.json": `{"id":"squat","name":"Squat","activation":{"quads":1},"metrics":["reps"]}`,
		}},
		{"routine with unknown exercise", map[string]string{
			"routines/legs.json": `{"id":"legs","name":"Legs","created_by":"xander","exercises":[{"exercise_id":"squat"}]}`,
		}},
		{"session with unknown exercise", map[string]string{
			"users/xander/sessions/2026-10-08T18-00-00Z.json": `{"id":"2026-10-08T18-00-00Z","user_id":"xander","started_at":"2026-10-08T18:00:00Z","entries":[{"exercise_id":"squat","sets":[{"reps":5}]}]}`,
		}},
		{"session set with negative weight", map[string]string{
			"exercises/squat.json":                            squat,
			"users/xander/sessions/2026-10-08T18-00-00Z.json": `{"id":"2026-10-08T18-00-00Z","user_id":"xander","started_at":"2026-10-08T18:00:00Z","entries":[{"exercise_id":"squat","sets":[{"reps":5,"weight_kg":-1}]}]}`,
		}},
		{"routine with unknown creator", map[string]string{
			"exercises/squat.json": squat,
			"routines/legs.json":   `{"id":"legs","name":"Legs","created_by":"nobody","exercises":[{"exercise_id":"squat"}]}`,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for rel, body := range tt.files {
				path := filepath.Join(dir, rel)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := Open(dir, testUsers(t, "xander")); !errors.Is(err, ErrInvalid) {
				t.Errorf("Open() err = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestOpenSessionUserMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users", "alice", "sessions", "2026-10-08T18-00-00Z.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"id":"2026-10-08T18-00-00Z","user_id":"bob","started_at":"2026-10-08T18:00:00Z","entries":[]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	// alice is a real profile, so the user_id check is what must fail.
	if _, err := Open(dir, testUsers(t, "alice")); err == nil {
		t.Error("Open() err = nil, want user_id/folder mismatch error")
	}
}

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, body string) {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("exercises/squat.json", `{"id":"squat","name":"Squat","activation":{"quadriceps":1},"metrics":["reps","weight"]}`)
	write("routines/legs.json", `{"id":"legs","name":"Legs","created_by":"xander","exercises":[{"exercise_id":"squat"}]}`)
	write("users/xander/sessions/2026-10-08T18-00-00Z.json", `{"id":"2026-10-08T18-00-00Z","user_id":"xander","started_at":"2026-10-08T18:00:00Z","entries":[]}`)
	write("users/xander/sessions/2026-10-07T18-00-00Z.json", `{"id":"2026-10-07T18-00-00Z","user_id":"xander","started_at":"2026-10-07T18:00:00Z","entries":[]}`)

	s, err := Open(dir, testUsers(t, "xander"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.exercises["squat"]; !ok {
		t.Error("exercise squat not loaded")
	}
	if _, ok := s.routines["legs"]; !ok {
		t.Error("routine legs not loaded")
	}
	got := s.sessions["xander"]
	if len(got) != 2 || got[0].ID != "2026-10-07T18-00-00Z" {
		t.Errorf("sessions = %+v, want 2 sorted oldest first", got)
	}
}

func TestOpenEmptyDir(t *testing.T) {
	s, err := Open(t.TempDir(), testUsers(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.exercises)+len(s.routines)+len(s.sessions) != 0 {
		t.Error("fresh store should be empty")
	}
}

// newTestStore returns a store with one profile (xander) and one exercise
// (squat: reps + weight), saved to disk so a reopened store can resolve
// references to them.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir(), testUsers(t, "xander"))
	if err != nil {
		t.Fatal(err)
	}
	squat := Exercise{ID: "squat", Name: "Squat",
		Activation: map[Muscle]float64{MuscleQuadriceps: 1},
		Metrics:    []Metric{MetricReps, MetricWeight}}
	if err := s.SaveExercise(squat); err != nil {
		t.Fatal(err)
	}
	return s
}

func at(day, hour int) time.Time {
	return time.Date(2026, 10, day, hour, 0, 0, 0, time.UTC)
}

func ids(list []Session) []string {
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = s.ID
	}
	return out
}

func TestSaveSessionKeepsOrder(t *testing.T) {
	s := newTestStore(t)
	save := func(start time.Time) Session {
		t.Helper()
		sess, err := s.SaveSession(Session{UserID: "xander", StartedAt: start})
		if err != nil {
			t.Fatal(err)
		}
		return sess
	}

	save(at(7, 18))
	latest := save(at(9, 18))
	save(at(8, 18)) // logged after the fact: must land in the middle

	// Edit the latest (the common case): replaced, not duplicated.
	latest.Entries = []Entry{{ExerciseID: "squat", Sets: []Set{{Reps: 5, WeightKg: 100}}}}
	if _, err := s.SaveSession(latest); err != nil {
		t.Fatal(err)
	}

	want := []string{"2026-10-07T18-00-00Z", "2026-10-08T18-00-00Z", "2026-10-09T18-00-00Z"}
	got := ids(s.sessions["xander"])
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
	if len(s.sessions["xander"][2].Entries) != 1 {
		t.Error("edit to latest session not applied")
	}

	// Disk matches memory: a fresh Open sees the same data.
	reopened, err := Open(s.dir, s.users)
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(reopened.sessions["xander"]); !slices.Equal(got, want) {
		t.Errorf("after reopen = %v, want %v", got, want)
	}
}

func TestSaveSessionRejects(t *testing.T) {
	tests := []struct {
		name string
		sess Session
	}{
		{"unknown exercise", Session{UserID: "xander", StartedAt: at(7, 18),
			Entries: []Entry{{ExerciseID: "nope"}}}},
		{"set breaks exercise rules", Session{UserID: "xander", StartedAt: at(7, 18),
			Entries: []Entry{{ExerciseID: "squat", Sets: []Set{{Reps: 5}}}}}},
		{"path traversal user", Session{UserID: "../../etc", StartedAt: at(7, 18)}},
		{"missing start", Session{UserID: "xander"}},
		{"unknown user", Session{UserID: "xandr", StartedAt: at(7, 18)}},
		{"ends before start", Session{UserID: "xander", StartedAt: at(7, 18), EndedAt: at(7, 17)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			_, err := s.SaveSession(tt.sess)
			if !errors.Is(err, ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
			if len(s.sessions) != 0 {
				t.Error("rejected session reached memory")
			}
		})
	}
}

func TestRecentSessions(t *testing.T) {
	s := newTestStore(t)
	for day := 1; day <= 5; day++ {
		if _, err := s.SaveSession(Session{UserID: "xander", StartedAt: at(day, 18)}); err != nil {
			t.Fatal(err)
		}
	}

	got := s.RecentSessions("xander", 2)
	want := []string{"2026-10-05T18-00-00Z", "2026-10-04T18-00-00Z"}
	if !slices.Equal(ids(got), want) {
		t.Errorf("RecentSessions = %v, want %v", ids(got), want)
	}

	got[0].ID = "changed" // must not affect the store
	if s.sessions["xander"][4].ID == "changed" {
		t.Error("RecentSessions returned the store's own memory")
	}

	if n := len(s.RecentSessions("xander", 99)); n != 5 {
		t.Errorf("n larger than history: got %d, want 5", n)
	}
	if n := len(s.RecentSessions("nobody", 3)); n != 0 {
		t.Errorf("unknown user: got %d, want 0", n)
	}
}

// Run with -race: concurrent saves and reads must not corrupt the maps.
func TestStoreConcurrent(t *testing.T) {
	s := newTestStore(t)
	var wg sync.WaitGroup
	for day := 1; day <= 20; day++ {
		wg.Go(func() {
			if _, err := s.SaveSession(Session{UserID: "xander", StartedAt: at(day, 18)}); err != nil {
				t.Error(err)
			}
		})
		wg.Go(func() { s.RecentSessions("xander", 5) })
	}
	wg.Wait()

	got := ids(s.sessions["xander"])
	if len(got) != 20 || !slices.IsSorted(got) {
		t.Errorf("after concurrent saves: %d sessions, sorted=%v", len(got), slices.IsSorted(got))
	}
}

func TestRecentSessionsSkipsArchived(t *testing.T) {
	s := newTestStore(t)
	for day := 1; day <= 3; day++ {
		if _, err := s.SaveSession(Session{UserID: "xander", StartedAt: at(day, 18), Archived: day == 3}); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"2026-10-02T18-00-00Z", "2026-10-01T18-00-00Z"}
	if got := ids(s.RecentSessions("xander", 5)); !slices.Equal(got, want) {
		t.Errorf("RecentSessions = %v, want %v", got, want)
	}
}

func TestSaveExerciseAndList(t *testing.T) {
	s := newTestStore(t) // has "Squat"
	bench := Exercise{ID: "bench_press", Name: "Bench Press",
		Activation: map[Muscle]float64{MuscleChest: 1},
		Metrics:    []Metric{MetricReps, MetricWeight},
		Equipment:  []Equipment{EquipmentBarbell, EquipmentBench}}
	if err := s.SaveExercise(bench); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveExercise(Exercise{ID: "bad"}); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid exercise: err = %v, want ErrInvalid", err)
	}

	list := s.Exercises()
	if len(list) != 2 || list[0].Name != "Bench Press" || list[1].Name != "Squat" {
		t.Errorf("Exercises() not sorted by name: %+v", list)
	}

	reopened, err := Open(s.dir, s.users)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.exercises["bench_press"].Equipment; !slices.Equal(got, bench.Equipment) {
		t.Errorf("after reopen equipment = %v", got)
	}
}

func TestSaveRoutineChecksCatalog(t *testing.T) {
	s := newTestStore(t)
	legs := Routine{ID: "legs", Name: "Legs", CreatedBy: "xander",
		Exercises: []RoutineExercise{{ExerciseID: "squat", SuggestedSets: 4}}}
	if err := s.SaveRoutine(legs); err != nil {
		t.Fatal(err)
	}
	legs.Exercises = append(legs.Exercises, RoutineExercise{ExerciseID: "nope"})
	if err := s.SaveRoutine(legs); !errors.Is(err, ErrInvalid) {
		t.Errorf("unknown exercise: err = %v, want ErrInvalid", err)
	}
	stranger := Routine{ID: "arms", Name: "Arms", CreatedBy: "nobody",
		Exercises: []RoutineExercise{{ExerciseID: "squat"}}}
	if err := s.SaveRoutine(stranger); !errors.Is(err, ErrInvalid) {
		t.Errorf("unknown creator: err = %v, want ErrInvalid", err)
	}
	if got := s.Routines(); len(got) != 1 || len(got[0].Exercises) != 1 {
		t.Errorf("rejected save changed memory: %+v", got)
	}
}

func TestOpenFolderForUnknownProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users", "ghost", "sessions", "2026-10-08T18-00-00Z.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"id":"2026-10-08T18-00-00Z","user_id":"ghost","started_at":"2026-10-08T18:00:00Z","entries":[]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(dir, testUsers(t, "xander")); err == nil {
		t.Error("Open() err = nil, want error for folder with no profile")
	}
}

// A profile created after fitness started (e.g. from another tool's
// picker) can log sessions immediately: fitness asks the shared store
// on every save instead of keeping its own copy.
func TestSessionForNewProfile(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.SaveSession(Session{UserID: "missy", StartedAt: at(8, 7)}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("before profile exists: err = %v, want ErrInvalid", err)
	}
	if err := s.users.Save(users.User{ID: "missy", Name: "Missy", Color: users.ColorBlue, Units: users.UnitsImperial}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveSession(Session{UserID: "missy", StartedAt: at(8, 7)}); err != nil {
		t.Fatalf("after profile exists: %v", err)
	}
}
