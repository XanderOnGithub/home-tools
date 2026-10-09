// Package games manages the household's game servers (Minecraft, Valheim),
// which run as Docker containers on the home server: status, start/stop/
// restart and live logs (decision #37). Which containers count as game
// servers is configured in files, one per server (data/games/servers/).
package games

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"slices"
	"strconv"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// Server is one game server, stored as servers/<id>.json. Container is the
// Docker container's name as `docker ps` shows it.
type Server struct {
	ID        string `json:"id"`
	Name      string `json:"name"` // shown in the UI, e.g. "Minecraft"
	Game      Game   `json:"game"`
	Container string `json:"container"`
	// Query is where to ask who's online, "host:port": Minecraft's RCON
	// port, Valheim's Steam query port. Empty = no player info.
	Query string `json:"query,omitempty"`
	// RCONPassword logs in to Minecraft's RCON. It stays on the server:
	// the API never sends it to a browser (see serverView).
	RCONPassword string `json:"rcon_password,omitempty"`
	Archived     bool   `json:"archived,omitempty"`
}

// Game says what's running, for icons now and game-specific features
// (players, console) later.
type Game string

const (
	GameMinecraft Game = "minecraft"
	GameValheim   Game = "valheim"
)

var AllGames = []Game{GameMinecraft, GameValheim}

// ErrInvalid marks data that breaks a rule (HTTP 400).
var ErrInvalid = errors.New("invalid")

// containerName is Docker's own rule for names. Checking it also keeps the
// name safe to put in a Docker API URL path.
var containerName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// Validate checks s's own rules: a safe ID, a name, a known game, a valid
// container name, and a usable query address (Minecraft's needs a password).
func (s Server) Validate() error {
	if s.Query != "" {
		host, port, err := net.SplitHostPort(s.Query)
		if n, perr := strconv.Atoi(port); err != nil || host == "" || perr != nil || n < 1 || n > 65535 {
			return fmt.Errorf("%w server %s: query must be host:port, got %q", ErrInvalid, s.ID, s.Query)
		}
	}
	switch {
	case !jsonfile.ValidID(s.ID):
		return fmt.Errorf("%w server: bad ID %q", ErrInvalid, s.ID)
	case s.Name == "":
		return fmt.Errorf("%w server %s: missing name", ErrInvalid, s.ID)
	case !slices.Contains(AllGames, s.Game):
		return fmt.Errorf("%w server %s: unknown game %q", ErrInvalid, s.ID, s.Game)
	case !containerName.MatchString(s.Container):
		return fmt.Errorf("%w server %s: bad container name %q", ErrInvalid, s.ID, s.Container)
	case s.Game == GameMinecraft && s.Query != "" && s.RCONPassword == "":
		return fmt.Errorf("%w server %s: Minecraft's query (RCON) needs rcon_password", ErrInvalid, s.ID)
	case s.Game != GameMinecraft && s.RCONPassword != "":
		return fmt.Errorf("%w server %s: rcon_password is only for Minecraft", ErrInvalid, s.ID)
	}
	return nil
}
