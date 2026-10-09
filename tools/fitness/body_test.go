package fitness

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestProfileValidate(t *testing.T) {
	base := Profile{UserID: "xander", Goal: GoalMuscle, HeightM: 1.8,
		Schedule: map[Weekday]string{"monday": "upper"}}
	with := func(change func(*Profile)) Profile {
		p := base
		p.Schedule = map[Weekday]string{"monday": "upper"}
		change(&p)
		return p
	}
	tests := []struct {
		name    string
		p       Profile
		wantErr bool
	}{
		{"ok", base, false},
		{"no height, no schedule", with(func(p *Profile) { p.HeightM = 0; p.Schedule = nil }), false},
		{"skipped week", with(func(p *Profile) { p.WeightPromptSkipped = "2026-W41" }), false},
		{"bad user", with(func(p *Profile) { p.UserID = "../x" }), true},
		{"no goal", with(func(p *Profile) { p.Goal = "" }), false},
		{"unknown goal", with(func(p *Profile) { p.Goal = "bulk" }), true},
		{"height in cm by mistake", with(func(p *Profile) { p.HeightM = 180 }), true},
		{"negative height", with(func(p *Profile) { p.HeightM = -1 }), true},
		{"unknown weekday", with(func(p *Profile) { p.Schedule["funday"] = "upper" }), true},
		{"bad plan id", with(func(p *Profile) { p.Schedule["monday"] = "a/b" }), true},
		{"bad week", with(func(p *Profile) { p.WeightPromptSkipped = "2026-41" }), true},
		{"week 54", with(func(p *Profile) { p.WeightPromptSkipped = "2026-W54" }), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWeightEntryValidate(t *testing.T) {
	tests := []struct {
		name    string
		w       WeightEntry
		wantErr bool
	}{
		{"ok", WeightEntry{"2026-10-05", 80.5}, false},
		{"bad date", WeightEntry{"2026-13-01", 80}, true},
		{"future", WeightEntry{"2999-01-01", 80}, true},
		{"pounds by mistake", WeightEntry{"2026-10-05", 800}, true},
		{"zero", WeightEntry{"2026-10-05", 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.w.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// newBodyStore is newTestStore plus one plan ("upper") to schedule.
func newBodyStore(t *testing.T) *Store {
	t.Helper()
	s := newTestStore(t)
	upper := Plan{ID: "upper", Name: "Upper", CreatedBy: "xander",
		Exercises: []PlanExercise{{ExerciseID: "squat"}}}
	if err := s.SavePlan(upper); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSaveProfile(t *testing.T) {
	s := newBodyStore(t)
	if _, ok := s.Profile("xander"); ok {
		t.Fatal("new user should have no profile (= onboarding)")
	}

	p := Profile{UserID: "xander", Goal: GoalStrength, Schedule: map[Weekday]string{"monday": "upper"}}
	if err := s.SaveProfile(p); err != nil {
		t.Fatal(err)
	}
	// The caller's map is copied: changing it later doesn't touch the store.
	p.Schedule["tuesday"] = "upper"
	if got, _ := s.Profile("xander"); len(got.Schedule) != 1 {
		t.Errorf("store shares the caller's map: %v", got.Schedule)
	}

	for name, bad := range map[string]Profile{
		"unknown plan": {UserID: "xander", Goal: GoalStrength, Schedule: map[Weekday]string{"friday": "nope"}},
		"unknown user": {UserID: "nobody", Goal: GoalStrength},
	} {
		if err := s.SaveProfile(bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}

	reopened, err := Open(s.dir, s.users)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := reopened.Profile("xander"); !ok || got.Goal != GoalStrength || got.Schedule["monday"] != "upper" {
		t.Errorf("after reopen = %+v, %v", got, ok)
	}
}

func TestSaveWeight(t *testing.T) {
	s := newBodyStore(t)
	for _, w := range []WeightEntry{
		{"2026-10-05", 81},
		{"2026-09-28", 82},   // logged out of order: lands first
		{"2026-10-05", 80.5}, // same day again: replaces
	} {
		if err := s.SaveWeight("xander", w); err != nil {
			t.Fatal(err)
		}
	}
	want := []WeightEntry{{"2026-09-28", 82}, {"2026-10-05", 80.5}}
	if got := s.Weights("xander"); !slices.Equal(got, want) {
		t.Errorf("Weights() = %v, want %v", got, want)
	}
	if err := s.SaveWeight("nobody", WeightEntry{"2026-10-05", 80}); !errors.Is(err, ErrInvalid) {
		t.Errorf("unknown user: err = %v, want ErrInvalid", err)
	}

	reopened, err := Open(s.dir, s.users)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.Weights("xander"); !slices.Equal(got, want) {
		t.Errorf("after reopen = %v, want %v", got, want)
	}
}

func TestOpenRejectsBadBodyFiles(t *testing.T) {
	tests := map[string]struct{ file, body string }{
		"profile user mismatch": {"fitness.json", `{"user_id":"missy","goal":"strength"}`},
		"profile invalid":       {"fitness.json", `{"user_id":"xander","goal":"bulk"}`},
		"profile unknown plan": {"fitness.json",
			`{"user_id":"xander","goal":"strength","schedule":{"monday":"nope"}}`},
		"duplicate weight dates": {"weights.json",
			`[{"date":"2026-10-05","weight_kg":80},{"date":"2026-10-05","weight_kg":81}]`},
		"invalid weight": {"weights.json", `[{"date":"2026-10-05","weight_kg":-1}]`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "users", "xander", tt.file)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tt.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := Open(dir, testUsers(t, "xander")); err == nil {
				t.Error("Open() err = nil, want an error")
			}
		})
	}
}

func TestBodyHandlers(t *testing.T) {
	srv := newTestServer(t)
	do := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		return rec
	}

	// Not onboarded yet → 404; weights start empty as [].
	if rec := do("GET", "/api/users/xander/fitness", ""); rec.Code != http.StatusNotFound {
		t.Errorf("GET fitness before onboarding = %d, want 404", rec.Code)
	}
	if rec := do("GET", "/api/users/xander/weights", ""); rec.Code != http.StatusOK || rec.Body.String() != "[]" {
		t.Errorf("GET weights = %d %s, want 200 []", rec.Code, rec.Body)
	}

	// Onboard.
	if rec := do("PUT", "/api/users/xander/fitness", `{"user_id":"xander","goal":"muscle","height_m":1.8}`); rec.Code != http.StatusOK {
		t.Fatalf("PUT fitness = %d %s", rec.Code, rec.Body)
	}
	var p Profile
	if err := json.Unmarshal(do("GET", "/api/users/xander/fitness", "").Body.Bytes(), &p); err != nil || p.Goal != GoalMuscle {
		t.Errorf("GET fitness after onboarding = %+v, %v", p, err)
	}

	// Log a weight.
	if rec := do("PUT", "/api/users/xander/weights/2026-10-05", `{"date":"2026-10-05","weight_kg":80.5}`); rec.Code != http.StatusOK {
		t.Fatalf("PUT weight = %d %s", rec.Code, rec.Body)
	}

	errorCases := []struct {
		name, method, path, body string
		want                     int
	}{
		{"fitness user mismatch", "PUT", "/api/users/xander/fitness", `{"user_id":"missy","goal":"muscle"}`, http.StatusBadRequest},
		{"fitness invalid", "PUT", "/api/users/xander/fitness", `{"user_id":"xander","goal":"bulk"}`, http.StatusBadRequest},
		{"weight date mismatch", "PUT", "/api/users/xander/weights/2026-10-06", `{"date":"2026-10-05","weight_kg":80}`, http.StatusBadRequest},
		{"weight invalid", "PUT", "/api/users/xander/weights/2026-10-05", `{"date":"2026-10-05","weight_kg":5}`, http.StatusBadRequest},
		{"weights unknown user", "GET", "/api/users/nobody/weights", "", http.StatusNotFound},
	}
	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(tt.method, tt.path, tt.body); rec.Code != tt.want {
				t.Errorf("status = %d, want %d; body %s", rec.Code, tt.want, rec.Body)
			}
		})
	}
}
