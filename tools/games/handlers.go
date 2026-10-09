package games

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

const (
	inspectTimeout = 3 * time.Second
	logTail        = 200 // lines of history a log view starts with
)

type handlers struct {
	store  *Store
	docker *Docker // nil: no Docker socket configured (e.g. local dev)
	log    *slog.Logger
	snap   snapshot // last looked-up views, so the list answers instantly
	acts   activity // joins and leaves, read from the logs
	heads  *heads   // Minecraft player faces
}

// Register mounts the games API on mux. docker may be nil: servers are
// still listed and editable, but show as unavailable.
func Register(mux *http.ServeMux, store *Store, docker *Docker, log *slog.Logger) {
	h := &handlers{store: store, docker: docker, log: log, heads: newHeads()}
	mux.HandleFunc("GET /api/servers", h.getServers)
	mux.HandleFunc("PUT /api/servers/{id}", h.putServer)
	mux.HandleFunc("POST /api/servers/{id}/{action}", h.postAction)
	mux.HandleFunc("GET /api/servers/{id}/logs", h.getLogs)
	mux.HandleFunc("GET /api/servers/{id}/heads/{name}", h.getHead)
}

// serverView is what the browser gets about a server: its config minus
// secrets (no RCON password, no query address), what Docker says about it
// right now, and who's online. Error explains a missing State (Docker
// unreachable, no such container); PlayersError a missing Players.
type serverView struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Game         Game     `json:"game"`
	Container    string   `json:"container"`
	Archived     bool     `json:"archived,omitempty"`
	State        *State   `json:"state,omitempty"`
	Error        string   `json:"error,omitempty"`
	Players      *Players `json:"players,omitempty"`
	PlayersError string   `json:"players_error,omitempty"`
	Activity     []Event  `json:"activity,omitempty"` // joins and leaves, newest first
}

func viewOf(srv Server) serverView {
	return serverView{ID: srv.ID, Name: srv.Name, Game: srv.Game, Container: srv.Container, Archived: srv.Archived}
}

// fill asks Docker for srv's state and, if it's running, who's online:
// Minecraft over RCON, Valheim from its log. Joins and leaves come from
// the log for both. Errors are worded for people; details go to the log.
func (h *handlers) fill(ctx context.Context, v *serverView, srv Server) {
	st, err := h.inspect(ctx, srv)
	if err != nil {
		v.Error = err.Error()
		return
	}
	v.State = &st
	if !st.Running {
		v.Activity = h.acts.recentEvents(srv.ID) // the history stays visible
		return
	}
	online, events, logErr := h.acts.update(ctx, h.docker, srv, st)
	if logErr != nil {
		h.log.Warn("activity from log", "server", srv.ID, "err", logErr)
	}
	v.Activity = events

	switch srv.Game {
	case GameMinecraft:
		if srv.Query == "" {
			return
		}
		p, uuids, err := minecraftPlayers(ctx, srv)
		if err != nil {
			h.log.Warn("player query", "server", srv.ID, "err", err)
			v.PlayersError = "couldn't ask the game who's online"
			return
		}
		for name, uuid := range uuids {
			h.heads.remember(name, uuid)
		}
		v.Players = &p
	case GameValheim:
		if logErr != nil {
			v.PlayersError = "couldn't read the server's log"
			return
		}
		p := Players{Online: len(online), Names: online}
		if srv.Query != "" {
			if p.Max, err = valheimMax(ctx, srv); err != nil {
				h.log.Debug("valheim query", "server", srv.ID, "err", err)
			}
		}
		v.Players = &p
	}
}

// getServers lists every server with its live state and players, from
// the snapshot (instant; refreshed in the background when stale).
func (h *handlers) getServers(w http.ResponseWriter, r *http.Request) {
	views := h.snap.get(r.Context(), h.lookAll)
	if views == nil { // the client gave up while the first lookup ran
		return
	}
	httpx.WriteJSON(w, http.StatusOK, views)
}

// lookAll looks up every server in parallel, each step with its own short
// timeout, so one stuck container or game can't hold up the rest. It runs
// in the background (see snapshot), so it doesn't use a request's context.
func (h *handlers) lookAll() []serverView {
	servers := h.store.Servers()
	out := make([]serverView, len(servers))
	var wg sync.WaitGroup
	for i, srv := range servers {
		out[i] = viewOf(srv)
		if srv.Archived {
			continue
		}
		wg.Go(func() { h.fill(context.Background(), &out[i], srv) })
	}
	wg.Wait()
	return out
}

// inspect returns srv's state, with errors worded for the UI.
func (h *handlers) inspect(ctx context.Context, srv Server) (State, error) {
	if h.docker == nil {
		return State{}, errors.New("Docker isn't connected")
	}
	ctx, cancel := context.WithTimeout(ctx, inspectTimeout)
	defer cancel()
	st, err := h.docker.Inspect(ctx, srv.Container)
	switch {
	case errors.Is(err, ErrNotFound):
		return State{}, fmt.Errorf("no container named %q", srv.Container)
	case err != nil:
		h.log.Error("docker inspect", "server", srv.ID, "err", err)
		return State{}, errors.New("couldn't reach Docker")
	}
	return st, nil
}

