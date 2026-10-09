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
| `docker.go` | Docker Engine API client: plain `net/http` over a unix socket, API `v1.41`; `Inspect`, `Do` (start/stop/restart), `Logs` (follow), `LogsSince` (timestamped, no follow) (+ `demux` for Docker's multiplexed stream). |
| `players.go` | Who's online: Minecraft over RCON (`list uuids`, falls back to `list`); Valheim's slots over A2S. |
| `activity.go` | Joins/leaves and Valheim's online players, parsed from the Docker log on demand (#40). |
| `console.go` | Minecraft console + whitelist over RCON (#41): the routes the Discord bot uses. |
| `heads.go` | Minecraft faces: Mojang profile → skin → 8×8 face + hat, cached. |
| `rcon.go` | Minecraft RCON client (Source RCON protocol over TCP). |
| `a2s.go` | Steam A2S query client (UDP), as Valheim answers it. |
| `handlers.go` | HTTP API, incl. the SSE log stream. `serverView` is what browsers see (no password, no query address). |
| `snapshot.go` | Stale-while-revalidate cache of all server views, so `GET /api/servers` answers instantly. |
| `*_test.go` | Rules, store, demux, RCON, A2S, log parsing, heads (fake Mojang), and every handler against fakes. |

UI: `web/apps/games` (Svelte, on `@home-tools/ui` incl. the shared
`AppShell`): `features/servers/` has the API module, `server-list` (home:
cards), `server-card` (icon, name, status, players; the whole card links
to the server page), `server-menu` (⋯ popover: Start or Stop, Restart;
confirm dialog), `server-page` (header + players + log), `player-list`,
`server-status`, `game-icon`, `activity-list` (joins/leaves by day), `console-view` (Minecraft commands), `log-view` (SSE client), `live-servers` (one
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
port (game port + 1; only for the slot count, players come from its log). The password and address never leave the server
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
| `GET /api/servers` | Every server + live `state` (`running`, `status`, `started_at`) or an `error` worded for people, and for running ones: `players` (`online`, `max` (0 = unknown), `names`; Minecraft needs a `query`) or `players_error`, plus `activity` (joins/leaves, newest first, ≤ 50; also kept for stopped servers). Served from a snapshot (instant); a snapshot older than 4 s triggers one background refresh (Docker 3 s, games 1 s, log read 5 s timeouts), so data lags by at most one poll. Only the first request after startup or a config change waits. |
| `PUT /api/servers/{id}` | Create or replace a server's config (`httpx.PutByID`). |
| `POST /api/servers/{id}/{start\|stop\|restart}` | Do it; answers with the new state. 502 = Docker refused / no such container; 503 = Docker not connected. |
| `GET /api/servers/{id}/heads/{name}` | A Minecraft player's face, 8×8 PNG (1-day cache); 404 when there's none. Only players RCON has listed. |
| `POST /api/servers/{id}/console` | Minecraft: `{"command": "say hi"}` → `{"output": "…"}` (§ codes removed; leading `/` fine; no line breaks). 409 = RCON not set up, 502 = game unreachable. |
| `GET /api/servers/{id}/whitelist` | Minecraft: `{"players": [...]}` |
| `PUT` / `DELETE /api/servers/{id}/whitelist/{player}` | Minecraft: add / remove; `{"output"}` is the game's own answer ("Added Steve to the whitelist", "That player does not exist"). 400 = not a valid username. |
| `GET /api/servers/{id}/logs` | Server-Sent Events: the last 200 lines, then live. `data:` = one line; `event: end` = the container stopped; `event: failure` = a message. |

## For the Discord bot (#41)
No login: the bot calls these over the LAN, and decides itself which
Discord users may use which. Stable routes:
`POST /api/servers/{id}/restart` (any game; waits up to a minute),
`GET/PUT/DELETE /api/servers/{id}/whitelist[/{player}]` and
`POST /api/servers/{id}/console` (Minecraft), `GET /api/servers` (status,
players). Errors are `{"error": "…"}` worded for people, safe to show.
Routing is by Host header (`internal/hostroute`): call
`https://games.<domain>/api/...`, or plain HTTP to port 8080 with
`Host: games.<domain>`. A bare IP gets 404.

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
- **Activity is log-derived (#40):** joins/leaves are parsed from lines
  since the last read (`activity.go` has the exact formats). It resets
  when the container restarts and is rebuilt from the container's start
  after Home Tools restarts; Docker log rotation can drop old lines.
  Valheim's join/leave lines were checked against the real server's log;
  "Connections 0" clears anyone missed. Use made-up names in tests and docs.
- **Heads call Mojang** (`sessionserver.mojang.com`, `textures.minecraft.net`)
  from the server, never from browsers. Offline-mode servers get no heads.
- **RCON is a password-protected admin port:** reachable from the LAN is
  fine, never forward it on the router.
- **Later (scope, not built):** a server config editor; the Discord bot
  itself (AGENTS.md §5).
