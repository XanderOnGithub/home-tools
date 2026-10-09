package httpx

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPutByID(t *testing.T) {
	type thing struct {
		ID   string `json:"id"`
		Size int    `json:"size"`
	}
	invalid := errors.New("invalid")
	var saved []thing
	save := func(v thing) error {
		switch v.Size {
		case -1:
			return fmt.Errorf("%w: size must be positive", invalid)
		case 500:
			return errors.New("disk full")
		}
		saved = append(saved, v)
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("PUT /things/{id}", PutByID(slog.New(slog.NewTextHandler(io.Discard, nil)),
		func(v thing) string { return v.ID }, save, invalid))

	tests := []struct {
		name, path, body string
		want             int
	}{
		{"ok", "/things/a", `{"id":"a","size":1}`, http.StatusOK},
		{"id mismatch", "/things/b", `{"id":"a","size":1}`, http.StatusBadRequest},
		{"unknown field", "/things/a", `{"id":"a","sise":1}`, http.StatusBadRequest},
		{"invalid", "/things/a", `{"id":"a","size":-1}`, http.StatusBadRequest},
		{"server error", "/things/a", `{"id":"a","size":500}`, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.body)))
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d; body %s", rec.Code, tt.want, rec.Body)
			}
		})
	}
	if len(saved) != 1 {
		t.Errorf("saved %d things, want 1", len(saved))
	}
}
