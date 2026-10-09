package games

import (
	"context"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"
)

// Who joined and left, read from each server's log (decision #40).
// On demand and incremental: whenever the server list is refreshed
// (only while someone looks), the log lines written since the last read
// are parsed. The first read after startup catches up from the
// container's start. Nothing runs while nobody's looking.

const (
	maxEvents       = 50              // per server, newest kept
	activityTimeout = 5 * time.Second // per refresh; a long catch-up continues next time
)

// Event is one join or leave.
type Event struct {
	At     time.Time `json:"at"`
	Player string    `json:"player"`
	Kind   string    `json:"kind"` // "join" or "leave"
}

// activity holds the log-derived state of every server, by server ID.
type activity struct {
	mu      sync.Mutex
	servers map[string]*serverActivity
}

// serverActivity is one server's state, for the container run that
// started at startedAt. Its own mutex is held during a read, so two
// refreshes never parse the same lines twice.
type serverActivity struct {
	mu        sync.Mutex
	startedAt time.Time            // the run this state belongs to
	readUpTo  time.Time            // timestamp of the last line parsed
	online    map[string]time.Time // name → when they joined
	peers     map[string]string    // Valheim: peer ID → character name, for leaves
	events    []Event              // oldest → newest, ≤ maxEvents
}

func (a *activity) of(id string) *serverActivity {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.servers == nil {
		a.servers = make(map[string]*serverActivity)
	}
	sa := a.servers[id]
	if sa == nil {
		sa = &serverActivity{}
		a.servers[id] = sa
	}
	return sa
}

// update reads srv's new log lines (st must be its current, running
// state) and returns who's online and the recent events, newest first.
// A failed or cut-short read keeps what it got; the next one continues.
func (a *activity) update(ctx context.Context, d *Docker, srv Server, st State) (online []string, events []Event, err error) {
	sa := a.of(srv.ID)
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if !sa.startedAt.Equal(st.StartedAt) { // a new run: nobody's on yet
		sa.startedAt, sa.readUpTo = st.StartedAt, st.StartedAt
		sa.online, sa.peers = map[string]time.Time{}, map[string]string{}
	}
	ctx, cancel := context.WithTimeout(ctx, activityTimeout)
	defer cancel()
	err = d.LogsSince(ctx, srv.Container, sa.readUpTo, func(at time.Time, text string) error {
		if !at.After(sa.readUpTo) { // `since` is inclusive; skip what we've seen
			return nil
		}
		sa.readUpTo = at
		sa.parse(srv.Game, at, text)
		return nil
	})
	return sa.snapshot(), sa.recent(), err
}

// recentEvents returns srv's events without reading the log (for a
// stopped server: the history stays visible).
func (a *activity) recentEvents(id string) []Event {
	sa := a.of(id)
	sa.mu.Lock()
	defer sa.mu.Unlock()
	return sa.recent()
}

func (sa *serverActivity) snapshot() []string {
	names := make([]string, 0, len(sa.online))
	for n := range sa.online {
		names = append(names, n)
	}
	// Longest online first, like a server's own list.
	slices.SortFunc(names, func(x, y string) int { return sa.online[x].Compare(sa.online[y]) })
	return names
}

func (sa *serverActivity) recent() []Event {
	out := slices.Clone(sa.events)
	slices.Reverse(out)
	return out
}

func (sa *serverActivity) join(at time.Time, name string) {
	if _, ok := sa.online[name]; ok {
		return
	}
	sa.online[name] = at
	sa.add(Event{At: at, Player: name, Kind: "join"})
}

func (sa *serverActivity) leave(at time.Time, name string) {
	if _, ok := sa.online[name]; !ok {
		return
	}
	delete(sa.online, name)
	sa.add(Event{At: at, Player: name, Kind: "leave"})
}

func (sa *serverActivity) add(e Event) {
	sa.events = append(sa.events, e)
	if len(sa.events) > maxEvents {
		sa.events = slices.Delete(sa.events, 0, len(sa.events)-maxEvents)
	}
}

// Log lines (after Docker's timestamp). Minecraft (vanilla, Paper):
//
//	[18:00:00] [Server thread/INFO]: Steve joined the game
//	[18:00:00 INFO]: Steve (formerly known as Alex) left the game
//
// Valheim:
//
//	Got character ZDOID from Ragnhild : 1234567890:1     (spawned; "0:0" = died)
//	Destroying abandoned non persistent zdo 1234567890:5 owner 1234567890   (left)
//	Connections 0 ZDOS:...                               (nobody's connected)
var (
	mcJoinLeave = regexp.MustCompile(`\]: (\w{1,16})(?: \(formerly known as \w+\))? (joined|left) the game\s*$`)
	vhCharacter = regexp.MustCompile(`Got character ZDOID from (.+) : (-?\d+):(-?\d+)\s*$`)
	vhAbandoned = regexp.MustCompile(`Destroying abandoned non persistent zdo -?\d+:\d+ owner (-?\d+)\s*$`)
	vhConns     = regexp.MustCompile(`Connections (\d+) ZDOS:`)
)

func (sa *serverActivity) parse(game Game, at time.Time, text string) {
	switch game {
	case GameMinecraft:
		if m := mcJoinLeave.FindStringSubmatch(text); m != nil {
			if m[2] == "joined" {
				sa.join(at, m[1])
			} else {
				sa.leave(at, m[1])
			}
		}
	case GameValheim:
		if m := vhCharacter.FindStringSubmatch(text); m != nil {
			if m[2] == "0" { // died; the respawn logs a new ZDOID
				return
			}
			sa.peers[m[2]] = m[1]
			sa.join(at, m[1])
		} else if m := vhAbandoned.FindStringSubmatch(text); m != nil {
			if name, ok := sa.peers[m[1]]; ok {
				delete(sa.peers, m[1])
				sa.leave(at, name)
			}
		} else if m := vhConns.FindStringSubmatch(text); m != nil {
			if n, _ := strconv.Atoi(m[1]); n == 0 { // a missed leave can't outlive this
				for name := range sa.online {
					sa.leave(at, name)
				}
				clear(sa.peers)
			}
		}
	}
}
