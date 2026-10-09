package fitness

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestPlanHandlers(t *testing.T) {
	const legs = `{"id":"legs","name":"Legs","created_by":"xander","exercises":[{"exercise_id":"squat"}]}`
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"create", "/api/plans/legs", legs, http.StatusOK},
		{"id mismatch", "/api/plans/arms", legs, http.StatusBadRequest},
		{"unknown exercise", "/api/plans/legs",
			`{"id":"legs","name":"Legs","created_by":"xander","exercises":[{"exercise_id":"nope"}]}`, http.StatusBadRequest},
		{"fails validation", "/api/plans/legs",
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

	// Empty list is 200 + [], then the created plan shows up.
	srv := newTestServer(t)
	list := func() string {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/plans", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/plans = %d, want 200", rec.Code)
		}
		return rec.Body.String()
	}
	if got := list(); got != "[]" {
		t.Errorf("empty list = %s, want []", got)
	}
	srv.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("PUT", "/api/plans/legs", strings.NewReader(legs)))
	if got := list(); !strings.Contains(got, `"id":"legs"`) {
		t.Errorf("list after PUT = %s, want it to contain legs", got)
	}
}

func TestSessionHandlers(t *testing.T) {
	const start = `{"user_id":"xander","started_at":"2026-10-08T18:00:00Z","entries":[]}`
	srv := newTestServer(t)
	do := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		return rec
	}

	// Start a session: the server assigns the ID from started_at.
	rec := do("POST", "/api/users/xander/sessions", start)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST = %d, want 201; body %s", rec.Code, rec.Body)
	}
	var sess Session
	if err := json.Unmarshal(rec.Body.Bytes(), &sess); err != nil {
		t.Fatal(err)
	}
	if sess.ID != "2026-10-08T18-00-00Z" {
		t.Fatalf("assigned ID = %q", sess.ID)
	}

	// Log a set and finish it with a PUT.
	update := `{"id":"2026-10-08T18-00-00Z","user_id":"xander","started_at":"2026-10-08T18:00:00Z",` +
		`"ended_at":"2026-10-08T19:00:00Z","entries":[{"exercise_id":"squat","sets":[{"reps":5,"weight_kg":100}]}]}`
	if rec := do("PUT", "/api/users/xander/sessions/2026-10-08T18-00-00Z", update); rec.Code != http.StatusOK {
		t.Fatalf("PUT = %d, want 200; body %s", rec.Code, rec.Body)
	}

	// The list shows the updated session.
	rec = do("GET", "/api/users/xander/sessions?limit=5", "")
	var list []Session
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || len(list[0].Entries) != 1 {
		t.Errorf("GET sessions = %+v, want 1 session with 1 entry", list)
	}

	errorCases := []struct {
		name, method, path, body string
		wantStatus               int
	}{
		{"list unknown user", "GET", "/api/users/nobody/sessions", "", http.StatusNotFound},
		{"bad limit", "GET", "/api/users/xander/sessions?limit=0", "", http.StatusBadRequest},
		{"POST with id", "POST", "/api/users/xander/sessions",
			`{"id":"x","user_id":"xander","started_at":"2026-10-08T18:00:00Z","entries":[]}`, http.StatusBadRequest},
		{"POST user mismatch", "POST", "/api/users/alice/sessions", start, http.StatusBadRequest},
		{"POST unknown user", "POST", "/api/users/nobody/sessions",
			`{"user_id":"nobody","started_at":"2026-10-08T18:00:00Z","entries":[]}`, http.StatusBadRequest},
		{"POST unknown exercise", "POST", "/api/users/xander/sessions",
			`{"user_id":"xander","started_at":"2026-10-09T18:00:00Z","entries":[{"exercise_id":"nope","sets":[]}]}`, http.StatusBadRequest},
		{"PUT id mismatch", "PUT", "/api/users/xander/sessions/other", update, http.StatusBadRequest},
		// Same second as the session above: must not overwrite it.
		{"POST same second", "POST", "/api/users/xander/sessions", start, http.StatusConflict},
	}
	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(tt.method, tt.path, tt.body); rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}

func TestImages(t *testing.T) {
	store := newTestStore(t)
	img := filepath.Join(store.dir, "images", "Squat", "0.jpg")
	if err := os.MkdirAll(filepath.Dir(img), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(img, []byte("fake jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A file outside images/ that must stay unreachable.
	if err := os.WriteFile(filepath.Join(store.dir, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	Register(mux, store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"photo", "/images/Squat/0.jpg", http.StatusOK},
		{"missing photo", "/images/Squat/9.jpg", http.StatusNotFound},
		{"no directory listing", "/images/Squat/", http.StatusNotFound},
		{"no root listing", "/images/", http.StatusNotFound},
		{"escape attempt", "/images/..%2Fsecret.txt", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest("GET", tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if strings.Contains(rec.Body.String(), "secret") {
				t.Error("served a file outside images/")
			}
		})
	}
}
