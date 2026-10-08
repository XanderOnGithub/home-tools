package httpx

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]int{"n": 1})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	if got := rec.Body.String(); got != `{"n":1}` {
		t.Errorf("body = %s", got)
	}
}

func TestWriteJSONUnencodable(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, make(chan int)) // channels can't be JSON
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestServerErrorHidesDetails(t *testing.T) {
	var logs bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logs, nil))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/x", nil)

	ServerError(rec, req, log, errors.New("open /data/secret.json: permission denied"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret") {
		t.Errorf("body leaks the real error: %s", rec.Body.String())
	}
	if !strings.Contains(logs.String(), "secret") {
		t.Errorf("log is missing the real error: %s", logs.String())
	}
}

func TestDecodeJSON(t *testing.T) {
	type thing struct {
		Name string `json:"name"`
	}
	tests := []struct {
		name    string
		body    io.Reader
		wantErr bool
	}{
		{"ok", strings.NewReader(`{"name":"squat"}`), false},
		{"empty", strings.NewReader(``), true},
		{"malformed", strings.NewReader(`{"name":`), true},
		{"unknown field", strings.NewReader(`{"name":"squat","nmae":"x"}`), true},
		{"trailing data", strings.NewReader(`{"name":"a"}{"name":"b"}`), true},
		{"too large", strings.NewReader(`{"name":"` + strings.Repeat("a", MaxBodyBytes) + `"}`), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("PUT", "/", tt.body)
			var got thing
			err := DecodeJSON(httptest.NewRecorder(), req, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
