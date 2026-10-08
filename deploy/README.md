# deploy

Runs home-tools on ZimaOS with Docker Compose: two containers.
- **home-tools** (image from the root `Dockerfile`, decision #31): the Go
  server with the UI built in. Data is a plain host folder (`DATA_DIR`)
  mounted at `/data`, never inside the container.
- **caddy** (decisions #8, #9): terminates TLS on :443 with a Let's Encrypt
  wildcard certificate for `*.<domain>`, obtained via Cloudflare DNS
  (DNS-01), and proxies to home-tools on plain HTTP. Nothing is exposed to the internet: no ports
forwarded, the certificate is proven through Cloudflare's API.

## Setup (once)
1. **Cloudflare token:** My Profile → API Tokens → Create Token →
   "Edit zone DNS" template, zone = your domain. Keep it separate from
   any other token (e.g. a DDNS updater's).
2. **Config:** `cp .env.example .env`, fill in `DOMAIN`, the token and
   `DATA_DIR`. `.env` is gitignored.
3. **Local DNS (UniFi):** Policy Table → DNS Record, type Host (A):
   `fitness.<domain>` → the server's LAN IP. Give the server a fixed IP.
   Add one record per tool (not a wildcard, so other public subdomains
   still resolve normally).
4. **Data:** either copy your existing `data/` folder into `DATA_DIR`, or
   import the exercise catalog fresh (~30 s, ~100 MB of photos):

       docker compose run --rm --entrypoint fitness-import home-tools \
         -data /data/fitness -users /data/users

5. **Run:** `docker compose up -d --build` in this folder.
6. **Check:** `docker compose logs caddy` shows "certificate obtained";
   open `https://fitness.<domain>` on a phone on the home Wi-Fi.

## Update
`git pull && docker compose up -d --build`. Data is untouched.

## Gotchas
- Phones must use the UniFi box for DNS: turn off Android "Private DNS";
  iCloud Private Relay may bypass it in Safari.
- Certificates live in the `caddy_data` volume; deleting it means
  re-issuing (Let's Encrypt has rate limits).
- The container runs as root so it can write to the CasaOS data folder.
