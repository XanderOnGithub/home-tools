package games

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
)

// Minecraft console and whitelist over RCON (decision #41). No login:
// anyone on the LAN may run commands (#10). The whitelist routes are the
// stable API for the Discord bot, which decides who may use them; the
// console route is for the server page. Restarting any server is the
// existing POST /api/servers/{id}/restart.

const (
	commandTimeout = 5 * time.Second
	maxCommand     = 1000 // RCON allows ~1.4 KB; no real command is close
)

// playerName is Minecraft's username rule. Checking it also keeps a
// whitelist call from carrying anything but a name into the command.
var playerName = regexp.MustCompile(`^[A-Za-z0-9_]{3,16}$`)

type commandResult struct {
	Output string `json:"output"` // what Minecraft answered, § codes removed
}

// rconServer looks up {id} and checks it's a Minecraft server with RCON
// set up, answering the error itself when it isn't.
func (h *handlers) rconServer(w http.ResponseWriter, r *http.Request) (Server, bool) {
	srv, ok := h.server(w, r)
	if !ok {
		return Server{}, false
	}
	switch {
	case srv.Game != GameMinecraft:
		httpx.WriteError(w, http.StatusNotFound, "only Minecraft servers have a console")
		return Server{}, false
	case srv.Query == "":
		httpx.WriteError(w, http.StatusConflict, "RCON isn't set up for this server (query + rcon_password)")
		return Server{}, false
	}
	return srv, true
}

// run sends one command and answers with its output. 502 = the game
// couldn't be reached (stopped, wrong password…).
func (h *handlers) run(w http.ResponseWriter, r *http.Request, srv Server, command string) {
	ctx, cancel := context.WithTimeout(r.Context(), commandTimeout)
	defer cancel()
	out, err := rconRun(ctx, srv.Query, srv.RCONPassword, command)
	if err != nil {
		h.log.Warn("rcon command", "server", srv.ID, "command", command, "err", err)
		httpx.WriteError(w, http.StatusBadGateway, "couldn't reach the game's console; is it running?")
		return
	}
	h.log.Info("rcon command", "server", srv.ID, "command", command)
	httpx.WriteJSON(w, http.StatusOK, commandResult{Output: strings.TrimSpace(colorCode.ReplaceAllString(out, ""))})
}

// postConsole runs any command: POST {"command": "say hi"} (a leading
// "/" is fine).
func (h *handlers) postConsole(w http.ResponseWriter, r *http.Request) {
	srv, ok := h.rconServer(w, r)
	if !ok {
		return
	}
	var body struct {
		Command string `json:"command"`
	}
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	cmd := strings.TrimPrefix(strings.TrimSpace(body.Command), "/")
	switch {
	case cmd == "":
		httpx.WriteError(w, http.StatusBadRequest, "command is empty")
		return
	case len(cmd) > maxCommand:
		httpx.WriteError(w, http.StatusBadRequest, "command is too long")
		return
	case strings.ContainsAny(cmd, "\r\n\x00"):
		httpx.WriteError(w, http.StatusBadRequest, "one command per request (no line breaks)")
		return
	}
	h.run(w, r, srv, cmd)
}

type whitelist struct {
	Players []string `json:"players"`
}

// getWhitelist lists the whitelisted players.
func (h *handlers) getWhitelist(w http.ResponseWriter, r *http.Request) {
	srv, ok := h.rconServer(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), commandTimeout)
	defer cancel()
	out, err := rconRun(ctx, srv.Query, srv.RCONPassword, "whitelist list")
	if err != nil {
		h.log.Warn("rcon whitelist", "server", srv.ID, "err", err)
		httpx.WriteError(w, http.StatusBadGateway, "couldn't reach the game's console; is it running?")
		return
	}
	players, err := parseWhitelist(out)
	if err != nil {
		h.log.Warn("rcon whitelist", "server", srv.ID, "err", err)
		httpx.WriteError(w, http.StatusBadGateway, "the game gave an unexpected answer")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, whitelist{Players: players})
}

// putWhitelist adds {player}; deleteWhitelist removes them. Both answer
// with Minecraft's own words ("Added Steve to the whitelist", "Player is
// already whitelisted", "That player does not exist"…), since only the
// game knows whether the name is a real account.
func (h *handlers) putWhitelist(w http.ResponseWriter, r *http.Request) {
	h.whitelistChange(w, r, "add")
}

func (h *handlers) deleteWhitelist(w http.ResponseWriter, r *http.Request) {
	h.whitelistChange(w, r, "remove")
}

func (h *handlers) whitelistChange(w http.ResponseWriter, r *http.Request, verb string) {
	srv, ok := h.rconServer(w, r)
	if !ok {
		return
	}
	player := r.PathValue("player")
	if !playerName.MatchString(player) {
		httpx.WriteError(w, http.StatusBadRequest, "not a Minecraft username: 3–16 letters, digits or _")
		return
	}
	h.run(w, r, srv, "whitelist "+verb+" "+player)
}

// Minecraft's `whitelist list`: "There are 2 whitelisted player(s): alice,
// bob" (wording varies a little by version), or "There are no whitelisted
// players".
var whitelistLine = regexp.MustCompile(`There (?:are|is) (\d+|no) whitelisted players?(?:\(s\))?[.:]?\s*(.*)`)

func parseWhitelist(out string) ([]string, error) {
	m := whitelistLine.FindStringSubmatch(colorCode.ReplaceAllString(out, ""))
	if m == nil {
		return nil, errors.New("unexpected answer to whitelist list: " + strings.TrimSpace(out))
	}
	players := []string{}
	for n := range strings.SplitSeq(m[2], ",") {
		if n = strings.TrimSpace(n); n != "" {
			players = append(players, n)
		}
	}
	return players, nil
}
