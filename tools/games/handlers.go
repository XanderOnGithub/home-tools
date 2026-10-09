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
}

// Register mounts the games API on mux. docker may be nil: servers are
// still listed and editable, but show as unavailable.
func Register(mux *http.ServeMux, store *Store, docker *Docker, log *slog.Logger) {
	h := &handlers{store: store, docker: docker, log: log}
	mux.HandleFunc("GET /api/servers", h.getServers)
	mux.HandleFunc("PUT /api/servers/{id}", httpx.PutByID(log, func(s Server) string { return s.ID }, store.SaveServer, ErrInvalid))
	mux.HandleFunc("POST /api/servers/{id}/{action}", h.postAction)
	mux.HandleFunc("GET /api/servers/{id}/logs", h.getLogs)
}

// serverStatus is a server plus what Docker says about it right now.
// Error explains a missing State (Docker unreachable, no such container).
type serverStatus struct {
	Server
	State *State `json:"state,omitempty"`
	Error string `json:"error,omitempty"`
}

// getServers lists every server with its live state. Containers are
// inspected in parallel, each with its own short timeout, so one stuck
// lookup can't hold up the list.
func (h *handlers) getServers(w http.ResponseWriter, r *http.Request) {
	servers := h.store.Servers()
	out := make([]serverStatus, len(servers))
	var wg sync.WaitGroup
	for i, srv := range servers {
		out[i].Server = srv
		if srv.Archived {
			continue
		}
		wg.Go(func() {
			st, err := h.inspect(r.Context(), srv)
			if err != nil {
				out[i].Error = err.Error()
				return
			}
			out[i].State = &st
		})
	}
	wg.Wait()
	httpx.WriteJSON(w, http.StatusOK, out)
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
	st, err := h.inspect(r.Context(), srv)
	if err != nil {
		httpx.WriteJSON(w, http.StatusOK, serverStatus{Server: srv, Error: err.Error()})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, serverStatus{Server: srv, State: &st})
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

// server looks up {id}, answering 404 itself when it's unknown.
func (h *handlers) server(w http.ResponseWriter, r *http.Request) (Server, bool) {
	srv, ok := h.store.Server(r.PathValue("id"))
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "no server "+r.PathValue("id"))
	}
	return srv, ok
}
