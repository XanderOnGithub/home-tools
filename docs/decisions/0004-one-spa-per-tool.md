# 0004. One SPA per tool

- Status: Accepted, 2026-10-07 (log #6)

## Context
Each tool has its own subdomain (ADR 0003). The UIs share a look (tokens,
buttons, the profile picker) but little else.

## Decision
- Each tool gets its own Vite app: `web/apps/<name>`, package
  `@home-tools/<name>`, in one pnpm workspace (`web/`).
- Shared UI goes into `@home-tools/ui` (`web/packages/ui`) **once a second
  app needs it**; until then it lives in the fitness app
  (`src/components`, `tokens.css`, `app.css`).
- Built apps are embedded in the Go binary (`-tags webembed`, ADR 0008).

## Consequences
- Small bundles per subdomain; a tool's UI can be rewritten alone.
- Static files, so the server's memory barely changes per app.
- Extracting `@home-tools/ui` is real work when tool #2 starts: move
  tokens, `app.css`, generic components and the profile features, then
  import them from both apps.

## Rejected
- One SPA for all tools: every tool ships every other tool's code, and
  subdomain routing would happen in the client.
