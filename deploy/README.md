# deploy

Runs home-tools on ZimaOS: three apps, from images GitHub Actions
publishes to GHCR on every push to `main` (decision #32,
`.github/workflows/images.yml`), plus a public socket-proxy image. The
server only pulls; it never builds.
- **home-tools** (`ghcr.io/xanderongithub/home-tools`, root `Dockerfile`,
  #31): the Go server with the UI built in. Data is a plain host folder
  mounted at `/data`, never inside the container.
- **caddy** (`ghcr.io/xanderongithub/home-tools-caddy`, `deploy/Dockerfile`,
  #8, #9): Caddy + Cloudflare DNS module, Caddyfile baked in. Terminates
  TLS on :443 with a Let's Encrypt wildcard certificate for `*.<domain>`,
  proven through Cloudflare's API, and proxies to home-tools on plain HTTP.
  Nothing is exposed to the internet: no ports forwarded.
- **docker-proxy** (`wollomatic/socket-proxy`, #37): gives the games tool
  a filtered Docker socket that only allows inspecting, reading logs of,
  starting, stopping and restarting containers.

Files: `zimaos-home-tools.yaml`, `zimaos-caddy.yaml`,
`zimaos-docker-proxy.yaml` (three ZimaOS apps: its importer keeps only one
service per app, so Caddy reaches home-tools through the host's :8080 and
the proxy shares its socket through `/run/home-tools-docker`),
`compose.yaml` + `.env.example` (one-file setup over SSH or for a local
test).

## Setup (once)
1. **Images public:** after the first workflow run, on GitHub → your
   profile → Packages → each image → Package settings → Change
   visibility → Public. (No secrets are in them; ZimaOS then needs no
   login.)
2. **Cloudflare token:** My Profile → API Tokens → Create Token →
   "Edit zone DNS" template, zone = your domain. Keep it separate from
   any other token (e.g. a DDNS updater's).
3. **Data:** copy your local `data/` folder's contents (`users/`,
   `fitness/`) into `Vault/Apps/HomeTools` over the network share.
   Or, over SSH, import the exercise catalog fresh (~30 s, ~100 MB):

       docker run --rm -v /var/lib/casaos_data/.media/Vault/Apps/HomeTools:/data \
         --entrypoint fitness-import ghcr.io/xanderongithub/home-tools \
         -data /data/fitness -users /data/users

4. **Install:** ZimaOS → App Store → Custom Install → Import, once per
   app: `zimaos-home-tools.yaml`, then `zimaos-caddy.yaml` with the domain
   and token filled in (keep that copy as the gitignored
   `zimaos-caddy.local.yaml`).
5. **Local DNS (UniFi):** Policy Table → DNS Record, type Host (A):
   `fitness.<domain>` → the server's LAN IP (likewise `games.`,
   `discord.`). Give the server a fixed IP.
   One record per tool (not a wildcard, so public subdomains like game
   servers still resolve normally).
6. **Check:** `sudo docker ps` (SSH needs sudo for Docker on ZimaOS);
   `sudo docker logs <caddy container>` shows "certificate obtained";
   open `https://fitness.<domain>` on a phone on the home Wi-Fi.

## Games (game server manager)
Once, over SSH and in ZimaOS:
1. Docker's group id: `stat -c %g /var/run/docker.sock`.
2. Import `zimaos-docker-proxy.yaml` with that number in `user:`.
3. Re-import (or edit) the Home Tools app from `zimaos-home-tools.yaml`:
   it now mounts `/run/home-tools-docker` and starts with `-docker …`.
4. The game containers' names: `sudo docker ps --format '{{.Names}}'`.
5. One file per server in `Vault/Apps/HomeTools/games/servers/`, then
   restart Home Tools:

       minecraft.json  {"id": "minecraft", "name": "Minecraft", "game": "minecraft", "container": "<name from step 4>"}
       valheim.json    {"id": "valheim", "name": "Valheim", "game": "valheim", "container": "<name from step 4>"}

   Players online (optional, see `tools/games/AGENTS.md`): add
   `"query": "<server IP>:<port>"`; for Minecraft also enable RCON in its
   `server.properties` and add `"rcon_password"`.

6. UniFi: DNS record `games.<domain>` → the server's LAN IP.
7. Check: `https://games.<domain>` shows both servers as Running or
   Stopped (not "Unavailable"), and a server's page shows its live log.

## Discord bot
Once:
1. <https://discord.com/developers/applications> → New Application (the
   name is only a fallback; the bot renames itself per server daily) →
   Bot → Reset Token, copy it. Leave every "Privileged Gateway Intent"
   off: the bot needs none.
2. ZimaOS → Home Tools app → settings: environment `DISCORD_TOKEN` = the
   token. Save (the app restarts). Never put the token in a file in git.
3. UniFi: DNS record `discord.<domain>` → the server's LAN IP.
4. Open `https://discord.<domain>`: it should say "Connected as …". Use
   "Add to a Discord server" there (only the permissions it needs).
5. Pick the status channels on that page. People who try `/restart` show
   up under People → "Asked for access": tap Verify.
6. Valheim whitelist (optional): mount the folder holding the Valheim
   server's `permittedlist.txt` into Home Tools, read-write (the
   commented volume in `zimaos-home-tools.yaml`; the folder, not the
   file), then add `"permitted_list": "/valheim/permittedlist.txt"` to
   `games/servers/valheim.json` and restart Home Tools. The game reads
   the list when it starts, so additions apply after a restart.

Features (poll, `/blob`) are switched on in the page's Features tab.
The poll needs "Send Polls" and `/blob` "Attach Files" in their
channels: both are in the invite link; a bot invited before they were
added may need them given to its role in Discord's server settings.

Check the log after the first start (`sudo docker logs <home-tools>`):
"discord connected", then "discord persona" per server. If the avatar
was refused, the log says so and the bot keeps a still image or just
the nickname.

## Update
Merge to `main` → wait for the "images" workflow (GitHub → Actions) →
update the app in ZimaOS (pulls `latest`). Data is untouched. To roll
back, set the image tag to an older `sha-<commit>`.

## Fix exercise metrics after an import-rule change
When a release changes how imported exercises are tracked (e.g. planks
timed, runs with distance), existing exercises keep their old metrics
until you re-apply the rules. Over SSH:

    # 1. Stop the Home Tools app in ZimaOS (it only reads files at startup).
    # 2. Re-apply the rules (only metrics change; prints each fix):
    sudo docker run --rm -v /var/lib/casaos_data/.media/Vault/Apps/HomeTools:/data \
      --entrypoint fitness-import ghcr.io/xanderongithub/home-tools \
      -data /data/fitness -users /data/users -images=false -fix-metrics
    # 3. Start the app again.

An exercise whose logged sets wouldn't fit the new metrics is kept as is
and reported ("kept …").

## Gotchas
- Phones must use the UniFi box for DNS: turn off Android "Private DNS";
  iCloud Private Relay may bypass it in Safari. A VPN on the phone (e.g.
  Google One VPN) sends DNS elsewhere too: "This site can't be reached"
  while `http://<server IP>:8080/healthz` answers means the VPN is on.
- Keep the Caddy data folder: deleting it means re-issuing certificates
  (Let's Encrypt has rate limits).
- The containers run as root so they can write to the CasaOS folders.
- home-tools picks the tool by subdomain, so `http://<server IP>:8080`
  answers 404 ("no tool here"); `/healthz` works on the IP. To debug
  without Caddy: `curl -H "Host: fitness.<domain>" http://<server IP>:8080/`.
