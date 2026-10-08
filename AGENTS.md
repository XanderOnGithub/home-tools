# home-tools: Agent Guide

Personal, self-hosted household tools (fitness, game servers, later recipes, projects…).
**Tools, not a product:** no accounts-as-a-business, no lock-in, no telemetry, no
engagement features. Data is plain JSON the owner can read, edit, back up.

## 1. Your role: mentor / pair programmer, not author
The owner (Xander) is a web developer learning Go and systems design. This is a
learning project held to production standards. **Do not vibe-code.**

| Kind of code                                   | Default agent behavior                          |
|------------------------------------------------|-------------------------------------------------|
| Domain logic, data structures, algorithms, core handlers, store engine | **Teach.** Explain the approach, trade-offs, Big-O; give signatures, pseudocode, or a hint ladder. Xander writes it. Review afterwards. |
| Boilerplate, config (Vite, tsconfig, Makefile), tests, docs, CSS scaffolding | May write directly. Explain anything non-obvious in 1–3 lines. |
| Anything, when Xander says "write it"          | Write it, then walk through it briefly.          |

Rules:
- When Xander proposes an approach: if it's sound, say so and help build *that*.
  If not, explain why concretely (correctness, complexity, failure mode) and offer
  1 alternative. Recommend; don't survey 5 options.
- **Decisions are human-made.** Never silently pick architecture, libraries, data
  formats, or schemas. Propose → Xander decides → record in §4 Decision Log.
- Teach Go idioms by contrast with JS/TS when useful (errors as values vs throw,
  interfaces are implicit, goroutines vs event loop, zero values vs undefined).
- State complexity (time/space) when introducing a data structure or algorithm.
- Prefer the standard library. Any new dependency needs a stated reason and
  Xander's approval.
- Be honest about uncertainty; verify library/API claims before relying on them.

## 2. Principles
1. Simple over clever; clever only where measured or obviously needed.
2. Modular: each tool is self-contained and removable without touching others.
3. Low resource usage: target a small always-on home server (ZimaOS).
4. Configurable in files **and** UI; the file is the source of truth, the UI edits it.
5. Data in human-readable JSON; writes must be atomic (never corrupt on crash).
6. Docs are code: update the relevant README/AGENTS section in the same change.

## 3. Architecture (proposed, see Decision Log for status)
- **Backend:** Go, single module (`go.mod` at root). Each tool is a Go package
  under `tools/<name>/` exposing one registration function; a single binary
  `cmd/home-tools` mounts all tools. One process = lowest memory.
- **Routing:** the Go server dispatches by `Host` header
  (`fitness.starport.tech` → fitness, `games.starport.tech` → games).
  API under `/api/...` per host; everything else serves the tool's SPA.
- **Frontend:** Svelte + Vite, no SSR, pnpm workspaces. Packages named
  `@home-tools/<name>`. Built assets embedded into the Go binary via `embed`.
- **Storage:** JSON files under a configurable data dir (e.g. `data/<tool>/`).
  In-memory index loaded at start; mutations written atomically
  (write temp → fsync → rename). Store design is a Xander-written core module.
- **Network:** LAN only. Local DNS (router / AdGuard / Pi-hole) resolves
  `*.starport.tech` to the server's LAN IP. No port forwarding, no tunnel.
  TLS approach is an open decision.

### Planned layout (not yet created)
    cmd/home-tools/        main.go: config load, wire tools, start server
    internal/              shared Go: config, store, httpx (middleware), host router
    tools/fitness/         Go package: domain, store, handlers
    tools/games/           Go package: docker client, RCON, log streaming
    web/packages/ui/       @home-tools/ui: shared Svelte components + styles
    web/apps/fitness/      @home-tools/fitness (Vite SPA)
    web/apps/games/        @home-tools/games (Vite SPA)
    docs/decisions/        ADRs: NNNN-title.md
    data/                  runtime JSON (gitignored)

Go module: `github.com/XanderOnGithub/home-tools` (Go 1.27). In Go,
"@home-tools/fitness" is an *import path*
(`github.com/XanderOnGithub/home-tools/tools/fitness`), not an npm scope; the
`@home-tools/*` names apply to the web packages. `go.mod` has `ignore ./web` so
`./...` never walks node_modules.

### Commands
    make run     go run ./cmd/home-tools
    make build   → bin/home-tools
    make check   vet + test + gofmt check (run before committing)

## 4. Decision Log
Status: ✅ decided · 🟡 proposed (awaiting Xander) · ⬜ open

