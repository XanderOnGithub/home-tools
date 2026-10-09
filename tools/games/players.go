package games

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// queryTimeout bounds one player lookup. Games on the LAN answer in
// milliseconds; one that doesn't answer at all costs this much per refresh.
const queryTimeout = 1 * time.Second

// Players is who's online. Max is 0 when unknown (Valheim without an
// answering query).
type Players struct {
	Online int      `json:"online"`
	Max    int      `json:"max"`
	Names  []string `json:"names"`
}

// errNoQuery means the server has no query address configured.
var errNoQuery = errors.New("no query address configured")

// minecraftPlayers asks Minecraft over RCON who's online, with each
// player's UUID (for heads) when the server gives them. `list uuids`
// needs Minecraft 1.13+; anything older gets a plain `list`.
func minecraftPlayers(ctx context.Context, srv Server) (Players, map[string]string, error) {
	if srv.Query == "" {
		return Players{}, nil, errNoQuery
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	out, err := rconRun(ctx, srv.Query, srv.RCONPassword, "list uuids")
	if err != nil {
		return Players{}, nil, err
	}
	if p, uuids, err := parseMinecraftList(out); err == nil {
		return p, uuids, nil
	}
	out, err = rconRun(ctx, srv.Query, srv.RCONPassword, "list")
	if err != nil {
		return Players{}, nil, err
	}
	return parseMinecraftList(out)
}

// valheimMax asks Valheim's Steam query (A2S) for its player slots. Who's
// online comes from the log instead (activity.go): on the real server the
// query often doesn't answer, and it rarely has names.
func valheimMax(ctx context.Context, srv Server) (int, error) {
	if srv.Query == "" {
		return 0, errNoQuery
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	p, err := a2sPlayers(ctx, srv.Query)
	return p.Max, err
}

// Minecraft's `list` answers "There are 2 of a max of 20 players online:
// alice, bob" (vanilla); some servers say "out of maximum". With `list
// uuids` each name is followed by " (<uuid>)". Text may carry § color codes.
var listLine = regexp.MustCompile(`There are (\d+) (?:of a max of|out of maximum) (\d+) players online[.:]?\s*(.*)`)
var colorCode = regexp.MustCompile(`§.`)
var withUUID = regexp.MustCompile(`^(\S+) \(([0-9a-fA-F-]{32,36})\)$`)

// parseMinecraftList returns the players, and name → UUID for those
// listed with one.
func parseMinecraftList(out string) (Players, map[string]string, error) {
	m := listLine.FindStringSubmatch(colorCode.ReplaceAllString(out, ""))
	if m == nil {
		return Players{}, nil, errors.New("unexpected answer to list: " + strings.TrimSpace(out))
	}
	online, _ := strconv.Atoi(m[1])
	maxPlayers, _ := strconv.Atoi(m[2])
	names := []string{}
	uuids := map[string]string{}
	for n := range strings.SplitSeq(m[3], ",") {
		n = strings.TrimSpace(n)
		if u := withUUID.FindStringSubmatch(n); u != nil {
			n = u[1]
			uuids[n] = u[2]
		}
		if n != "" {
			names = append(names, n)
		}
	}
	return Players{Online: online, Max: maxPlayers, Names: names}, uuids, nil
}