// putServer creates or replaces a server's config. Like httpx.PutByID, but
// it answers with the view (no password), not the body it was sent.
func (h *handlers) putServer(w http.ResponseWriter, r *http.Request) {
	var srv Server
	if err := httpx.DecodeJSON(w, r, &srv); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if id := r.PathValue("id"); srv.ID != id {
		httpx.WriteError(w, http.StatusBadRequest, "body id "+srv.ID+" does not match URL id "+id)
		return
	}
	if err := h.store.SaveServer(srv); err != nil {
		if errors.Is(err, ErrInvalid) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.ServerError(w, r, h.log, err)
		return
	}
	h.snap.reset()
	httpx.WriteJSON(w, http.StatusOK, viewOf(srv))
}

// postAction starts, stops or restarts a server's container and answers
// with its new state. Stop and restart can take up to a minute (the game
// saves its world first), so the request waits that long.
func (h *handlers) postAction(w http.ResponseWriter, r *http.Request) {
	srv, ok := h.server(w, r)
	if !ok {
		return
	}
	a := Action(r.PathValue("action"))
	if !slices.Contains([]Action{ActionStart, ActionStop, ActionRestart}, a) {
		httpx.WriteError(w, http.StatusNotFound, "unknown action "+string(a))
		return
	}
	if h.docker == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "Docker isn't connected")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), (stopTimeoutSec+30)*time.Second)
	defer cancel()
	if err := h.docker.Do(ctx, srv.Container, a); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusBadGateway, fmt.Sprintf("no container named %q", srv.Container))
			return
		}
		h.log.Error("docker action", "server", srv.ID, "action", a, "err", err)
		httpx.WriteError(w, http.StatusBadGateway, "Docker couldn't "+string(a)+" it; see the server log")
		return
	}
	h.log.Info("server action", "server", srv.ID, "action", a)
	v := viewOf(srv)
	h.fill(r.Context(), &v, srv)
	h.snap.put(v) // the list shows the new state at once
	httpx.WriteJSON(w, http.StatusOK, v)
}

// getLogs streams a server's log as Server-Sent Events (decision #37): the
// last logTail lines, then new ones live. Each line is one "data:" event;
// "event: end" means the container stopped (the stream is over), and
// "event: failure" carries a message (not "error": EventSource uses that
// name for its own connection errors). The browser's EventSource reconnects
// by itself after a network blip; the UI clears its lines on (re)open,
// since the stream always starts with the tail again.
func (h *handlers) getLogs(w http.ResponseWriter, r *http.Request) {
	srv, ok := h.server(w, r)
	if !ok {
		return
	}
	if h.docker == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "Docker isn't connected")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	rc := http.NewResponseController(w)

	send := func(event, data string) error {
		var b strings.Builder
		if event != "" {
			b.WriteString("event: " + event + "\n")
		}
		// A data field can't contain a newline; lines never do here.
		b.WriteString("data: " + data + "\n\n")
		if _, err := w.Write([]byte(b.String())); err != nil {
			return err // the browser went away
		}
		return rc.Flush()
	}

	err := h.docker.Logs(r.Context(), srv.Container, logTail, func(line string) error {
		return send("", line)
	})
	switch {
	case r.Context().Err() != nil:
		// The viewer closed the page; nothing to say.
	case errors.Is(err, ErrNotFound):
		_ = send("failure", fmt.Sprintf("no container named %q", srv.Container))
	case err != nil:
		h.log.Error("docker logs", "server", srv.ID, "err", err)
		_ = send("failure", "lost the connection to Docker")
	default:
		_ = send("end", "")
	}
}

// getHead answers a Minecraft player's face as an 8×8 PNG (decision #40),
// or 404 when there's none (unknown player, offline-mode server, Mojang
// unreachable). Only players this server has listed have one.
func (h *handlers) getHead(w http.ResponseWriter, r *http.Request) {
	srv, ok := h.server(w, r)
	if !ok {
		return
	}
	if srv.Game != GameMinecraft {
		httpx.WriteError(w, http.StatusNotFound, "only Minecraft players have heads")
		return
	}
	face, err := h.heads.face(r.Context(), r.PathValue("name"))
	if err != nil {
		if !errors.Is(err, errNoHead) {
			h.log.Warn("player head", "server", srv.ID, "err", err)
		}
		httpx.WriteError(w, http.StatusNotFound, "no head for "+r.PathValue("name"))
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400") // a day, like the server's own cache
	w.Write(face)
}

// server looks up {id}, answering 404 itself when it's unknown.
func (h *handlers) server(w http.ResponseWriter, r *http.Request) (Server, bool) {
	srv, ok := h.store.Server(r.PathValue("id"))
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "no server "+r.PathValue("id"))
	}
	return srv, ok
}
