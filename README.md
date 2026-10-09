# home-tools

Small self-hosted tools for one household, running on a home server and
reachable only on the home network. **Tools, not a product:** no accounts,
no tracking, no lock-in. All data is plain JSON files you can read, edit
and back up by copying a folder.

## Tools
- **Fitness** (`fitness.<domain>`): plans, a weekly routine per person,
  workout logging built for a phone at the gym (pre-filled sets, rest
  timer, exercise photos and instructions, cardio with time and distance),
  weekly body-weight check-ins. 876 exercises from
  [free-exercise-db](https://github.com/yuhonas/free-exercise-db)
  (public domain).
- **Game servers** (`games.<domain>`): status, start/stop/restart and live
  logs for the household's Minecraft and Valheim servers (Docker
  containers, reached through a filtered socket proxy). Players online and
  console commands are next.

Everyone picks their profile Netflix-style; there are no passwords (it's a
trusted home network).

## Stack
Go (standard library first) serving a JSON API and the web apps, built into
one binary. Svelte 5 + Vite + TypeScript, no SSR. Storage is JSON files,
written atomically. Deployed as a Docker image behind Caddy for HTTPS.
Why each of these: [`docs/decisions/`](docs/decisions/).

## Run it locally
Needs Go 1.27+, Node 24+ and pnpm 10.

    pnpm --dir web install
    go run ./cmd/fitness-import -data data/fitness -users data/users   # catalog + photos, ~30 s
    make run    # API on :8080 (data in ./data); one tool: TOOL=fitness (default) or TOOL=games
    make web    # fitness UI on http://localhost:5173 (proxies the API)
    make web-games   # games UI on http://localhost:5174

Before committing: `make check` and `pnpm --dir web check`.

## Deploy
Push to `main` → GitHub Actions runs the checks and publishes images to
GHCR → update the app on the server. Setup (HTTPS certificate, local DNS,
ZimaOS apps) and day-to-day tasks: [`deploy/README.md`](deploy/README.md).

## Layout
    cmd/home-tools/       the server: loads data, mounts tools, serves the apps
    cmd/fitness-import/   imports the exercise catalog
    internal/             shared Go: profiles (users), JSON files, HTTP helpers
    tools/fitness/        fitness: model, rules, store, API
    tools/games/          game servers: config, Docker API client, API
    web/packages/ui/      shared UI: tokens, styles, components, profiles
    web/apps/fitness/     fitness web app (Svelte)
    web/apps/games/       games web app (Svelte)
    web/DESIGN.md         UI rules: tokens, accessibility, shared pieces
    deploy/               Docker, Caddy, ZimaOS app files
    docs/                 decisions (ADRs), how to add a tool
    data/                 runtime JSON (not in git)

## Working on it
[`AGENTS.md`](AGENTS.md) is the guide for anyone (human or AI) changing the
code: principles, conventions, the decision log and current status. Each
tool has its own `AGENTS.md`. Adding a tool:
[`docs/adding-a-tool.md`](docs/adding-a-tool.md).
