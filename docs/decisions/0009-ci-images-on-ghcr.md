# 0009. CI publishes public multi-arch images; ZimaOS pulls them

- Status: Accepted, 2026-10-08 (log #32)

## Context
ZimaOS installs apps from a compose form that needs ready-made images; it
can't build from source. The repo is public.

## Decision
- On every push to `main`, GitHub Actions (`.github/workflows/images.yml`)
  runs the same checks as local (`make check`, `pnpm --dir web check`),
  then builds and pushes two images to GHCR:
  `ghcr.io/xanderongithub/home-tools` and `…/home-tools-caddy`, tags
  `latest` and `sha-<commit>`, for amd64 and arm64 (cross-compiled, no
  emulation).
- Both images are public: nothing secret is inside (domain and token are
  environment variables in the ZimaOS form).
- ZimaOS imports two apps (its importer keeps one service per app):
  `deploy/zimaos-home-tools.yaml` and `deploy/zimaos-caddy.yaml`.

## Consequences
- Deploy = merge to `main`, wait for the workflow, update the app in
  ZimaOS. A failing check means nothing is published.
- Roll back by pinning an older `sha-<commit>` tag.
- `main` must always be deployable (it is what gets published).

## Rejected
- Building on the server: the ZimaOS form can't, and it would need the
  source, Node and Go on the server.
- Private images: ZimaOS would need a registry login, for no secrecy
  gained (the source is public).
