package games

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"
	"time"
)

// Fuzz tests for everything that parses bytes from outside: game servers'
// replies (RCON, Steam A2S), their logs, and Docker's log stream. `make
// check` runs only the seeds below (fast, like any test); to search for
// new failures, run one for a while:
//
//	go test ./tools/games -fuzz FuzzRCONRead -fuzztime 30s
//
// Anything it finds is saved under testdata/fuzz/ and becomes a seed.

func FuzzRCONRead(f *testing.F) {
	packet := func(id int32, body string) []byte {
		var b bytes.Buffer
		binary.Write(&b, binary.LittleEndian, int32(10+len(body)))
		binary.Write(&b, binary.LittleEndian, id)
		binary.Write(&b, binary.LittleEndian, int32(0))
		b.WriteString(body + "\x00\x00")
		return b.Bytes()
	}
	f.Add(packet(1, "There are 0 of a max of 20 players online:"))
	f.Add(packet(-1, ""))
	f.Add([]byte{0xff, 0xff, 0xff, 0x7f}) // a huge size
	f.Fuzz(func(t *testing.T, b []byte) {
		_, body, err := rconRead(bytes.NewReader(b))
		if err == nil && len(body) > rconMaxSize {
			t.Errorf("body of %d bytes is over the packet limit", len(body))
		}
	})
}

func FuzzA2S(f *testing.F) {
	f.Add([]byte("\x11Valheim\x00map\x00valheim\x00Valheim\x00\x00\x00\x03\x0a"))
	f.Add([]byte("\x02\x00Ragnhild\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01Bjorn\x00"))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) {
		if p, err := parseA2SInfo(b); err == nil && (p.Online > 255 || p.Max > 255) {
			t.Errorf("counts out of a byte's range: %+v", p)
		}
		if names := parseA2SPlayers(b); len(names) > 255 {
			t.Errorf("%d players from a one-byte count", len(names))
		}
	})
}

func FuzzMinecraftReplies(f *testing.F) {
	f.Add("There are 2 of a max of 20 players online: alice, bob")
	f.Add("There are 1 out of maximum 10 players online. §6Steve (069a79f4-44e9-4726-a5be-fca90e38aaf5)")
	f.Add("There are 2 whitelisted player(s): alice, bob")
	f.Add("There are no whitelisted players")
	f.Fuzz(func(t *testing.T, out string) {
		if p, uuids, err := parseMinecraftList(out); err == nil {
			for _, n := range p.Names {
				if n == "" {
					t.Error("an empty player name")
				}
			}
			plain := colorCode.ReplaceAllString(out, "") // names may carry § color codes
			for name := range uuids {
				if !strings.Contains(plain, name) {
					t.Errorf("UUID for %q, who isn't in the reply", name)
				}
			}
		}
		if players, err := parseWhitelist(out); err == nil {
			for _, n := range players {
				if n == "" {
					t.Error("an empty whitelisted name")
				}
			}
		}
	})
}

func FuzzActivity(f *testing.F) {
	f.Add(true, "[12:00:00] [Server thread/INFO]: Steve joined the game")
	f.Add(true, "[12:05:00] [Server thread/INFO]: Steve left the game")
	f.Add(false, "10/09/2026 13:54:44: Got character ZDOID from Ragnhild : 1234567890:1")
	f.Add(false, "Destroying abandoned non persistent zdo 1234567890:5 owner 1234567890")
	f.Add(false, "Connections 0 ZDOS:462399  sent:0 recv:0")
	f.Fuzz(func(t *testing.T, minecraft bool, text string) {
		game := GameValheim
		if minecraft {
			game = GameMinecraft
		}
		sa := &serverActivity{online: map[string]time.Time{}, peers: map[string]string{}}
		at := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
		// Every line twice: a join is only ever counted once.
		for _, line := range strings.Split(text, "\n") {
			sa.parse(game, at, line)
			sa.parse(game, at, line)
		}
		if len(sa.events) > len(sa.online)+2*strings.Count(text, "\n")+2 {
			t.Errorf("%d events from %q", len(sa.events), text)
		}
		for name := range sa.online {
			if name == "" {
				t.Error("an empty player name is online")
			}
		}
	})
}

func FuzzDemux(f *testing.F) {
	chunk := func(stream byte, s string) []byte {
		h := []byte{stream, 0, 0, 0, 0, 0, 0, 0}
		binary.BigEndian.PutUint32(h[4:], uint32(len(s)))
		return append(h, s...)
	}
	f.Add(append(chunk(1, "hello\n"), chunk(2, "oops\n")...))
	f.Add([]byte{1, 0, 0, 0, 0xff, 0xff, 0xff, 0xff}) // a 4 GiB chunk that isn't there
	f.Fuzz(func(t *testing.T, b []byte) {
		out, _ := io.ReadAll(demux(bytes.NewReader(b)))
		if len(out) > len(b) {
			t.Errorf("%d bytes out of %d in", len(out), len(b))
		}
	})
}
