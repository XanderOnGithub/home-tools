package fitness

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

// progressStore: xander did squat on the 7th and 9th (twice on the 9th),
// bench only on the 8th; plus an archived and an unfinished workout with
// squats that must not count.
func progressStore(t *testing.T) *Store {
	t.Helper()
	s := newTestStore(t)
	bench := Exercise{ID: "bench", Name: "Bench",
		Activation: map[Muscle]float64{MuscleChest: 1},
		Metrics:    []Metric{MetricReps, MetricWeight}}
	if err := s.SaveExercise(bench); err != nil {
		t.Fatal(err)
	}
	squat := func(kg float64) Entry { return Entry{ExerciseID: "squat", Sets: []Set{{Reps: 5, WeightKg: kg}}} }
	for _, sess := range []Session{
		{StartedAt: at(7, 18), EndedAt: at(7, 19), Entries: []Entry{squat(100)}},
		{StartedAt: at(8, 18), EndedAt: at(8, 19), Entries: []Entry{
			{ExerciseID: "bench", Sets: []Set{{Reps: 8, WeightKg: 60}}},
			{ExerciseID: "squat"}, // picked, nothing logged
		}},
		{StartedAt: at(9, 18), EndedAt: at(9, 19), Entries: []Entry{squat(105), squat(110)}},
		{StartedAt: at(10, 18), EndedAt: at(10, 19), Entries: []Entry{squat(500)}, Archived: true},
		{StartedAt: at(11, 18), Entries: []Entry{squat(120)}}, // in progress
	} {
		sess.UserID = "xander"
		if _, err := s.SaveSession(sess); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func TestExerciseLog(t *testing.T) {
	s := progressStore(t)
	want := []LoggedExercise{
		{ExerciseID: "squat", LastDone: at(9, 18), Workouts: 2},
		{ExerciseID: "bench", LastDone: at(8, 18), Workouts: 1},
	}
	if got := s.ExerciseLog("xander"); !slices.Equal(got, want) {
		t.Errorf("ExerciseLog = %+v, want %+v", got, want)
	}
	if got := s.ExerciseLog("nobody"); len(got) != 0 {
		t.Errorf("ExerciseLog(nobody) = %+v, want empty", got)
	}
}

func TestExerciseHistory(t *testing.T) {
	s := progressStore(t)
	got := s.ExerciseHistory("xander", "squat")
	if len(got) != 2 {
		t.Fatalf("ExerciseHistory = %+v, want 2 workouts", got)
	}
	// Newest first; the 9th's two entries merged.
	if got[0].SessionID != "2026-10-09T18-00-00Z" || len(got[0].Sets) != 2 || got[1].Sets[0].WeightKg != 100 {
		t.Errorf("ExerciseHistory = %+v", got)
	}
	// The result is a copy.
	got[0].Sets[0].WeightKg = 1
	if again := s.ExerciseHistory("xander", "squat"); again[0].Sets[0].WeightKg != 105 {
		t.Error("ExerciseHistory returned the store's own memory")
	}
}

func TestProgressHandlers(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, progressStore(t), slog.New(slog.NewTextHandler(io.Discard, nil)))

	tests := []struct {
		path       string
		wantStatus int
		wantLen    int
	}{
		{"/api/users/xander/exercise-log", http.StatusOK, 2},
		{"/api/users/xander/exercise-log/squat", http.StatusOK, 2},
		{"/api/users/xander/exercise-log/bench", http.StatusOK, 1},
		{"/api/users/nobody/exercise-log", http.StatusNotFound, 0},
		{"/api/users/nobody/exercise-log/squat", http.StatusNotFound, 0},
		{"/api/users/xander/exercise-log/nope", http.StatusNotFound, 0},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest("GET", tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tt.wantStatus, rec.Body)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var list []json.RawMessage
			if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list) != tt.wantLen {
				t.Errorf("body = %s, want %d items", rec.Body, tt.wantLen)
			}
		})
	}
}
