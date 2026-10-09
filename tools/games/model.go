// Package games manages the household's game servers (Minecraft, Valheim),
// which run as Docker containers on the home server: status, start/stop/
// restart and live logs (decision #37). Which containers count as game
// servers is configured in files, one per server (data/games/servers/).
package games

import (
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// Server is one game server, stored as servers/<id>.json. Container is the
// Docker container's name as `docker ps` shows it.
type Server struct {
	ID        string `json:"id"`
	Name      string `json:"name"` // shown in the UI, e.g. "Minecraft"
	Game      Game   `json:"game"`
	Container string `json:"container"`
	Archived  bool   `json:"archived,omitempty"`
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

// Validate checks s's own rules: a safe ID, a name, a known game and a
// valid container name.
func (s Server) Validate() error {
	switch {
	case !jsonfile.ValidID(s.ID):
		return fmt.Errorf("%w server: bad ID %q", ErrInvalid, s.ID)
	case s.Name == "":
		return fmt.Errorf("%w server %s: missing name", ErrInvalid, s.ID)
	case !slices.Contains(AllGames, s.Game):
		return fmt.Errorf("%w server %s: unknown game %q", ErrInvalid, s.ID, s.Game)
	case !containerName.MatchString(s.Container):
		return fmt.Errorf("%w server %s: bad container name %q", ErrInvalid, s.ID, s.Container)
	}
	return nil
}
