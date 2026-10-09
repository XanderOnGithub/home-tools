package games

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// queryTimeout bounds one player lookup, so a game that doesn't answer
// can't hold up the server list.
const queryTimeout = 2 * time.Second

// Players is who's online. Names may be fewer than Online: Valheim often
// reports a count without names.
type Players struct {
	Online int      `json:"online"`
	Max    int      `json:"max"`
	Names  []string `json:"names"`
}

// errNoQuery means the server has no query address configured.
var errNoQuery = errors.New("no query address configured")

// queryPlayers asks the game itself who's online: Minecraft over RCON
// (`list`), Valheim over Steam's A2S query.
func queryPlayers(ctx context.Context, srv Server) (Players, error) {
	if srv.Query == "" {
		return Players{}, errNoQuery
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	switch srv.Game {
	case GameMinecraft:
		out, err := rconRun(ctx, srv.Query, srv.RCONPassword, "list")
		if err != nil {
			return Players{}, err
		}
		return parseMinecraftList(out)
	case GameValheim:
		return a2sPlayers(ctx, srv.Query)
	}
	return Players{}, errNoQuery
}

// Minecraft's `list` answers "There are 2 of a max of 20 players online:
// alice, bob" (vanilla); some servers say "out of maximum". Text may carry
// § color codes.
var listLine = regexp.MustCompile(`There are (\d+) (?:of a max of|out of maximum) (\d+) players online[.:]?\s*(.*)`)
var colorCode = regexp.MustCompile(`§.`)

func parseMinecraftList(out string) (Players, error) {
	m := listLine.FindStringSubmatch(colorCode.ReplaceAllString(out, ""))
	if m == nil {
		return Players{}, errors.New("unexpected answer to list: " + strings.TrimSpace(out))
	}
	online, _ := strconv.Atoi(m[1])
	maxPlayers, _ := strconv.Atoi(m[2])
	names := []string{}
	for n := range strings.SplitSeq(m[3], ",") {
		if n = strings.TrimSpace(n); n != "" {
			names = append(names, n)
		}
	}
	return Players{Online: online, Max: maxPlayers, Names: names}, nil
}
