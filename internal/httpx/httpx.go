// Package httpx holds small JSON helpers shared by every tool's handlers.
// It knows nothing about any tool: each tool decides which of its errors
// are the client's fault (400) and which are ours (500).
package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// MaxBodyBytes caps request bodies. Our largest payload is one session or
// exercise (a few KB); 1 MiB is generous and stops a runaway client.
const MaxBodyBytes = 1 << 20

// WriteJSON writes v as JSON with the given status. It marshals before
// writing anything, so an encoding failure becomes a clean 500 instead of
// a 200 with half a body (headers can't be changed once sent).
func WriteJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

// WriteError writes {"error": msg} with the given status. Only pass
// messages that are safe for the client to see.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// ServerError logs err with request context and sends the client a
// generic 500. Real errors (file paths, OS details) stay in the log.
func ServerError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	log.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	WriteError(w, http.StatusInternalServerError, "internal server error")
}

// DecodeJSON reads exactly one JSON value from r's body into dst.
// It is strict: unknown fields are rejected (a typo like "weigth_kg"
// would otherwise be silently dropped), as is trailing data and any body
// over MaxBodyBytes. Every error it returns is the client's fault, so
// callers can send err.Error() back with a 400.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body is empty")
		}
		return fmt.Errorf("bad JSON body: %w", err)
	}
	if dec.More() {
		return errors.New("bad JSON body: unexpected data after the JSON value")
	}
	return nil
}
