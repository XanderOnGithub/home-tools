# 0001. Go backend, Svelte + Vite SPAs, no SSR

- Status: Accepted, 2026-10-07 (log #3)

## Context
home-tools is a learning project for a web developer learning Go and
systems design, held to production standards. It runs on a small always-on
home server, so memory and CPU matter more than developer convenience.

## Decision
- Backend: Go, standard library first (`net/http` routing, `log/slog`,
  `encoding/json`). New dependencies need a stated reason.
- Frontend: Svelte 5 + Vite + strict TypeScript, built to static files.
  No SSR and no meta-framework (no SvelteKit).

## Consequences
- One small static binary with a small idle footprint.
- The browser does all rendering; the server only serves JSON and files.
  Routing in the UI is a tiny History-API router (`src/router`).
- Learning value: Go idioms (errors as values, implicit interfaces,
  goroutines) are practiced on real code.
- No server-rendered first paint. Irrelevant on a LAN.

## Rejected
- Node/TypeScript backend: familiar, so less to learn; heavier at idle.
- SvelteKit / Next: SSR and a second server runtime for a LAN app that
  doesn't need either.
