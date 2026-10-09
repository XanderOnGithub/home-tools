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
| `players.go` | Who's online: `queryPlayers` picks RCON or A2S by game; parses Minecraft's `list`. |
| `rcon.go` | Minecraft RCON client (Source RCON protocol over TCP). |
| `a2s.go` | Steam A2S query client (UDP), as Valheim answers it. |
| `handlers.go` | HTTP API, incl. the SSE log stream. `serverView` is what browsers see (no password, no query address). |
| `snapshot.go` | Stale-while-revalidate cache of all server views, so `GET /api/servers` answers instantly. |
| `games_test.go`, `players_test.go` | Rules, store, demux, RCON, A2S, and every handler against fakes. |

UI: `web/apps/games` (Svelte, on `@home-tools/ui` incl. the shared
`AppShell`): `features/servers/` has the API module, `server-list` (home:
cards), `server-card` (icon, name, status, players; the whole card links
to the server page), `server-menu` (⋯ popover: Start or Stop, Restart;
confirm dialog), `server-page` (header + players + log), `player-list`,
`server-status`, `game-icon`, `log-view` (SSE client), `live-servers` (one
shared list for every page, polled every 5 s while visible, so moving
between pages never refetches).

## Configuration
One file per server in `data/games/servers/<id>.json`:

    {"id": "minecraft", "name": "Minecraft", "game": "minecraft", "container": "<docker ps name>",
     "query": "192.168.3.3:25575", "rcon_password": "<server.properties rcon.password>"}
    {"id": "valheim", "name": "Valheim", "game": "valheim", "container": "<docker ps name>",
     "query": "192.168.3.3:2457"}

`container` must be a valid Docker name (it goes into API URLs). `query`
(optional) is where to ask who's online, as seen from the Home Tools
container: Minecraft's RCON port (needs `enable-rcon=true` and a password
in `server.properties`, and `rcon_password` here), Valheim's Steam query
port (game port + 1). The password and address never leave the server
(`serverView`). The server reads these files at startup; there's no
editor in the UI yet (`PUT /api/servers/{id}` works).

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
| `GET /api/servers` | Every server + live `state` (`running`, `status`, `started_at`) or an `error` worded for people, and for running ones with a `query`: `players` (`online`, `max`, `names`) or `players_error`. Served from a snapshot (instant); a snapshot older than 4 s triggers one background refresh (Docker 3 s, games 1 s timeouts), so data lags by at most one poll. Only the first request after startup or a config change waits. |
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
- **Never make the list wait on games:** a game that doesn't answer costs
  its whole query timeout. That's why the list comes from the snapshot
  and lookups happen in the background.
- **Valheim names:** A2S usually reports a count without names; the UI
  says so instead of showing blanks.
- **RCON is a password-protected admin port:** reachable from the LAN is
  fine, never forward it on the router.
- **Later (scope, not built):** console commands (RCON), who-may-do-what
  permissions, a server config editor, an HTTP API the Discord bot can
  call (AGENTS.md §5).
