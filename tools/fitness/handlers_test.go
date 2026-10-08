package fitness

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer mounts the fitness API over a fresh test store (which has
// one exercise: squat). Requests go through the real mux, so route
// patterns and methods are tested too, not just the handler funcs.
func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	Register(mux, newTestStore(t), slog.New(slog.NewTextHandler(io.Discard, nil)))
	return mux
}

func TestGetExercise(t *testing.T) {
	srv := newTestServer(t)
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"found", "GET", "/api/exercises/squat", http.StatusOK},
		{"unknown id", "GET", "/api/exercises/nope", http.StatusNotFound},
		{"wrong method", "DELETE", "/api/exercises/squat", http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}

	// The body round-trips into the model type.
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/exercises/squat", nil))
	var ex Exercise
	if err := json.Unmarshal(rec.Body.Bytes(), &ex); err != nil {
		t.Fatal(err)
	}
	if ex.ID != "squat" || ex.Name != "Squat" {
		t.Errorf("got %+v, want squat", ex)
	}
}

func TestPutExercise(t *testing.T) {
	const bench = `{"id":"bench","name":"Bench Press","activation":{"chest":1},"metrics":["reps","weight"]}`
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"create", "/api/exercises/bench", bench, http.StatusOK},
		{"update existing", "/api/exercises/squat",
			`{"id":"squat","name":"Back Squat","activation":{"quadriceps":1},"metrics":["reps","weight"]}`, http.StatusOK},
		{"id mismatch", "/api/exercises/deadlift", bench, http.StatusBadRequest},
		{"fails validation", "/api/exercises/bench",
			`{"id":"bench","name":"Bench Press","activation":{"pecs":1},"metrics":["reps"]}`, http.StatusBadRequest},
		{"unknown field", "/api/exercises/bench",
			`{"id":"bench","name":"Bench Press","nmae":"x","activation":{"chest":1},"metrics":["reps"]}`, http.StatusBadRequest},
		{"malformed JSON", "/api/exercises/bench", `{"id":`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer(t) // fresh store per case: no leakage between rows
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest("PUT", tt.path, strings.NewReader(tt.body)))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tt.wantStatus, rec.Body)
			}
			if rec.Code != http.StatusOK {
				return
			}
			// A successful PUT is visible to a following GET.
			get := httptest.NewRecorder()
			srv.ServeHTTP(get, httptest.NewRequest("GET", tt.path, nil))
			if get.Code != http.StatusOK {
				t.Errorf("GET after PUT = %d, want 200", get.Code)
			}
		})
	}
}

func TestRoutineHandlers(t *testing.T) {
	const legs = `{"id":"legs","name":"Legs","created_by":"xander","exercises":[{"exercise_id":"squat"}]}`
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"create", "/api/routines/legs", legs, http.StatusOK},
		{"id mismatch", "/api/routines/arms", legs, http.StatusBadRequest},
		{"unknown exercise", "/api/routines/legs",
			`{"id":"legs","name":"Legs","created_by":"xander","exercises":[{"exercise_id":"nope"}]}`, http.StatusBadRequest},
		{"fails validation", "/api/routines/legs",
			`{"id":"legs","name":"Legs","created_by":"xander","exercises":[]}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer(t)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest("PUT", tt.path, strings.NewReader(tt.body)))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}

	// Empty list is 200 + [], then the created routine shows up.
	srv := newTestServer(t)
	list := func() string {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/routines", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/routines = %d, want 200", rec.Code)
		}
		return rec.Body.String()
	}
	if got := list(); got != "[]" {
		t.Errorf("empty list = %s, want []", got)
	}
	srv.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("PUT", "/api/routines/legs", strings.NewReader(legs)))
	if got := list(); !strings.Contains(got, `"id":"legs"`) {
		t.Errorf("list after PUT = %s, want it to contain legs", got)
	}
}
