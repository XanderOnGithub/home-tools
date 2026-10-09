package games

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// apiVersion pins the Docker Engine API (Docker 20.10+ speaks it). The
// socket proxy's allowlist matches this prefix (deploy/zimaos-docker-proxy.yaml).
const apiVersion = "/v1.41"

// stopTimeoutSec gives a server time to save its world before Docker kills
// it. Minecraft and Valheim both save on a clean stop.
const stopTimeoutSec = 60

// ErrNotFound means Docker has no container with that name.
var ErrNotFound = errors.New("container not found")

// Docker talks to the Docker Engine API with plain net/http: the API is
// just HTTP (here over a unix socket), so no SDK dependency is needed.
type Docker struct {
	client *http.Client
	base   string // "http://docker" over a socket; a test server's URL in tests
}

// NewDocker returns a client for the Docker API at the unix socket path
// (in production, the socket proxy's filtered socket, decision #37).
func NewDocker(socket string) *Docker {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socket)
		},
	}
	// No client timeout: log streams stay open for as long as someone
	// watches. Each call sets its own deadline through its context.
	return &Docker{client: &http.Client{Transport: transport}, base: "http://docker"}
}

// State is what the UI shows about a container.
type State struct {
	Running   bool      `json:"running"`
	Status    string    `json:"status"` // Docker's word: running, exited, restarting…
	StartedAt time.Time `json:"started_at,omitzero"`
	tty       bool      // logs are a raw stream instead of multiplexed
}

// Inspect returns the container's state.
func (d *Docker) Inspect(ctx context.Context, container string) (State, error) {
	resp, err := d.do(ctx, http.MethodGet, "/containers/"+url.PathEscape(container)+"/json", nil)
	if err != nil {
		return State{}, err
	}
	defer resp.Body.Close()
	var body struct {
		State struct {
			Status    string    `json:"Status"`
			Running   bool      `json:"Running"`
			StartedAt time.Time `json:"StartedAt"`
		} `json:"State"`
		Config struct {
			Tty bool `json:"Tty"`
		} `json:"Config"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return State{}, fmt.Errorf("inspect %s: %w", container, err)
	}
	st := State{Running: body.State.Running, Status: body.State.Status, tty: body.Config.Tty}
	if body.State.Running {
		st.StartedAt = body.State.StartedAt
	}
	return st, nil
}

// Action is something a container can be told to do.
type Action string

const (
	ActionStart   Action = "start"
	ActionStop    Action = "stop"
	ActionRestart Action = "restart"
)

// ing is the action as a word for messages: "restarting".
func (a Action) ing() string {
	switch a {
	case ActionStart:
		return "starting"
	case ActionStop:
		return "stopping"
	case ActionRestart:
		return "restarting"
	}
	return string(a)
}

// Do starts, stops or restarts the container. Stop and restart wait up to
// stopTimeoutSec for a clean shutdown. Starting a running container (or
// stopping a stopped one) is not an error.
func (d *Docker) Do(ctx context.Context, container string, a Action) error {
	path := "/containers/" + url.PathEscape(container) + "/" + string(a)
	if a != ActionStart {
		path += fmt.Sprintf("?t=%d", stopTimeoutSec)
	}
	resp, err := d.do(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Logs streams the container's output: the last tail lines, then new
// lines as they come, until ctx is cancelled or the container stops.
// Each line goes to line (without its newline).
func (d *Docker) Logs(ctx context.Context, container string, tail int, line func(string) error) error {
	q := url.Values{"follow": {"1"}, "stdout": {"1"}, "stderr": {"1"}, "tail": {fmt.Sprint(tail)}}
	return d.logs(ctx, container, q, line)
}

// LogsSince reads the lines written after since, each with Docker's
// timestamp, and returns when it has read them all (no following).
func (d *Docker) LogsSince(ctx context.Context, container string, since time.Time, line func(at time.Time, text string) error) error {
	q := url.Values{"stdout": {"1"}, "stderr": {"1"}, "timestamps": {"1"},
		"since": {fmt.Sprintf("%d.%09d", since.Unix(), since.Nanosecond())}}
	return d.logs(ctx, container, q, func(l string) error {
		// "2026-10-09T18:00:00.123456789Z text"
		stamp, text, _ := strings.Cut(l, " ")
		at, err := time.Parse(time.RFC3339Nano, stamp)
		if err != nil {
			return nil // not a timestamped line (shouldn't happen); skip it
		}
		return line(at, text)
	})
}

// logs reads the container's log with the query q, line by line.
func (d *Docker) logs(ctx context.Context, container string, q url.Values, line func(string) error) error {
	st, err := d.Inspect(ctx, container)
	if err != nil {
		return err
	}
	resp, err := d.do(ctx, http.MethodGet, "/containers/"+url.PathEscape(container)+"/logs?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r io.Reader = resp.Body
	if !st.tty {
		r = demux(resp.Body)
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024) // allow long lines (stack traces)
	for sc.Scan() {
		if err := line(sc.Text()); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil && ctx.Err() == nil {
		return fmt.Errorf("logs %s: %w", container, err)
	}
	return nil
}

// demux turns Docker's multiplexed log stream into plain text. Without a
// TTY, Docker prefixes every chunk with an 8-byte header: byte 0 is the
// stream (1 stdout, 2 stderr), bytes 4–7 the chunk length (big-endian).
// Both streams go to the same output here.
func demux(r io.Reader) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		var header [8]byte
		for {
			if _, err := io.ReadFull(r, header[:]); err != nil {
				pw.CloseWithError(eofIsDone(err))
				return
			}
			size := int64(binary.BigEndian.Uint32(header[4:]))
			if _, err := io.CopyN(pw, r, size); err != nil {
				pw.CloseWithError(eofIsDone(err))
				return
			}
		}
	}()
	return pr
}

// eofIsDone treats the stream ending as a clean finish.
func eofIsDone(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// do sends one API request. Docker's errors come back as
// {"message": "..."}; 404 becomes ErrNotFound, and 304 (already started /
// already stopped) counts as success.
func (d *Docker) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, d.base+apiVersion+path, body)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docker %s %s: %w", method, path, err)
	}
	switch {
	case resp.StatusCode < 300, resp.StatusCode == http.StatusNotModified:
		return resp, nil
	case resp.StatusCode == http.StatusNotFound:
		resp.Body.Close()
		return nil, ErrNotFound
	default:
		defer resp.Body.Close()
		var e struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&e)
		return nil, fmt.Errorf("docker %s %s: %s: %s", method, path, resp.Status, e.Message)
	}
}
