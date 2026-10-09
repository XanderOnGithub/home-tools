# 0003. One binary; tools are packages; host-based routing

- Status: Accepted, 2026-10-07 (log #5). Host routing not built yet.

## Context
Several tools (fitness, game servers, later recipes…) on one small server.
Each should be removable without touching the others (principle 2), but a
process per tool multiplies memory and deployment work.

## Decision
- One Go module, one binary (`cmd/home-tools`).
- Each tool is a package `tools/<name>` exposing
  `Register(mux, store, log)`; shared code lives in `internal/`
  (`httpx`, `jsonfile`, `users`).
- Requests are routed by `Host`: `fitness.<domain>` → fitness,
  `games.<domain>` → games. Every host gets `/api/users` (ADR 0007).

## Consequences
- One process, one container, one port.
- Tools share a process: a panic in a handler is contained (`net/http`
  recovers it per request), but a store that fails to load at startup
  stops every tool.
- Removing a tool = deleting its package, its web app and one line in
  `main.go`.
- Today there is one tool, so `main.go` mounts everything on one mux and
  serves the fitness UI for every host (`TODO(#5)`). Host routing is the
  first step of tool #2 (`docs/adding-a-tool.md`).

## Rejected
- A binary/container per tool: isolation we don't need, at several times
  the memory and deploy work.
- Path prefixes (`/fitness/…`): subdomains keep each app's URLs and
  cookies clean and let each SPA assume it owns `/`.
