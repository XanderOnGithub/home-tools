# deploy

HTTPS for home-tools on the LAN (decisions #8, #9): Caddy in Docker
terminates TLS on :443 with a Let's Encrypt wildcard certificate for
`*.<domain>`, obtained via Cloudflare DNS (DNS-01), and proxies to the Go
server on plain HTTP. Nothing is exposed to the internet: no ports
forwarded, the certificate is proven through Cloudflare's API.

## Setup (once)
1. **Cloudflare token:** My Profile → API Tokens → Create Token →
   "Edit zone DNS" template, zone = your domain. Keep it separate from
   any other token (e.g. a DDNS updater's).
2. **Config:** `cp .env.example .env`, fill in `DOMAIN` and the token.
   `.env` is gitignored.
3. **Local DNS (UniFi):** Policy Table → DNS Record, type Host (A):
   `fitness.<domain>` → the server's LAN IP. Give the server a fixed IP.
   Add one record per tool (not a wildcard, so other public subdomains
   still resolve normally).
4. **Run:** `docker compose up -d --build` in this folder, with
   home-tools running on the host (`-addr :8080`).
5. **Check:** `docker compose logs caddy` shows "certificate obtained";
   open `https://fitness.<domain>` on a phone on the home Wi-Fi.

## Gotchas
- Phones must use the UniFi box for DNS: turn off Android "Private DNS";
  iCloud Private Relay may bypass it in Safari.
- Certificates live in the `caddy_data` volume; deleting it means
  re-issuing (Let's Encrypt has rate limits).
