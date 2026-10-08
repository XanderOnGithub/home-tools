package users

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	xander := User{ID: "xander", Name: "Xander", Color: ColorGreen, Units: UnitsMetric}
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
		{"bad id", with(func(u *User) { u.ID = "../root" }), true},
		{"blank name", with(func(u *User) { u.Name = "   " }), true},
		{"missing color", with(func(u *User) { u.Color = "" }), true},
		{"unknown color", with(func(u *User) { u.Color = "red" }), true},
		{"missing units", with(func(u *User) { u.Units = "" }), true},
		{"unknown units", with(func(u *User) { u.Units = "furlongs" }), true},
		{"birthday", with(func(u *User) { u.Birthday = "2000-01-02" }), false},
		{"birthday not a date", with(func(u *User) { u.Birthday = "2000-02-30" }), true},
		{"birthday wrong format", with(func(u *User) { u.Birthday = "01/02/2000" }), true},
		{"birthday in the future", with(func(u *User) { u.Birthday = "2999-01-01" }), true},
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

func TestStoreSaveAndReopen(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []User{
		{ID: "xander", Name: "Xander", Color: ColorGreen, Units: UnitsMetric},
		{ID: "missy", Name: "Missy", Color: ColorBlue, Units: UnitsImperial},
	} {
		if err := s.Save(u); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Save(User{ID: "bad", Name: "Bad"}); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid user: err = %v, want ErrInvalid", err)
	}

	names := func(us []User) []string {
		out := make([]string, len(us))
		for i, u := range us {
			out[i] = u.Name
		}
		return out
	}
	want := []string{"Missy", "Xander"}
	if got := names(s.Users()); !slices.Equal(got, want) {
		t.Errorf("Users() = %v, want %v", got, want)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(reopened.Users()); !slices.Equal(got, want) {
		t.Errorf("after reopen = %v, want %v", got, want)
	}
	if u, ok := reopened.User("missy"); !ok || u.Units != UnitsImperial {
		t.Errorf("User(missy) = %+v, %v", u, ok)
	}
}

func TestHandlers(t *testing.T) {
	const missy = `{"id":"missy","name":"Missy","color":"blue","units":"imperial"}`
	newServer := func(t *testing.T) http.Handler {
		s, err := Open(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		mux := http.NewServeMux()
		Register(mux, s, slog.New(slog.NewTextHandler(io.Discard, nil)))
		return mux
	}
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"create", "/api/users/missy", missy, http.StatusOK},
		{"id mismatch", "/api/users/bob", missy, http.StatusBadRequest},
		{"fails validation", "/api/users/missy", `{"id":"missy","name":"Missy","color":"red","units":"metric"}`, http.StatusBadRequest},
		{"unknown field", "/api/users/missy", `{"id":"missy","name":"Missy","color":"blue","units":"metric","avatar_emoji":"x"}`, http.StatusBadRequest},
		{"path trick", "/api/users/..%2Froot", `{"id":"../root","name":"R","color":"blue","units":"metric"}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newServer(t).ServeHTTP(rec, httptest.NewRequest("PUT", tt.path, strings.NewReader(tt.body)))
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}

	srv := newServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/users", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "[]" {
		t.Errorf("empty list = %d %s, want 200 []", rec.Code, rec.Body)
	}
	srv.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("PUT", "/api/users/missy", strings.NewReader(missy)))
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/users", nil))
	var list []User
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != "missy" {
		t.Errorf("list = %+v, want [missy]", list)
	}
}
