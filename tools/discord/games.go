package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// The bot reaches game servers through the games tool's HTTP API (decision
// #41), never by importing its package: either tool can be removed without
// touching the other, and the bot gets the same rules (one action at a
// time, input checks) as the web page.

// GameServer is what the games API says about one server (its serverView).
type GameServer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Game     string `json:"game"` // minecraft, valheim
	Archived bool   `json:"archived"`
	State    *struct {
		Running   bool      `json:"running"`
		Status    string    `json:"status"`
		StartedAt time.Time `json:"started_at"`
	} `json:"state"`
	Error   string `json:"error"`
	Players *struct {
		Online int      `json:"online"`
		Max    int      `json:"max"`
		Names  []string `json:"names"`
	} `json:"players"`
	PlayersError string `json:"players_error"`
	Whitelist    bool   `json:"whitelist"` // whitelist routes work
	Busy         string `json:"busy"`      // start, stop, restart: running now
}

// WhitelistResult is the games API's answer to a whitelist change.
type WhitelistResult struct {
	Output        string `json:"output"`
	RestartNeeded bool   `json:"restart_needed"`
}

// Games is the part of the games API the bot uses. An interface so tests
// can fake it.
type Games interface {
	Servers(ctx context.Context) ([]GameServer, error)
	Restart(ctx context.Context, id string) error
	Whitelist(ctx context.Context, id string) ([]string, error)
	EditWhitelist(ctx context.Context, id, player string, add bool) (WhitelistResult, error)
}

// GamesError is an error answer from the games API. Its message is worded
// for people (the API promises that), so the bot can show it as is.
type GamesError struct {
	Status  int
	Message string
}

func (e *GamesError) Error() string { return e.Message }

// errGamesDown means the games API couldn't be reached at all.
var errGamesDown = errors.New("couldn't reach the games tool")

// GamesHTTP calls the games API over HTTP. Routing is by Host header
// (internal/hostroute), so a call to this same binary on localhost just
// needs a Host starting with "games.".
type GamesHTTP struct {
	base   string // e.g. http://127.0.0.1:8080
	host   string // Host header, e.g. games.internal
	client *http.Client
}

// NewGamesHTTP returns a client for the games API at base. Restarts wait
// for the game to save (up to ~90 s), so timeouts are per call, not here.
func NewGamesHTTP(base string) *GamesHTTP {
	return &GamesHTTP{base: strings.TrimSuffix(base, "/"), host: "games.internal", client: &http.Client{}}
}

func (g *GamesHTTP) Servers(ctx context.Context) ([]GameServer, error) {
	var out []GameServer
	return out, g.call(ctx, http.MethodGet, "/api/servers", &out)
}

func (g *GamesHTTP) Restart(ctx context.Context, id string) error {
	return g.call(ctx, http.MethodPost, "/api/servers/"+url.PathEscape(id)+"/restart", nil)
}

func (g *GamesHTTP) Whitelist(ctx context.Context, id string) ([]string, error) {
	var out struct {
		Players []string `json:"players"`
	}
	err := g.call(ctx, http.MethodGet, "/api/servers/"+url.PathEscape(id)+"/whitelist", &out)
	return out.Players, err
}

func (g *GamesHTTP) EditWhitelist(ctx context.Context, id, player string, add bool) (WhitelistResult, error) {
	method := http.MethodPut
	if !add {
		method = http.MethodDelete
	}
	var out WhitelistResult
	err := g.call(ctx, method, "/api/servers/"+url.PathEscape(id)+"/whitelist/"+url.PathEscape(player), &out)
	return out, err
}

// call does one request and decodes a 2xx JSON answer into out (if not
// nil). Error answers become *GamesError; no answer is errGamesDown.
func (g *GamesHTTP) call(ctx context.Context, method, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, g.base+path, nil)
	if err != nil {
		return err
	}
	req.Host = g.host
	res, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", errGamesDown, err)
	}
	defer res.Body.Close()
	body := io.LimitReader(res.Body, 1<<20)
	if res.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(body).Decode(&e) != nil || e.Error == "" || res.StatusCode == http.StatusInternalServerError {
			e.Error = "the games tool had a problem (" + res.Status + ")"
		}
		return &GamesError{Status: res.StatusCode, Message: e.Error}
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(body).Decode(out); err != nil {
		return fmt.Errorf("games %s %s: %w", method, path, err)
	}
	return nil
}
