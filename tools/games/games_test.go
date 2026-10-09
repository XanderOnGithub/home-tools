package games

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerValidate(t *testing.T) {
	ok := Server{ID: "minecraft", Name: "Minecraft", Game: GameMinecraft, Container: "big-bear-minecraft"}
	with := func(f func(*Server)) Server { s := ok; f(&s); return s }
	tests := []struct {
		name    string
		s       Server
		wantErr bool
	}{
		{"valid", ok, false},
		{"bad id", with(func(s *Server) { s.ID = "../x" }), true},
		{"no name", with(func(s *Server) { s.Name = "" }), true},
		{"unknown game", with(func(s *Server) { s.Game = "doom" }), true},
		{"no container", with(func(s *Server) { s.Container = "" }), true},
		{"container with slash", with(func(s *Server) { s.Container = "a/../b" }), true},
		{"container with space", with(func(s *Server) { s.Container = "my server" }), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	srv := Server{ID: "valheim", Name: "Valheim", Game: GameValheim, Container: "valheim"}
	if err := s.SaveServer(srv); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveServer(Server{ID: "x"}); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid save: err = %v, want ErrInvalid", err)
	}
	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.Servers(); len(got) != 1 || got[0] != srv {
		t.Errorf("after reopen: %+v", got)
	}
}

// frame builds one multiplexed log chunk: 8-byte header, then the payload.
func frame(stream byte, payload string) []byte {
	h := make([]byte, 8)
	h[0] = stream
	binary.BigEndian.PutUint32(h[4:], uint32(len(payload)))
	return append(h, payload...)
}

func TestDemux(t *testing.T) {
	var in bytes.Buffer
	in.Write(frame(1, "Starting server\nDone (3.2s)!"))
	in.Write(frame(2, "\nWARN low memory\n"))
	got, err := io.ReadAll(demux(&in))
	if err != nil {
		t.Fatal(err)
	}
	if want := "Starting server\nDone (3.2s)!\nWARN low memory\n"; string(got) != want {
		t.Errorf("demux = %q, want %q", got, want)
	}
}

// fakeDocker answers like the Docker API for two containers: "mc" (running,
// multiplexed logs) and "vh" (exited, TTY logs). Anything else is 404.
func fakeDocker(t *testing.T) (*Docker, *[]string) {
	t.Helper()
	var calls []string
	mux := http.NewServeMux()
	inspect := map[string]string{
		"mc":    `{"State":{"Status":"running","Running":true,"StartedAt":"2026-10-08T18:00:00Z"},"Config":{"Tty":false}}`,
		"vh":    `{"State":{"Status":"exited","Running":false,"StartedAt":"2026-10-07T18:00:00Z"},"Config":{"Tty":true}}`,
		"vhrun": `{"State":{"Status":"running","Running":true,"StartedAt":"2026-10-08T18:00:00Z"},"Config":{"Tty":true}}`,
	}
	mux.HandleFunc("GET /v1.41/containers/{name}/json", func(w http.ResponseWriter, r *http.Request) {
		body, ok := inspect[r.PathValue("name")]
		if !ok {
			http.Error(w, `{"message":"No such container"}`, http.StatusNotFound)
			return
		}
		w.Write([]byte(body))
	})
	mux.HandleFunc("POST /v1.41/containers/{name}/{action}", func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.PathValue("action")+" "+r.PathValue("name")+" "+r.URL.RawQuery)
		if _, ok := inspect[r.PathValue("name")]; !ok {
			http.Error(w, `{"message":"No such container"}`, http.StatusNotFound)
			return
		}
		if r.PathValue("action") == "start" && r.PathValue("name") == "mc" {
			w.WriteHeader(http.StatusNotModified) // already running
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /v1.41/containers/{name}/logs", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("follow") != "1" { // a read for activity: timestamped lines since a time
			if q.Get("timestamps") != "1" || q.Get("since") == "" {
				t.Errorf("logs query = %q", r.URL.RawQuery)
			}
			switch r.PathValue("name") {
			case "mc":
				w.Write(frame(1, "2026-10-08T18:05:00.000000001Z [18:05:00] [Server thread/INFO]: Steve joined the game\n"))
			case "vhrun":
				w.Write([]byte("2026-10-08T18:06:00Z 10/08/2026 18:06:00: Got character ZDOID from Ragnhild : 42:1\n"))
			}
			return
		}
		if q.Get("tail") == "" {
			t.Errorf("logs query = %q", r.URL.RawQuery)
		}
		switch r.PathValue("name") {
		case "mc":
			w.Write(frame(1, "line one\nline two\n"))
		case "vh":
			w.Write([]byte("tty line\n"))
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Docker{client: srv.Client(), base: srv.URL}, &calls
}

func newTestHandlers(t *testing.T, docker *Docker) http.Handler {
	t.Helper()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []Server{
		{ID: "minecraft", Name: "Minecraft", Game: GameMinecraft, Container: "mc"},
		{ID: "valheim", Name: "Valheim", Game: GameValheim, Container: "vh"},
		{ID: "gone", Name: "Gone", Game: GameMinecraft, Container: "missing"},
	} {
		if err := store.SaveServer(s); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	Register(mux, store, docker, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return mux
}

func do(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestGetServers(t *testing.T) {
	docker, _ := fakeDocker(t)
	rec := do(newTestHandlers(t, docker), "GET", "/api/servers")
	var got []serverView
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	byID := map[string]serverView{}
	for _, s := range got {
		byID[s.ID] = s
	}
	if s := byID["minecraft"]; s.State == nil || !s.State.Running {
		t.Errorf("minecraft = %+v, want running", s)
	}
	if s := byID["valheim"]; s.State == nil || s.State.Running || s.State.Status != "exited" || !s.State.StartedAt.IsZero() {
		t.Errorf("valheim = %+v, want exited with no start time", s)
	}
	if s := byID["gone"]; s.State != nil || !strings.Contains(s.Error, "no container") {
		t.Errorf("gone = %+v, want a 'no container' error", s)
	}
}

func TestGetServersWithoutDocker(t *testing.T) {
	rec := do(newTestHandlers(t, nil), "GET", "/api/servers")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Docker isn't connected") {
		t.Errorf("status %d, body %s", rec.Code, rec.Body)
	}
	if rec := do(newTestHandlers(t, nil), "POST", "/api/servers/minecraft/start"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("action without docker = %d, want 503", rec.Code)
	}
}

func TestPostAction(t *testing.T) {
	docker, calls := fakeDocker(t)
	h := newTestHandlers(t, docker)
	tests := []struct {
		path string
		want int
	}{
		{"/api/servers/minecraft/start", http.StatusOK}, // 304 from Docker = fine
		{"/api/servers/valheim/stop", http.StatusOK},
		{"/api/servers/valheim/restart", http.StatusOK},
		{"/api/servers/valheim/delete", http.StatusNotFound},
		{"/api/servers/nope/start", http.StatusNotFound},
		{"/api/servers/gone/start", http.StatusBadGateway},
	}
	for _, tt := range tests {
		if rec := do(h, "POST", tt.path); rec.Code != tt.want {
			t.Errorf("POST %s = %d, want %d (%s)", tt.path, rec.Code, tt.want, rec.Body)
		}
	}
	want := []string{"start mc ", "stop vh t=60", "restart vh t=60", "start missing "}
	if strings.Join(*calls, "|") != strings.Join(want, "|") {
		t.Errorf("docker calls = %q, want %q", *calls, want)
	}
}

func TestGetLogs(t *testing.T) {
	docker, _ := fakeDocker(t)
	h := newTestHandlers(t, docker)
	tests := []struct {
		id, want string
	}{
		{"minecraft", "data: line one\n\ndata: line two\n\nevent: end\ndata: \n\n"},
		{"valheim", "data: tty line\n\nevent: end\ndata: \n\n"},
	}
	for _, tt := range tests {
		rec := do(h, "GET", "/api/servers/"+tt.id+"/logs")
		if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
			t.Errorf("%s: Content-Type = %q", tt.id, ct)
		}
		if rec.Body.String() != tt.want {
			t.Errorf("%s: body = %q, want %q", tt.id, rec.Body, tt.want)
		}
	}
}
