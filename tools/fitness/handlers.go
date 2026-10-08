package fitness

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

// handlers holds what every fitness handler needs. Methods on a struct
// (instead of plain funcs) let us pass dependencies explicitly: no globals.
type handlers struct {
	store *Store
	log   *slog.Logger
}

// Register mounts the fitness API on mux.
func Register(mux *http.ServeMux, store *Store, log *slog.Logger) {
	h := &handlers{store: store, log: log}

	// "METHOD /path/{param}": the mux matches the method too, and answers
	// 405 for the wrong one. A method value (h.getExercise) is a func
	// with h already bound, like .bind(this) in JS.
	mux.HandleFunc("GET /api/exercises/{id}", h.getExercise)
	mux.HandleFunc("GET /api/exercises", h.getExercises)
	mux.HandleFunc("PUT /api/exercises/{id}", h.putExercise)
	mux.HandleFunc("GET /api/routines", h.getRoutines)
	mux.HandleFunc("PUT /api/routines/{id}", h.putRoutine)
}

// getExercise returns one exercise, or 404 if the ID is unknown.
func (h *handlers) getExercise(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	ex, ok := h.store.Exercise(id)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "exercise "+id+" not found")
		return // writing a response doesn't stop the function; return does
	}
	httpx.WriteJSON(w, http.StatusOK, ex)
}

// putExercise creates or replaces the exercise at /api/exercises/{id}.
// PUT, not POST: the client names the resource, and sending the same body
// twice leaves the same result (idempotent), matching SaveExercise.
// Archiving is a PUT with "archived": true.
func (h *handlers) putExercise(w http.ResponseWriter, r *http.Request) {
	var ex Exercise
	if err := httpx.DecodeJSON(w, r, &ex); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// The URL says which file gets written; the body must agree, or a PUT
	// to /squat could silently overwrite /deadlift.
	if id := r.PathValue("id"); ex.ID != id {
		httpx.WriteError(w, http.StatusBadRequest, "body id "+ex.ID+" does not match URL id "+id)
		return
	}

	if err := h.store.SaveExercise(ex); err != nil {
		// errors.Is walks the %w chain: the store wraps ErrInvalid with
		// details, so this matches any validation failure. That message
		// was written for users; anything else (disk full, permissions)
		// is ours, so the client gets a generic 500 and the log gets it all.
		if errors.Is(err, ErrInvalid) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.ServerError(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ex)
}

// getExercises returns all exercises, or 200 if no exercises are found
func (h *handlers) getExercises(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.store.Exercises())
}

// getRoutines returns all routines sorted by name, archived included.
// An empty list is still a success: 200 with [].
func (h *handlers) getRoutines(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.store.Routines())
}

// putRoutine creates or replaces the routine at /api/routines/{id}.
// Same shape as putExercise; SaveRoutine also rejects unknown exercise
// IDs with ErrInvalid, so those become 400s through the same check.
func (h *handlers) putRoutine(w http.ResponseWriter, r *http.Request) {
	var rt Routine // not "r": that name is taken by the request
	if err := httpx.DecodeJSON(w, r, &rt); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if id := r.PathValue("id"); rt.ID != id {
		httpx.WriteError(w, http.StatusBadRequest, "body id "+rt.ID+" does not match URL id "+id)
		return
	}
	if err := h.store.SaveRoutine(rt); err != nil {
		if errors.Is(err, ErrInvalid) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.ServerError(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rt)
}
