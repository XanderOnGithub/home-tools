# 0008. One Docker image; data in a bind-mounted host folder

- Status: Accepted, 2026-10-08 (log #31)

## Context
The server runs ZimaOS, where apps (including the Minecraft and Valheim
servers) are Docker containers whose data sits in plain host folders. The
data must survive image updates and stay easy to back up and edit.

## Decision
- One image from the root `Dockerfile`: pnpm builds the UI, Go builds a
  static binary with the UI embedded (`-tags webembed`) plus the
  `fitness-import` tool, on a distroless base (no shell).
- The data folder is a bind mount (e.g.
  `/var/lib/casaos_data/.media/Vault/Apps/HomeTools` → `/data`), never
  baked into the image.
- Without `-tags webembed` the binary serves only the API (dev uses Vite),
  so `make check` needs no web build.

## Consequences
- Updating = pulling a new image; data is untouched.
- The image is small and has no shell: debugging happens through logs
  and the API, and one-off tasks run as `docker run --entrypoint
  fitness-import …`.
- The container runs as root so it can write to the CasaOS folders.

## Rejected
- A binary on the host: unlike every other app on the server.
- Data inside a Docker volume: harder to see, back up and hand-edit.
