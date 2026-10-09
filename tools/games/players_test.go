package games

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParseMinecraftList(t *testing.T) {
	tests := []struct {
		in    string
		want  Players
		isErr bool
	}{
		{"There are 2 of a max of 20 players online: alice, bob", Players{2, 20, []string{"alice", "bob"}}, false},
		{"There are 0 of a max of 20 players online: ", Players{0, 20, []string{}}, false},
		{"§6There are §c1§6 out of maximum §c10§6 players online.", Players{1, 10, []string{}}, false},
		{"Unknown command", Players{}, true},
	}
	for _, tt := range tests {
		got, err := parseMinecraftList(tt.in)
		if (err != nil) != tt.isErr {
			t.Errorf("%q: err = %v", tt.in, err)
			continue
		}
		if !tt.isErr && (got.Online != tt.want.Online || got.Max != tt.want.Max || !slices.Equal(got.Names, tt.want.Names)) {
			t.Errorf("%q = %+v, want %+v", tt.in, got, tt.want)
		}
	}
}

func TestServerValidateQuery(t *testing.T) {
	mc := Server{ID: "mc", Name: "MC", Game: GameMinecraft, Container: "mc", Query: "192.168.3.3:25575", RCONPassword: "pw"}
	vh := Server{ID: "vh", Name: "VH", Game: GameValheim, Container: "vh", Query: "192.168.3.3:2457"}
	with := func(s Server, f func(*Server)) Server { f(&s); return s }
	tests := []struct {
		name    string
		s       Server
		wantErr bool
	}{
		{"minecraft rcon", mc, false},
		{"valheim a2s", vh, false},
		{"no port", with(mc, func(s *Server) { s.Query = "192.168.3.3" }), true},
		{"bad port", with(mc, func(s *Server) { s.Query = "host:99999" }), true},
		{"no host", with(mc, func(s *Server) { s.Query = ":25575" }), true},
		{"minecraft without password", with(mc, func(s *Server) { s.RCONPassword = "" }), true},
		{"password on valheim", with(vh, func(s *Server) { s.RCONPassword = "pw" }), true},
	}
	for _, tt := range tests {
		if err := tt.s.Validate(); (err != nil) != tt.wantErr {
			t.Errorf("%s: Validate() = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

// fakeRCON answers one connection like Minecraft: auth (id -1 if the
// password is wrong), then `list`.
func fakeRCON(t *testing.T, password, listReply string) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				for {
					id, body, err := rconRead(conn)
					if err != nil {
						return
					}
					switch body {
					case password:
						rconWrite(conn, id, 2, "")
					case "list":
						rconWrite(conn, id, 0, listReply)
					default:
						rconWrite(conn, -1, 2, "")
					}
				}
			}()
		}
	}()
	return l.Addr().String()
}

func TestRCON(t *testing.T) {
	addr := fakeRCON(t, "secret", "There are 1 of a max of 20 players online: xander")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := rconRun(ctx, addr, "secret", "list")
	if err != nil || !strings.Contains(out, "xander") {
		t.Fatalf("rconRun = %q, %v", out, err)
	}
	if _, err := rconRun(ctx, addr, "wrong", "list"); err == nil || !strings.Contains(err.Error(), "wrong password") {
		t.Errorf("wrong password: err = %v", err)
	}
}

// fakeA2S answers like a Steam server that wants a challenge first.
func fakeA2S(t *testing.T) string {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	header := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	challenge := []byte{1, 2, 3, 4}
	go func() {
		buf := make([]byte, 1400)
		for {
			n, from, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			req := buf[:n]
			hasChallenge := strings.HasSuffix(string(req), string(challenge))
			var reply []byte
			switch {
			case !hasChallenge:
				reply = append(append(append([]byte{}, header...), 'A'), challenge...)
			case req[4] == 'T':
				reply = append(append([]byte{}, header...), 'I', 17)
				reply = append(reply, "Starport\x00Valheim\x00valheim\x00Valheim\x00"...)
				reply = binary.LittleEndian.AppendUint16(reply, 0)
				reply = append(reply, 3, 10, 0) // 3 online, max 10, 0 bots
			case req[4] == 'U':
				reply = append(append([]byte{}, header...), 'D', 3)
				for _, name := range []string{"Ragnar", "", "Freya"} {
					reply = append(reply, 0)
					reply = append(reply, name+"\x00"...)
					reply = append(reply, 0, 0, 0, 0, 0, 0, 0, 0)
				}
			}
			conn.WriteTo(reply, from)
		}
	}()
	return conn.LocalAddr().String()
}

func TestA2S(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := a2sPlayers(ctx, fakeA2S(t))
	if err != nil {
		t.Fatal(err)
	}
	if got.Online != 3 || got.Max != 10 || !slices.Equal(got.Names, []string{"Ragnar", "Freya"}) {
		t.Errorf("a2sPlayers = %+v", got)
	}
}

func TestGetServersWithPlayers(t *testing.T) {
	docker, _ := fakeDocker(t)
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rcon := fakeRCON(t, "secret", "There are 2 of a max of 20 players online: alice, bob")
	for _, s := range []Server{
		{ID: "minecraft", Name: "Minecraft", Game: GameMinecraft, Container: "mc", Query: rcon, RCONPassword: "secret"},
		{ID: "valheim", Name: "Valheim", Game: GameValheim, Container: "vh", Query: fakeA2S(t)}, // stopped: not asked
	} {
		if err := store.SaveServer(s); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	Register(mux, store, docker, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := do(mux, "GET", "/api/servers")
	if strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "rcon_password") {
		t.Fatalf("the RCON password reached the browser: %s", rec.Body)
	}
	var got []serverView
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	for _, v := range got {
		switch v.ID {
		case "minecraft":
			if v.Players == nil || v.Players.Online != 2 || !slices.Equal(v.Players.Names, []string{"alice", "bob"}) {
				t.Errorf("minecraft players = %+v (%s)", v.Players, v.PlayersError)
			}
		case "valheim":
			if v.Players != nil {
				t.Errorf("stopped valheim was asked for players: %+v", v.Players)
			}
		}
	}
}

func TestPutServerHidesPassword(t *testing.T) {
	h := newTestHandlers(t, nil)
	body := `{"id":"minecraft","name":"Minecraft","game":"minecraft","container":"mc","query":"10.0.0.1:25575","rcon_password":"secret"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("PUT", "/api/servers/minecraft", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT = %d: %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "10.0.0.1") {
		t.Errorf("PUT echoed secrets: %s", rec.Body)
	}
}
