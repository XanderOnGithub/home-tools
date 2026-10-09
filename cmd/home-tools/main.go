// Command home-tools is the single server binary. It loads config, mounts each
// tool (tools/<name>) under its hostname, and serves API + embedded SPAs.
//
//	go run ./cmd/home-tools -addr :8080 -data data            # route by subdomain
//	go run ./cmd/home-tools -tool fitness                     # local dev: one tool on every host
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/hostroute"
	"github.com/XanderOnGithub/home-tools/internal/httpx"
	"github.com/XanderOnGithub/home-tools/internal/users"
	"github.com/XanderOnGithub/home-tools/tools/fitness"
	"github.com/XanderOnGithub/home-tools/tools/games"
	fitnessweb "github.com/XanderOnGithub/home-tools/web/apps/fitness"
)

// config is everything the command line sets.
type config struct {
	addr    string // listen address
	dataDir string // one subfolder per tool
	tool    string // serve only this tool (local dev); "" = route by subdomain
	docker  string // Docker API socket for games (the socket proxy's); "" = none
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8080", "listen address")
	flag.StringVar(&cfg.dataDir, "data", "data", "data folder (one subfolder per tool)")
	flag.StringVar(&cfg.tool, "tool", "", "serve only this tool, on every host (local dev: localhost has no subdomain); empty = route by subdomain")
	flag.StringVar(&cfg.docker, "docker", "", "Docker API unix socket for the games tool (the socket proxy's); empty = games can't see or control servers")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(cfg, log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// run is main with errors as values: it returns instead of exiting, so
// deferred cleanup runs and the logic stays testable.
func run(cfg config, log *slog.Logger) error {
	// Cancelled on Ctrl-C (SIGINT) or `docker stop` (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Shared profiles first: tools check their data against them.
	us, err := users.Open(filepath.Join(cfg.dataDir, "users"))
	if err != nil {
		return fmt.Errorf("open users store: %w", err)
	}
	store, err := fitness.Open(filepath.Join(cfg.dataDir, "fitness"), us)
	if err != nil {
		return fmt.Errorf("open fitness store: %w", err)
	}
	gameServers, err := games.Open(filepath.Join(cfg.dataDir, "games"))
	if err != nil {
		return fmt.Errorf("open games store: %w", err)
	}
	var docker *games.Docker
	if cfg.docker != "" {
		docker = games.NewDocker(cfg.docker)
	}
	log.Info("stores loaded", "users", len(us.Users()), "exercises", len(store.Exercises()), "plans", len(store.Plans()))

	// Each tool gets its own mux on its own subdomain (#5, ADR 0003).
	tools := hostroute.New(cfg.tool)
	tools.Handle("fitness", toolMux(us, fitnessweb.Dist, log, func(mux *http.ServeMux) {
		fitness.Register(mux, store, log)
	}))
	tools.Handle("games", toolMux(us, nil, log, func(mux *http.ServeMux) {
		games.Register(mux, gameServers, docker, log)
	}))
	if cfg.tool != "" && !slices.Contains(tools.Names(), cfg.tool) {
		return fmt.Errorf("-tool %q: no such tool (have %v)", cfg.tool, tools.Names())
	}
	log.Info("tools mounted", "tools", tools.Names(), "only", cfg.tool, "docker", cfg.docker != "")

	// Health checks answer on any host, including a bare IP.
	root := http.NewServeMux()
	root.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})
	root.Handle("/", tools)

	srv := &http.Server{
		Addr:    cfg.addr,
		Handler: root,
		// Slow-client limits. No WriteTimeout: it would cut off long-lived
		// responses like the games tool's live log stream.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	// ListenAndServe blocks, so it runs in a goplan while this one
	// waits for either a server failure or a shutdown signal.
	errc := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", cfg.addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return err // failed to start, e.g. port already in use
	case <-ctx.Done():
	}

	// Graceful shutdown: stop accepting, let in-flight requests finish.
	// Store writes happen inside requests, so no write is cut off midway.
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// toolMux builds one tool's mux: the shared profiles API (every tool shows
// the same picker, ADR 0007), the tool's own API via register, and its UI
// when built in (-tags webembed; dist is nil otherwise). Unknown API paths
// stay a JSON 404 instead of falling through to index.html.
func toolMux(us *users.Store, dist fs.FS, log *slog.Logger, register func(*http.ServeMux)) *http.ServeMux {
	mux := http.NewServeMux()
	users.Register(mux, us, log)
	register(mux)
	if dist != nil {
		mux.HandleFunc("GET /api/", func(w http.ResponseWriter, r *http.Request) {
			httpx.WriteError(w, http.StatusNotFound, "not found")
		})
		mux.Handle("GET /", httpx.SPA(dist))
	}
	return mux
}
