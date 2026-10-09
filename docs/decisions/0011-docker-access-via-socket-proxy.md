# 0011. Game servers via the Docker API, through a filtered socket proxy

- Status: Accepted, 2026-10-08 (log #37)

## Context
The games tool must see, start, stop and restart the Minecraft and Valheim
containers and stream their logs. That means the Docker Engine API. The
raw Docker socket is root on the server: whoever can talk to it can run
any container with any host folder mounted. home-tools has no
authentication (ADR 0006), so anyone on the LAN can call its API, and any
bug in it would have that power too.

## Decision
- Talk to the Docker Engine API with Go's `net/http` over a unix socket
  (API `v1.41`). No Docker SDK dependency.
- Never mount the real socket into home-tools. A separate ZimaOS app runs
  `wollomatic/socket-proxy`, which holds the real socket read-only and
  offers a filtered one at `/run/home-tools-docker/docker.sock` (not on a
  network share). Its allowlist: `GET` container inspect and logs, `POST`
  container start, stop and restart. Nothing else gets through, and
  nothing listens on the network.
- Which containers are game servers is configuration:
  `data/games/servers/<id>.json` (`id`, `name`, `game`, `container`).
- Live logs reach the browser as Server-Sent Events: one-way, plain HTTP,
  standard library on both sides (`http.ResponseController` +
  `EventSource`, which reconnects by itself).

## Consequences
- Worst case through home-tools: someone on the LAN stops or restarts any
  container or reads its logs. Annoying, not a takeover.
- One more app to install and keep running; if it's down, the games page
  says Docker is unreachable and fitness is unaffected.
- New Docker features (e.g. `exec` for console commands) need the
  allowlist widened deliberately, and `exec` would deserve its own
  decision (it runs commands inside a container).
- The proxy needs the Docker group id of the server (`user: 65534:<gid>`).

## Rejected
- Mounting `/var/run/docker.sock` into home-tools: simplest, but makes any
  bug or LAN visitor root on the server.
- The proxy on a TCP port: the Docker API reachable from the whole LAN
  unless carefully firewalled.
- WebSocket for logs: needs a library or a hand-written protocol for a
  stream that only flows one way.
