# Adding a tool

Checklist for tool #2 (the game server manager) and every tool after it.
A tool is one Go package + one web app on its own subdomain (ADRs 0003,
0004). Steps marked **first time only** are one-off setup that tool #2 has
to do because fitness was alone until now.

## 0. Decide
- [ ] Scope from the brief (`AGENTS.md` §5), name, subdomain
      (`<name>.<domain>`), what it stores and where (`data/<name>/`).
- [ ] Record the decisions in the log (`AGENTS.md` §4); big ones get an ADR
      (`docs/decisions/`).

## 1. Host routing (done, 2026-10-08)
`internal/hostroute` picks a tool's mux by subdomain; `main.go` builds each
tool's mux with `toolMux` (shared `/api/users`, the tool's API, its UI).
Local dev: `-tool <name>` (`make run TOOL=<name>`).
- [ ] Mount the new tool: `tools.Handle("<name>", toolMux(...))` in
      `main.go`.

## 2. Go package `tools/<name>`
- [ ] `model.go`, `validate.go`, `store.go`, `handlers.go`, following
      `tools/fitness` (read its `AGENTS.md`).
- [ ] `Open(dir string, us *users.Store)` loads `data/<name>/` strictly;
      `Register(mux, store, log)` mounts its API under `/api/…`.
- [ ] Reuse `internal/`: `jsonfile` (atomic writes, `LoadDir`, `ValidID`),
      `httpx` (`DecodeJSON`, `PutByID`, `WriteError`, `ServerError`),
      `users` (key per-person data by profile ID, ADR 0007).
- [ ] Tests (table-driven) for rules, store and handlers.
- [ ] In `main.go`: open its store after users, mount on its host.

## 3. Shared UI package (first time only)
ADR 0004: extract `@home-tools/ui` once a second app needs it.
- [ ] Create `web/packages/ui` and move what isn't fitness-specific out of
      `web/apps/fitness/src`: `tokens.css`, `app.css`, `components/`,
      `features/profiles/`, `router/`, `api/`, `dates/`, `units/`, `ids/`,
      `wake-lock/`.
- [ ] Fitness imports them from `@home-tools/ui`; `pnpm --dir web check`
      still passes; the favicon script still finds the blob code.
- [ ] Move `web/DESIGN.md` references accordingly.

## 4. Web app `web/apps/<name>`
- [ ] Copy fitness's scaffold (`package.json` as `@home-tools/<name>`,
      `vite.config.ts`, `tsconfig*.json`, `svelte.config.js`,
      `index.html`), then `src/` with only what this tool needs.
- [ ] Profile picker + accent from `@home-tools/ui`: the chosen profile is
      already remembered across subdomains (cookie on `.<domain>`).
- [ ] Icon: run the favicon generator with a new seed and accent
      (`scripts/favicon/`), plus the iPhone home-screen icon and title.
- [ ] Dev: a `dev:<name>` script in `web/package.json` and a Makefile
      target.

## 5. Embed and image
- [ ] `web/apps/<name>/embed.go` + `noembed.go` (package `<name>web`, same
      `webembed` build tag as fitness); `main.go` serves it on its host.
- [ ] `Dockerfile`: add the app's `package.json` to the manifest copy step
      and copy its `dist/` into the Go stage.
- [ ] CI needs no change (it builds and checks the whole repo); confirm the
      "images" workflow passes after merging.

## 6. Deploy
- [ ] UniFi: DNS record (Host/A) `<name>.<domain>` → the server's LAN IP.
      Caddy's wildcard certificate already covers it.
- [ ] If the tool needs host access (the game manager needs the Docker
      socket), add the mount to `deploy/zimaos-home-tools.yaml` and record
      why: the Docker socket is root-equivalent on the server.
- [ ] Update the Home Tools app in ZimaOS.

## 7. Docs
- [ ] `tools/<name>/AGENTS.md` (copy the shape of the fitness one).
- [ ] Root `AGENTS.md`: §3 layout, §5 brief, §7 status, §4 decisions.
- [ ] `README.md`: the tool in the list.
