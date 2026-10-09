// Command home-tools is the single server binary. It loads config, mounts each
// tool (tools/<name>) under its hostname, and serves API + embedded SPAs.
//
//	go run ./cmd/home-tools -addr :8080 -data data
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
	"github.com/XanderOnGithub/home-tools/internal/users"
	"github.com/XanderOnGithub/home-tools/tools/fitness"
	fitnessweb "github.com/XanderOnGithub/home-tools/web/apps/fitness"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dataDir := flag.String("data", "data", "data folder (one subfolder per tool)")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(*addr, *dataDir, log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// run is main with errors as values: it returns instead of exiting, so
// deferred cleanup runs and the logic stays testable.
func run(addr, dataDir string, log *slog.Logger) error {
	// Cancelled on Ctrl-C (SIGINT) or `docker stop` (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Shared profiles first: tools check their data against them.
	us, err := users.Open(filepath.Join(dataDir, "users"))
	if err != nil {
		return fmt.Errorf("open users store: %w", err)
	}
	store, err := fitness.Open(filepath.Join(dataDir, "fitness"), us)
	if err != nil {
		return fmt.Errorf("open fitness store: %w", err)
	}
	log.Info("stores loaded", "users", len(us.Users()), "exercises", len(store.Exercises()), "plans", len(store.Plans()))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})
	users.Register(mux, us, log)
	fitness.Register(mux, store, log)

	// The UI, when built in (-tags webembed). Unknown API paths stay a JSON
	// 404 instead of falling through to index.html.
	// TODO(#5): pick the SPA by Host once a second tool exists.
	if fitnessweb.Dist != nil {
		mux.HandleFunc("GET /api/", func(w http.ResponseWriter, r *http.Request) {
			httpx.WriteError(w, http.StatusNotFound, "not found")
		})
		mux.Handle("GET /", httpx.SPA(fitnessweb.Dist))
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
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
		log.Info("listening", "addr", addr)
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
