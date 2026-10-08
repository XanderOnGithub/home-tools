package users

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

type handlers struct {
	store *Store
	log   *slog.Logger
}

// Register mounts the profiles API on mux. Every tool's host gets it, so
// each app can show the same picker.
func Register(mux *http.ServeMux, store *Store, log *slog.Logger) {
	h := &handlers{store: store, log: log}
	mux.HandleFunc("GET /api/users", h.getUsers)
	mux.HandleFunc("PUT /api/users/{id}", h.putUser)
}

// getUsers returns every profile sorted by name, archived included.
// Empty (fresh install) is 200 with [].
func (h *handlers) getUsers(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.store.Users())
}

// putUser creates or replaces the profile at /api/users/{id}.
func (h *handlers) putUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := httpx.DecodeJSON(w, r, &u); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if id := r.PathValue("id"); u.ID != id {
		httpx.WriteError(w, http.StatusBadRequest, "body id "+u.ID+" does not match URL id "+id)
		return
	}
	if err := h.store.Save(u); err != nil {
		if errors.Is(err, ErrInvalid) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.ServerError(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}
