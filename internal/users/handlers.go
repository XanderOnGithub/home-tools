package users

import (
	"log/slog"
	"net/http"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

type handlers struct {
	store *Store
}

// Register mounts the profiles API on mux. Every tool's host gets it, so
// each app can show the same picker.
func Register(mux *http.ServeMux, store *Store, log *slog.Logger) {
	h := &handlers{store: store}
	mux.HandleFunc("GET /api/users", h.getUsers)
	mux.HandleFunc("PUT /api/users/{id}", httpx.PutByID(log, func(u User) string { return u.ID }, store.Save, ErrInvalid))
}

// getUsers returns every profile sorted by name, archived included.
// Empty (fresh install) is 200 with [].
func (h *handlers) getUsers(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.store.Users())
}