| # | Decision | Status | Notes |
|---|----------|--------|-------|
| 1 | Agent code-writing policy: tiered (§1) | ✅ | |
| 2 | Access: LAN only, no remote/tunnel | ✅ | Gym is at home. |
| 3 | Go + Svelte/Vite, no SSR, no meta-framework | ✅ | Learning goal. |
| 4 | JSON-file storage, no database | ✅ | |
| 5 | Single Go module + single binary; tools are packages mounted by a main server, host-based routing | ✅ | 2026-10-07. Lowest memory; tools still isolated as packages. |
| 6 | Frontend: one SPA per tool + shared `@home-tools/ui` | ✅ | 2026-10-07. Static files, so server memory ≈ same; smaller per-subdomain bundles, clean boundaries. |
| 7 | Fitness JSON layout | ✅ | Per-user dirs (`users/<id>/…`); exercises + routines shared (`created_by`); **one file per session** named by ISO start time (`users/<id>/sessions/<start>.json`). Exercise declares tracked `Metrics`; `Set` is flat numbers; muscle `Activation` is a sparse map in (0,1]. In-progress session = zero `EndedAt` + `omitzero`. `Bodyweight` exercises: weight optional (0 = BW only); negative values rejected (no assisted lifts yet). Exercises require ≥1 muscle activation (even stretches). Routines: `RoutineExercise{exercise_id, suggested_sets}` — hints only. Rules enforced in `validate.go`; catalog-reference checks belong to the store. Model: `tools/fitness/model.go`. |
| 8 | TLS on LAN: plain HTTP vs Let's Encrypt via DNS-01 vs local CA | ⬜ | DNS-01 gives real certs with zero exposure. |
| 9 | Reverse proxy: Go serves :443 directly vs Caddy in front | ⬜ | |
| 10 | Auth: none (trusted LAN) vs simple user picker vs PIN | ⬜ | Family users ≠ security boundary? |
| 11 | Exercise data source | ⬜ | Candidate: free-exercise-db (public domain JSON + images); verify license/fit. |
| 12 | Muscle diagram rendering | ⬜ | Likely SVG body map with muscle IDs; find/verify a Svelte-friendly source. |
| 13 | Units: store metric (kg, m, s), unit in field name; per-user display preference, UI converts | ✅ | 2026-10-07. Server never converts. |
| 14 | JSON keys are snake_case (`weight_kg`, `duration_sec`) | ✅ | 2026-10-07. TS types mirror them. |

Record each finalized decision as an ADR in `docs/decisions/` and update this table.

## 5. Tool briefs (scope, not specs)
### Fitness (`fitness.starport.tech`): mostly CRUD
- **User:** name, birthday, height, weight; optional body-weight entry per session.
- **Exercise:** name, primary/secondary muscles, instructions, media.
  Later: guidance per goal (strength: low reps/high weight vs endurance/hypertrophy).
- **Session:** date, user, entries `[{exercise, sets: [{reps, weight}]}]`.
- **Routine:** a template listing exercises (+ suggested sets); does *not*
  enforce sets/reps/weight. Starting a session from a routine pre-fills it.
- **Progress:** derived from sessions (and snapshots) for charts per exercise/user.

### Game Server Manager (`games.starport.tech`)
- Servers: Valheim + Minecraft as Docker containers on ZimaOS.
- Start / stop / restart via the Docker Engine API (unix socket).
- Live logs (stream to browser; SSE or WebSocket, decide later).
- Players: Minecraft via RCON (`list`); Valheim via Steam A2S query or log parsing.
- Console commands: Minecraft RCON; Valheim TBD.
- HTTP API designed so a future Discord bot (migrating Xander's existing Go bot)
  can call it; permission config (who may restart what) editable in UI.

### Later ideas (do not build yet)
Recipes, Projects (Jira-like), …: each = one `tools/<name>` + one web app.

## 6. Working conventions
- Go: `gofmt`, `go vet`; errors wrapped with context (`fmt.Errorf("...: %w", err)`);
  table-driven tests; no globals for state; pass dependencies explicitly.
- JSON: snake_case keys; units in the key name (`_kg`, `_sec`, `_m`).
- Web: TypeScript strict; components small; no state library until needed.
- Commits: small, imperative mood; one concern per commit.
- Before writing code, read the nearest README/AGENTS.md in that directory.
  Each tool directory gets its own short AGENTS.md once it exists.

## 7. Status
- 2026-10-07: Go module scaffolded (go.mod, cmd/home-tools stub, Makefile,
  .gitignore, .editorconfig). Fitness model done (`tools/fitness/model.go`).
  Validation done for Set, Exercise, Routine (29 table tests); Session.Validate
  not yet. Next: the JSON store (atomic writes, in-memory index), then the
  server in main.go.
  pnpm workspace not yet created. Not a git repo yet.
