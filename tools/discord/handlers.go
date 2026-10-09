package discord

import (
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

// The settings page's API (decision #42). LAN-only like every tool (#2);
// no login (#10). The bot token never appears here: it's an environment
// variable, not a setting.

type handlers struct {
	store *Store
	bot   *Bot
	log   *slog.Logger
}

// Register mounts the Discord tool's API on mux.
func Register(mux *http.ServeMux, store *Store, bot *Bot, log *slog.Logger) {
	h := &handlers{store: store, bot: bot, log: log}
	mux.HandleFunc("GET /api/config", h.getConfig)
	mux.HandleFunc("PUT /api/config", h.putConfig)
	mux.HandleFunc("GET /api/bot", h.getBot)
	mux.HandleFunc("DELETE /api/requests/{user}", h.deleteRequest)
	mux.HandleFunc("GET /api/persona.gif", h.getPersonaImage)
	mux.HandleFunc("GET /api/persona.png", h.getPersonaImage) // still: for reduced motion
}

func (h *handlers) getConfig(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.store.Config())
}

// putConfig replaces the whole config. The body's revision must be the
// current one (409 otherwise), so a stale tab can't undo newer changes.
func (h *handlers) putConfig(w http.ResponseWriter, r *http.Request) {
	var cfg Config
	if err := httpx.DecodeJSON(w, r, &cfg); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// JSON null for a list decodes to nil; store them as empty lists.
	if cfg.Names == nil {
		cfg.Names = []string{}
	}
	if cfg.StatusBoards == nil {
		cfg.StatusBoards = []StatusBoard{}
	}
	if cfg.Verified == nil {
		cfg.Verified = []VerifiedUser{}
	}
	saved, err := h.store.SaveConfig(cfg)
	switch {
	case errors.Is(err, ErrInvalid):
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, ErrConflict):
		httpx.WriteError(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		httpx.ServerError(w, r, h.log, err)
		return
	}
	h.log.Info("discord settings saved", "revision", saved.Revision)
	h.bot.Changed()
	httpx.WriteJSON(w, http.StatusOK, saved)
}

// botView is everything the settings page shows besides the config.
type botView struct {
	BotView
	Requests     []AccessRequest `json:"requests"`
	Servers      []serverOption  `json:"servers"` // for picking a status board's server
	ServersError string          `json:"servers_error,omitempty"`
}

type serverOption struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Game      string `json:"game"`
	Whitelist bool   `json:"whitelist"`
}

func (h *handlers) getBot(w http.ResponseWriter, r *http.Request) {
	v := botView{BotView: h.bot.View(), Requests: h.store.Requests(), Servers: []serverOption{}}
	list, err := h.bot.gameServers(r.Context(), 10*time.Second)
	if err != nil {
		h.log.Warn("discord games list", "err", err)
		v.ServersError = "Couldn't reach the games tool, so its servers can't be listed right now."
	}
	for _, s := range list {
		v.Servers = append(v.Servers, serverOption{ID: s.ID, Name: s.Name, Game: s.Game, Whitelist: s.Whitelist})
	}
	slices.SortFunc(v.Servers, func(a, b serverOption) int { return compareFold(a.Name, b.Name) })
	httpx.WriteJSON(w, http.StatusOK, v)
}

func (h *handlers) deleteRequest(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DismissRequest(r.PathValue("user")); err != nil {
		httpx.ServerError(w, r, h.log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getPersonaImage serves today's avatar, exactly what Discord gets: the
// animated GIF, or the still PNG at /api/persona.png. The ETag is the
// persona, so browsers revalidate cheaply and pick up a new day (or a
// renamed list) at once.
func (h *handlers) getPersonaImage(w http.ResponseWriter, r *http.Request) {
	p := h.bot.Persona()
	still := strings.HasSuffix(r.URL.Path, ".png")
	h32 := fnv.New32a()
	h32.Write([]byte(p.Name)) // names may be non-ASCII; headers shouldn't be
	etag := fmt.Sprintf(`"%s-%s-%x-%t"`, p.Date, p.Color, h32.Sum32(), still)
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache") // always revalidate (cheap: 304)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	gifData, pngData, err := h.bot.avatars.get(p)
	if err != nil {
		httpx.ServerError(w, r, h.log, err)
		return
	}
	if still {
		w.Header().Set("Content-Type", "image/png")
		w.Write(pngData)
		return
	}
	w.Header().Set("Content-Type", "image/gif")
	w.Write(gifData)
}
