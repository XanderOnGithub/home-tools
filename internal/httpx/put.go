package httpx

import (
	"errors"
	"log/slog"
	"net/http"
)

// PutByID returns a handler for "PUT /…/{id}" that creates or replaces one
// resource: it decodes a T strictly, checks the body's ID matches the
// URL's {id}, saves it, and answers 200 with it.
//
// PUT, not POST: the client names the resource, and sending the same body
// twice leaves the same result (idempotent). The URL says which file gets
// written, so the body must agree, or a PUT to /squat could overwrite
// /deadlift.
//
// A save error that errors.Is invalid (the tool's validation sentinel,
// wrapped with details written for users) becomes a 400 with its message.
// Anything else (disk full, permissions) is ours: a generic 500, details
// in the log.
//
// T is a type parameter, like a TS generic: one function, checked at
// compile time for each type it's used with.
func PutByID[T any](log *slog.Logger, id func(T) string, save func(T) error, invalid error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var v T
		if err := DecodeJSON(w, r, &v); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if urlID := r.PathValue("id"); id(v) != urlID {
			WriteError(w, http.StatusBadRequest, "body id "+id(v)+" does not match URL id "+urlID)
			return
		}
		if err := save(v); err != nil {
			if errors.Is(err, invalid) {
				WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			ServerError(w, r, log, err)
			return
		}
		WriteJSON(w, http.StatusOK, v)
	}
}
