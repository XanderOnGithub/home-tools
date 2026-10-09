# Games tool: agent guide

Game server manager at `games.<domain>`: status, start/stop/restart and
live logs for the household's Minecraft and Valheim servers, which run as
Docker containers on the home server (decision #37, ADR 0011). Read the
root `AGENTS.md` first; this file is what's specific to games.

## Where things are
| File | What |
|---|---|
| `model.go` | `Server` (`id, name, game, container`), `Game` enum, `Validate`. |
| `store.go` | Loads `data/games/servers/*.json` at startup; `SaveServer`. |
| `docker.go` | Docker Engine API client: plain `net/http` over a unix socket, API `v1.41`; `Inspect`, `Do` (start/stop/restart), `Logs` (+ `demux` for Docker's multiplexed stream). |
| `handlers.go` | HTTP API, incl. the SSE log stream. |
| `games_test.go` | Rules, store, demux, and every handler against a fake Docker server. |

UI: `web/apps/games` (Svelte, on `@home-tools/ui`): `features/servers/`
has the API module, `server-list` (home), `server-page` (one server + its
log), `server-card` (status + actions), `log-view` (SSE client),
`live-servers` (the list, refreshed every 10 s while visible).

## Configuration
One file per server in `data/games/servers/<id>.json`:

    {"id": "minecraft", "name": "Minecraft", "game": "minecraft", "container": "<docker ps name>"}

`container` must be a valid Docker name (it goes into API URLs). The
server reads these at startup; there's no editor in the UI yet
(`PUT /api/servers/{id}` works).

## Docker access
- The server talks to Docker only through the socket proxy
  (`deploy/zimaos-docker-proxy.yaml`): it allows `GET
  /v1.41/containers/<name>/(json|logs)` and `POST …/(start|stop|restart)`,
  nothing else. A new Docker call means widening that allowlist on purpose.
- `-docker <socket>` points at the proxy's socket. Without it (local dev),
  servers are listed as "Docker isn't connected" and actions answer 503.
- Stop and restart pass `t=60`: Docker waits up to a minute for the game
  to save its world before killing it. The HTTP request waits with it.

## API
| Method + path | Does |
|---|---|
| `GET /api/servers` | Every server + live `state` (`running`, `status`, `started_at`) or an `error` worded for people. Containers are inspected in parallel, 3 s each. |
| `PUT /api/servers/{id}` | Create or replace a server's config (`httpx.PutByID`). |
| `POST /api/servers/{id}/{start\|stop\|restart}` | Do it; answers with the new state. 502 = Docker refused / no such container; 503 = Docker not connected. |
| `GET /api/servers/{id}/logs` | Server-Sent Events: the last 200 lines, then live. `data:` = one line; `event: end` = the container stopped; `event: failure` = a message. |

## Gotchas
- **SSE event names:** never send `event: error`; `EventSource` uses that
  name for its own connection errors. Ours is `failure`.
- **Reconnects repeat the tail:** each (re)connect starts with the last 200
  lines again, so the UI clears its lines on open.
- **Logs end when the container stops**; the UI offers "Reconnect", and the
  server page reconnects by itself after an action.
- **Testing without a server:** `make run TOOL=games` + `make web-games`
  shows the UI with "Docker isn't connected". For a full flow, point
  `-docker` at a fake Docker API on a unix socket (see `games_test.go`'s
  `fakeDocker` for the endpoints it needs).
- **Later (scope, not built):** players online (Minecraft RCON, Valheim
  A2S or log parsing), console commands, who-may-do-what permissions, an
  HTTP API the Discord bot can call (AGENTS.md §5).
