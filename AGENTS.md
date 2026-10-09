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
| Anything, when Xander says "write it" / "just do it" | Write it, explain what and why, and still ask at least one design question. Xander often writes a partial version first; build on it rather than replacing it. |

**Current mode (2026-10-08):** Xander is busy, so the agent writes backend
logic and frontend directly. Xander still makes every architecture and
design decision (propose → decide → record in §4) and reviews; explanations
stay short. Build on any partial code of Xander's instead of replacing it.

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

## 3. Architecture
Decided; the reasons are in `docs/decisions/` (ADRs).
- **Backend:** Go, single module (`go.mod` at root). Each tool is a Go package
  under `tools/<name>/` exposing one registration function; a single binary
  `cmd/home-tools` mounts all tools. One process = lowest memory.
- **Routing:** the Go server dispatches by `Host` header
  (`fitness.<domain>` → fitness, `games.<domain>` → games; the real
  domain is only in the deploy config, never in git).
  API under `/api/...` per host; everything else serves the tool's SPA.
  `internal/hostroute`; a bare IP or unknown subdomain gets a 404, and
  `/healthz` answers on any host. Local dev: `-tool <name>` serves one
  tool on every host (`make run TOOL=fitness`).
- **Frontend:** Svelte + Vite, no SSR, pnpm workspaces. Packages named
  `@home-tools/<name>`. Built assets embedded into the Go binary via `embed`
  (`-tags webembed`). Shared UI lives in `@home-tools/ui` (`web/packages/ui`).
- **Storage:** JSON files under a configurable data dir (e.g. `data/<tool>/`).
  In-memory index loaded at start; mutations written atomically
  (write temp → fsync → rename); files are loaded as strictly as API input.
- **Network:** LAN only. Local DNS (UniFi DNS records, one per tool)
  resolves `<tool>.<domain>` to the server's LAN IP. No port forwarding,
  no tunnel. HTTPS: Caddy in Docker with a Let's Encrypt wildcard cert via
  Cloudflare DNS-01, proxying to Go on plain HTTP (`deploy/`).

### Layout
    cmd/home-tools/        main.go: open stores, mount tools, serve (flags -addr, -data)
    cmd/fitness-import/    exercise catalog importer (-fix-metrics)
    internal/              shared Go: jsonfile, httpx, users (shared profiles), hostroute
    tools/fitness/         Go package: model, rules, store, API (+ its AGENTS.md)
    web/packages/ui/       @home-tools/ui: tokens, base CSS, shared components, profiles, api/router helpers
    web/apps/fitness/      @home-tools/fitness (Vite SPA) + embed.go
    web/DESIGN.md          UI rules and shared building blocks
    deploy/                Caddy image (HTTPS), compose + ZimaOS app files, setup steps
    .github/workflows/     CI: checks, then publish images to GHCR
    docs/decisions/        ADRs (index in its README)
    docs/adding-a-tool.md  checklist for the next tool
    data/                  runtime JSON (gitignored)
    tools/games/           Go package: game servers via the Docker API (socket proxy)
    web/apps/games/        @home-tools/games (Vite SPA) + embed.go

Go module: `github.com/XanderOnGithub/home-tools` (Go 1.27). In Go,
"@home-tools/fitness" is an *import path*
(`github.com/XanderOnGithub/home-tools/tools/fitness`), not an npm scope; the
`@home-tools/*` names apply to the web packages. `go.mod` has `ignore ./web` so
`./...` never walks node_modules.

### Commands
    make run     go run ./cmd/home-tools -tool $(TOOL)  (API on :8080; TOOL=fitness default)
    make web     fitness UI dev server on :5173 (proxies /api, /images to :8080)
    pnpm --dir web install | check | build   web deps, type-check, production build
    make build   → bin/home-tools (UI built in via -tags webembed)
    docker build -t home-tools .   production image (CI publishes it; see deploy/)
    make check   vet + test (-race) + gofmt check (run before committing)
    go run ./cmd/fitness-import -data data/fitness -users data/users   import exercises + photos (idempotent)
      -fix-metrics   also update metrics of existing imported exercises to the current rules

