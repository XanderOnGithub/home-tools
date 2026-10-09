package fitness

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	mux.HandleFunc("GET /api/plans", h.getPlans)
	mux.HandleFunc("PUT /api/plans/{id}", h.putPlan)
	// Profiles themselves (GET/PUT /api/users) come from internal/users.
	mux.HandleFunc("GET /api/users/{user}/sessions", h.getSessions)
	mux.HandleFunc("POST /api/users/{user}/sessions", h.postSession)
	mux.HandleFunc("PUT /api/users/{user}/sessions/{id}", h.putSession)
	mux.HandleFunc("GET /api/users/{user}/fitness", h.getProfile)
	mux.HandleFunc("PUT /api/users/{user}/fitness", h.putProfile)
	mux.HandleFunc("GET /api/users/{user}/weights", h.getWeights)
	mux.HandleFunc("PUT /api/users/{user}/weights/{date}", h.putWeight)

	// Exercise photos: an exercise's "images": ["Barbell_Squat/0.jpg"] is
	// served at /images/Barbell_Squat/0.jpg.
	mux.Handle("GET /images/", http.StripPrefix("/images/", imageServer(filepath.Join(store.dir, "images"))))
}

// imageServer serves files from dir. os.DirFS refuses paths that escape
// dir (like "../"), so a crafted URL can't read other files. Directory
// listings are turned off: only exact file paths work.
func imageServer(dir string) http.Handler {
	files := http.FileServerFS(os.DirFS(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		// Photos never change once imported, so browsers may cache them
		// for a day instead of re-asking on every page view.
		w.Header().Set("Cache-Control", "public, max-age=86400")
		files.ServeHTTP(w, r)
	})
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

// getPlans returns all plans sorted by name, archived included.
// An empty list is still a success: 200 with [].
func (h *handlers) getPlans(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.store.Plans())
}

// putPlan creates or replaces the plan at /api/plans/{id}.
// Same shape as putExercise; SavePlan also rejects unknown exercise
// IDs with ErrInvalid, so those become 400s through the same check.
func (h *handlers) putPlan(w http.ResponseWriter, r *http.Request) {
	var rt Plan // not "r": that name is taken by the request
	if err := httpx.DecodeJSON(w, r, &rt); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if id := r.PathValue("id"); rt.ID != id {
		httpx.WriteError(w, http.StatusBadRequest, "body id "+rt.ID+" does not match URL id "+id)
		return
	}
	if err := h.store.SavePlan(rt); err != nil {
		if errors.Is(err, ErrInvalid) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.ServerError(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rt)
}

// getSessions returns the user's most recent non-archived sessions, newest
// first. ?limit=n picks how many (default 20, max 100).
func (h *handlers) getSessions(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	if _, ok := h.store.users.User(user); !ok {
		httpx.WriteError(w, http.StatusNotFound, "user "+user+" not found")
		return
	}

	limit := 20
	if q := r.URL.Query().Get("limit"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil || n < 1 || n > 100 {
			httpx.WriteError(w, http.StatusBadRequest, "limit must be a number from 1 to 100")
			return
		}
		limit = n
	}
	httpx.WriteJSON(w, http.StatusOK, h.store.RecentSessions(user, limit))
}

// postSession starts (or logs) a new session. POST, not PUT: the server
// picks the ID from started_at, so the client sends no id.
func (h *handlers) postSession(w http.ResponseWriter, r *http.Request) {
	var sess Session
	if err := httpx.DecodeJSON(w, r, &sess); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if sess.ID != "" {
		httpx.WriteError(w, http.StatusBadRequest, "id is assigned by the server; use PUT to update a session")
		return
	}
	h.saveSession(w, r, sess, http.StatusCreated)
}

// putSession replaces an existing session (adding sets, finishing it,
// archiving it). The URL and body must agree on both user and id.
func (h *handlers) putSession(w http.ResponseWriter, r *http.Request) {
	var sess Session
	if err := httpx.DecodeJSON(w, r, &sess); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if id := r.PathValue("id"); sess.ID != id {
		httpx.WriteError(w, http.StatusBadRequest, "body id "+sess.ID+" does not match URL id "+id)
		return
	}
	h.saveSession(w, r, sess, http.StatusOK)
}

// saveSession is the shared tail of POST and PUT: check the user in the
// URL matches the body, save, and map errors. status is the success code.
func (h *handlers) saveSession(w http.ResponseWriter, r *http.Request, sess Session, status int) {
	if user := r.PathValue("user"); sess.UserID != user {
		httpx.WriteError(w, http.StatusBadRequest, "body user_id "+sess.UserID+" does not match URL user "+user)
		return
	}
	saved, err := h.store.SaveSession(sess)
	if err != nil {
		if errors.Is(err, ErrInvalid) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.ServerError(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, status, saved) // saved has the server-assigned ID
}

// getProfile returns the user's fitness profile. 404 means either the
// user doesn't exist or hasn't onboarded; the UI already knows the user
// exists (it just picked them), so 404 = show onboarding.
func (h *handlers) getProfile(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	p, ok := h.store.Profile(user)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "no fitness profile for "+user)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

// putProfile creates or replaces the user's fitness profile (onboarding,
// schedule edits, skipping a weight check-in).
func (h *handlers) putProfile(w http.ResponseWriter, r *http.Request) {
	var p Profile
	if err := httpx.DecodeJSON(w, r, &p); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if user := r.PathValue("user"); p.UserID != user {
		httpx.WriteError(w, http.StatusBadRequest, "body user_id "+p.UserID+" does not match URL user "+user)
		return
	}
	if err := h.store.SaveProfile(p); err != nil {
		h.saveError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

// getWeights returns the user's weight log, oldest first ([] if empty).
func (h *handlers) getWeights(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	if _, ok := h.store.users.User(user); !ok {
		httpx.WriteError(w, http.StatusNotFound, "user "+user+" not found")
		return
	}
	list := h.store.Weights(user)
	if list == nil {
		list = []WeightEntry{} // JSON [] instead of null
	}
	httpx.WriteJSON(w, http.StatusOK, list)
}

// putWeight records the weight for one date (replacing that day's entry).
func (h *handlers) putWeight(w http.ResponseWriter, r *http.Request) {
	var entry WeightEntry
	if err := httpx.DecodeJSON(w, r, &entry); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if date := r.PathValue("date"); entry.Date != date {
		httpx.WriteError(w, http.StatusBadRequest, "body date "+entry.Date+" does not match URL date "+date)
		return
	}
	if err := h.store.SaveWeight(r.PathValue("user"), entry); err != nil {
		h.saveError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, entry)
}

// saveError maps a store save error: rule violations are the client's
// fault (400, with the message); anything else is ours (500, logged).
func (h *handlers) saveError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrInvalid) {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.ServerError(w, r, h.log, err)
}