## 4. Decision Log
Status: ✅ decided · 🟡 proposed (awaiting Xander) · ⬜ open

| # | Decision | Status | Notes |
|---|----------|--------|-------|
| 1 | Agent code-writing policy (§1) | ✅ | 2026-10-08 update: Xander makes architecture/design decisions; the agent writes backend logic and frontend. |
| 2 | Access: LAN only, no remote/tunnel | ✅ | Gym is at home. |
| 3 | Go + Svelte/Vite, no SSR, no meta-framework | ✅ | Learning goal. |
| 4 | JSON-file storage, no database | ✅ | |
| 5 | Single Go module + single binary; tools are packages mounted by a main server, host-based routing | ✅ | 2026-10-07. Lowest memory; tools still isolated as packages. |
| 6 | Frontend: one SPA per tool + shared `@home-tools/ui` | ✅ | 2026-10-07. Static files, so server memory ≈ same; smaller per-subdomain bundles, clean boundaries. |
| 7 | Fitness JSON layout | ✅ | Per-user dirs (`users/<id>/…`); exercises + plans shared (`created_by`); **one file per session** named by ISO start time (`users/<id>/sessions/<start>.json`). Exercise declares tracked `Metrics`; `Set` is flat numbers; muscle `Activation` is a sparse map in (0,1]. In-progress session = zero `EndedAt` + `omitzero`. `Bodyweight` exercises: weight optional (0 = BW only); negative values rejected (no assisted lifts yet). Exercises require ≥1 muscle activation (even stretches). Plans: `PlanExercise{exercise_id, suggested_sets, rest_sec}` — hints only (#35). Rules enforced in `validate.go`; catalog-reference checks belong to the store. Model: `tools/fitness/model.go`. |
| 8 | TLS on LAN: Let's Encrypt wildcard `*.<domain>` via Cloudflare DNS-01 | ✅ | 2026-10-08. Real certs (no CA to install on phones), nothing exposed, subdomain names stay out of CT logs. Scoped Cloudflare token in gitignored `deploy/.env`; domain also only there. Rejected: plain HTTP (no Wake Lock on phones), local CA (install on every phone). |
| 9 | Reverse proxy: Caddy (Docker, custom build with `caddy-dns/cloudflare`) on :443 only → Go on plain HTTP | ✅ | 2026-10-08. Caddy renews certs; Go needs no TLS code or new dependency (`autocert` can't do DNS-01). Port 80 left to the ZimaOS dashboard. Config: `deploy/`. |
| 10 | Auth: none; Netflix-style profile picker | ✅ | 2026-10-08. Trusted LAN; family members aren't a security boundary. UI remembers the chosen profile; API takes the user ID in the URL (`/api/users/{id}/...`). A PIN can be added later as a field. |
| 11 | Exercise data: free-exercise-db (Unlicense, 876 exercises, 2 photos each) | ✅ | 2026-10-07. Import: `cmd/fitness-import` + `tools/fitness/fedb.go`. Primary→1.0, secondary→0.5; stretching/cardio/static holds→duration, distance cardio→duration+distance (#36), else reps+weight; bodyweight-style equipment→weight optional. Never overwrites existing IDs. Rejected: Gym Visual dataset (media needs own license), wger (AGPL/per-entry CC). |
| 12 | Muscle diagram: `body-highlighter` (npm, MIT, framework-agnostic, zero deps) | 🟡 | Verify when building the UI. Its region names differ from our `Muscle` values; one frontend map translates (e.g. shoulders→front+back deltoids, lats/middle_back→upper-back). |
| 13 | Units: store metric (kg, m, s), unit in field name; per-user display preference, UI converts | ✅ | 2026-10-07. Server never converts. |
| 14 | JSON keys are snake_case (`weight_kg`, `duration_sec`) | ✅ | 2026-10-07. TS types mirror them. |
| 15 | Store holds its write lock across the disk write (writes fully serialized) | ✅ | 2026-10-07. Household traffic; a few ms of blocked reads beats disk/memory ordering bugs. Revisit only if measured. |
| 16 | Nothing is deleted: Exercise, Plan, Session all have `archived`; archiving = a normal save | ✅ | 2026-10-07. Lists return archived items (history needs names); UI hides them from pickers. |
| 17 | Exercise `equipment: []Equipment` (enum, empty = none) for UI filtering; enums validated against `All*` lists | ✅ | 2026-10-07. Final value lists will be aligned with the chosen exercise data source (#11). |
| 18 | `Muscle` enum = free-exercise-db's 17 names (snake_case); Exercise gains `category`, `level`, `instructions`, `images`, `source` | ✅ | 2026-10-07. Lossless import; diagram mapping lives in the frontend. |
| 19 | Exercise photos copied to `data/fitness/images/<path>` (~100 MB), served by the Go binary | ✅ | 2026-10-07. Works offline on the LAN. |
| 20 | API request bodies decoded strictly: unknown fields, trailing data, >1 MiB → 400 | ✅ | 2026-10-08. A typo'd key (`weigth_kg`) must fail, not silently save a set without weight. Frontend sends exactly the model's fields. `httpx.DecodeJSON`. |
| 21 | Fitness keeps no profile data of its own: sessions live in `users/<id>/sessions/` keyed by the shared profile ID (#24). **No weight on User:** body weight is an optional `body_weight_kg` on `Session`; "current weight" = latest logged. Height/birthday dropped until a screen needs them | ✅ | 2026-10-08. One source of truth; weight history comes free. |
| 22 | Profile `color` enum (green, blue, orange, purple) = avatar background **and** UI accent. Avatar shape: an organic SVG blob: a circle whose radius follows layered waves of 3, 4 and 5 bumps (random strength and angle), normalized to the same 30% spread so none look round or broken, smoothed with Catmull-Rom curves, computed in the frontend from the profile ID (`features/profiles/blob`), so it is never stored and is always the same for a person. Face: two plain dark oval eyes (same in light/dark mode; white eyes with pupils were rejected as creepy), per-ID placement/size/tilt and timing (`blobFace`); eyes blink, glance around as a pair, and the body leans slightly with them; with reduced motion the face stays still. No initial, no emoji. Component: `features/profiles/profile-avatar`. A flat 2D avatar maker (face shape/eyes/mouth parts) is a later feature | ✅ | 2026-10-08. Preset, contrast-checked palettes instead of free hex: derived colors fail contrast. |
| 23 | UI foundation: CSS custom-property tokens (raw palette → semantic layer), font Figtree (bundled via `@fontsource/figtree`, OFL, Latin subset only, weights 400–800; system font fallback; rejected: M PLUS Rounded 1c, too rounded), `rem` type scale, 4px spacing scale; light/dark via `prefers-color-scheme` + per-device override; WCAG 2.2 AA; 44px touch targets; every interactive element defines rest/hover/pressed/focus/disabled/loading | ✅ | 2026-10-08. Neutrals, type, spacing shared by all tools; only the accent varies (per profile). Red only for errors/destructive actions. Rules in `web/DESIGN.md`. |
| 24 | Profiles are shared by all tools: `internal/users` (`User{id, name, color, units, archived}`, files `data/users/<id>.json`, `GET/PUT /api/users`). Tools key their own data by user ID and check it against this store | ✅ | 2026-10-08. Pick once, same people everywhere. Units live here (recipes need them too). Remembering the chosen profile across subdomains needs a cookie on `.<domain>` (localStorage is per subdomain). |
| 25 | Shared profile gains optional `birthday` (date, `omitzero`), asked in "Add profile" | ✅ | 2026-10-08. Not fitness-specific; other tools may use it. |
| 26 | Fitness "workout profile" `users/<id>/fitness.json`: `goal` (strength / muscle / endurance / general), `height_m`, `schedule` (weekday → plan ID; missing = rest), `weight_prompt_skipped` (ISO week, e.g. `2026-W41`). Onboarding = height → current weight, both required (weight: "a rough estimate is fine"), in the profile's units (ft + in / lb, or cm / kg); schedule can be set later. "Add profile" asks units, default imperial. `goal` is optional and not asked (2026-10-08: "a tool, not a product") | ✅ | 2026-10-08. No profile file yet = show onboarding. Weekly target is derived from the schedule (count of workout days), not stored. |
| 27 | A routine is a fixed weekly schedule (Mon = Upper, Wed = rest…), not a rotation. Home's "Up next" = today's plan | ✅ | 2026-10-08. Easier to understand; missed days aren't carried over. |
| 28 | Body weight is a per-user log `users/<id>/weights.json` (`[{date, weight_kg}]`, ≤1 entry per day). Weekly check-in card on home: input pre-filled with the last weight; Skip records nothing and hides the card for that ISO week. `body_weight_kg` removed from `Session` | ✅ | 2026-10-08. Supersedes the weight part of #21: one source of truth; skipping never invents a measurement. |
| 29 | Manage profiles = a mode of the picker ("Manage profiles" / "Done"): tiles open an edit dialog (same form as Add; ID never changes on rename). "Remove" archives after an inline confirm; archived profiles are listed in manage mode with Restore | 🟡 | 2026-10-08. Agent's call (Xander delegated); review. |
| 30 | Plans are one shared household list (anyone creates/edits); each person's routine (`schedule` in their `fitness.json`) picks which plan on which day. Plans page = "Your routine" planner + the shared list; create/edit is its own page (`/plans/new`, `/plans/<id>`) | ✅ | 2026-10-08. Confirms #7/#27. A page, not a dialog: picking from 876 exercises needs room on phones. |
| 31 | Deployment: one Docker image (root `Dockerfile`: pnpm build → static Go build with `-tags webembed` → distroless), run with Caddy via `deploy/compose.yaml`; data is a bind-mounted host folder (`DATA_DIR`, e.g. under `/var/lib/casaos_data/.media/Vault/`), never in the image | ✅ | 2026-10-08. Same as the game servers: data stays plain files on the host. Without the tag the binary serves no UI (dev uses Vite), so `make check` needs no web build. Embed package: `web/apps/fitness/embed.go`. |
| 32 | Images built by GitHub Actions on push to `main` (checks first), multi-arch (amd64 + arm64, cross-compiled, no emulation), published **public** on GHCR (`ghcr.io/xanderongithub/home-tools`, `…/home-tools-caddy`, tags `latest` + `sha-<commit>`); ZimaOS installs via its compose form as two apps (its importer keeps one service per app): `deploy/zimaos-home-tools.yaml` (publishes :8080) + `deploy/zimaos-caddy.yaml` (proxies to `host.docker.internal:8080`). Caddyfile baked into the Caddy image | ✅ | 2026-10-08. Fits how other apps are installed; server never builds. No secrets in images: domain + token are env vars in the form. |
| 33 | Vocabulary: **Plan** = a named list of exercises (was "Routine"; Go `Plan`, `/api/plans`, `fitness/plans/`, `plan_id` on sessions). **Routine** = a person's weekly schedule of plans (JSON key stays `schedule`). **Workout** (UI) = **Session** (code) | ✅ | 2026-10-08. Matches how Xander thinks about it. Renamed everywhere with no data migration (nothing to keep yet). |
| 34 | Starting and discarding workouts: home's Today card has "Other workout" (or "Start a workout" when nothing is scheduled) → dialog: any plan or an **empty workout**; workout mode has "+ Add exercise" (catalog picker, appended at the end; removable until a set is logged). An in-progress workout can be **discarded** from the Today card (trash icon → inline confirm) = archived (#16) | ✅ | 2026-10-08. Xander: archive, not delete. |
| 35 | Rest between sets: optional `rest_sec` per plan exercise (0 = default 60 s, shown as the field's placeholder; max 3600), a hint like `suggested_sets`; typed as seconds in the plan editor. In workout mode ±15 s changes that exercise's rest for the rest of the workout. Rest end: vibrate (Android) + a Web Audio chime (iOS, unlocked by the ✓ tap). Timed sets are entered as min + sec boxes, distance in km/mi per profile units | ✅ | 2026-10-08. Phone number pads have no ":" key, hence seconds / two boxes instead of "1:30". |
| 36 | Cardio (category `cardio`) is one effort, not sets: workout mode starts it with one row (no set number), no rest timer, "+ Add interval" for more; the plan editor hides its rest field. Distance cardio (running, cycling, rowing, elliptical, skating, walking; `fedbDistance`) tracks duration + **optional** distance (Set rule: distance 0 = not measured); stair/rope/prowler time only. `SaveExercise` refuses metric changes that would invalidate logged sets. `fitness-import -fix-metrics` re-applies the import metric rules to existing free-exercise-db exercises (stop the server first) | ✅ | 2026-10-08. Fixes existing data (server) without overwriting hand edits to anything but metrics. |
| 37 | Game server manager v1: status, start/stop/restart, live logs (players + console later). Servers are config files `data/games/servers/<id>.json` (`{id, name, game: minecraft\|valheim, container}`), one per server like every other store. Docker via its HTTP API with `net/http` over a unix socket (no SDK), through **wollomatic/socket-proxy** (own ZimaOS app; allowlist: container inspect/logs GET, start/stop/restart POST; socket in `/run/home-tools-docker`, not on a share). Logs stream as **Server-Sent Events** (last 200 lines, then live). Stop/restart wait up to 60 s for the world to save. Local dev: `-tool <name>`; host routing in `internal/hostroute` | ✅ | 2026-10-08. Xander chose file config, the proxy (the raw socket is root on the server and the app has no auth), SSE and the small v1. |
| 38 | Games v2: same look as fitness (shared `AppShell` in `@home-tools/ui`, nav links per tool, no phone tab bar with one link). Cards: game icon (own drawings in the accent color, not the games' logos), name, status, "Up 2 h · 2 of 20 online"; the whole card opens the server page; a ⋯ menu top-right holds Start or Stop + Restart (confirm dialog). Server page: header, Players, Log. **Players online:** Minecraft via **RCON** (`list`; password in the server file as `rcon_password`, never sent to browsers), Valheim via Steam **A2S**; address = a `query` field per server | ✅ | 2026-10-08. Xander chose RCON over Server List Ping (it also opens the door to console commands), a query field over guessing from Docker, and the menu top-right. |

Every decision goes in this table. Decisions that shape the architecture
also get an ADR in `docs/decisions/` (index and template in its README;
ADRs exist for #3–#6, #8–#10, #24, #31–#33, #37).

## 5. Tool briefs (scope, not specs)
### Fitness (`fitness.<domain>`): mostly CRUD
- **Profile:** shared name, color, units, birthday (#24); fitness adds
  height, a routine and a weight log (#26, #28).
- **Exercise:** name, primary/secondary muscles, instructions, media.
  Later: guidance per goal (strength: low reps/high weight vs endurance/hypertrophy).
- **Session (workout):** user, start/end, entries `[{exercise, sets}]`;
  a set holds whichever metrics the exercise tracks.
- **Plan:** a template listing exercises (+ suggested sets); does *not*
  enforce sets/reps/weight. Starting a session from a plan pre-fills it.
- **Routine:** per person, which plan on which weekday (#33).
- **Progress:** derived from sessions (and snapshots) for charts per exercise/user.

### Game Server Manager (`games.<domain>`): v1 built (#37), see `tools/games/AGENTS.md`
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
- Web layout: one folder per component or module, never a loose file.
  Folders are kebab-case; Svelte files are PascalCase (Svelte components
  must be capitalized when used: `<ProfilePicker />`).
  Generic UI (knows no domain) and anything two apps share lives in
  `@home-tools/ui` (`web/packages/ui/src/<module>/`); an app's own
  generic pieces go in its `src/components/<name>/`.
  `src/features/<feature>/` = everything for one feature (its components,
  api, types, helpers), e.g. `features/profiles/profile-picker/`:
  `ProfilePicker.svelte` + `index.ts` (re-export; Vite can't resolve an
  `index.svelte`) + `types.ts`/`utils.ts` only when needed. Import the
  folder via `@/...` (= `src/`; alias in vite.config + tsconfig).
- Commits: small, imperative mood; one concern per commit.
- Branches: `main` always works (deployable at any moment). Work happens on
  short-lived branches named `feature/<thing>`, `fix/<thing>` or
  `experiment/<thing>`; merge to `main` when `make check` and
  `pnpm --dir web check` pass, then delete the branch. Days, not weeks.
- Before writing code, read the nearest README/AGENTS.md in that directory.
  Each tool has its own `tools/<name>/AGENTS.md` (fitness: data layout,
  rules, API, how to add a feature). New tool: `docs/adding-a-tool.md`.

## 7. Status (2026-10-08)
Live at `https://fitness.<domain>` on ZimaOS (#31, #32). History of how it
got here: `git log`.
- **Done:** host routing (`-tool` for dev); `@home-tools/ui`; shared profiles (picker, manage mode, blob avatars, accent
  colors); fitness onboarding (height, weight); home (Today card, this
  week, weekly weight check-in); plans + routine planner; workout mode
  (pre-filled sets, rest timer + chime, photo demo + how-to, cardio with
  optional distance, add exercises, empty workouts, discard); history
  list; HTTPS (Caddy, DNS-01); CI images on GHCR; docs (ADRs, tool guide,
  adding-a-tool).
- **Games (#37, #38, ADR 0011):** server cards with status and players,
  ⋯ menu for start/stop/restart, server page with players and live log;
  runs on the real server via the socket proxy. Players online tested
  against fakes; needs `query` (+ RCON setup for Minecraft) on the server.
- **Not built yet:** Progress (charts per exercise); games console
  commands, permissions, server config editor.
- **Open decisions:** #12 (muscle diagram library), #29 (review the
  profile management flow).
- **Next:** set up player queries on the server (RCON for Minecraft).

## 8. Improvements (later, not urgent)
- Avatar maker (#22): flat 2D avatars from SVG parts on the profile color.
- Muscle recovery map (own screen): per-muscle fatigue computed from recent
  sessions × exercise `activation`, decaying over days; "needs rest" view.
- Workout mode: a mute toggle for the rest chime, if it turns out to be
  annoying in a shared gym.
- Profile picker → app: a smooth wipe transition in the chosen person's
  accent color when a profile is tapped (respect reduced motion: fade or
  instant). Likely the View Transitions API or a full-screen accent
  overlay that sweeps across, then reveals home.
- Games, Valheim players: its Steam query (A2S) doesn't answer on the
  real server, so it shows no players. Fallback: follow its log, where a
  join is logged as "Got character ZDOID from <character name>" (plus a
  periodic "Connections N" count). Less reliable than a query (restarts,
  missed leaves). Use made-up names in tests and docs, never real players'.
- Games, Minecraft console: send commands over the RCON connection that
  already works (`say`, `whitelist add`, …) from the server page; needs a
  decision on who may run what (§5 permissions).
- Fitness Progress: charts per exercise from logged sessions (weight ×
  reps, volume, time/distance for cardio); the last big piece of the
  fitness brief (§5).
