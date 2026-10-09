# 0005. HTTPS on the LAN: Let's Encrypt via Cloudflare DNS-01, Caddy in front

- Status: Accepted, 2026-10-08 (log #8, #9)

## Context
The app is LAN-only (log #2). Phones need HTTPS for the Screen Wake Lock API
(the workout screen must stay on), and secure cookies are good hygiene.
The domain's DNS is on Cloudflare. Port 80 on the server belongs to the
ZimaOS dashboard.

## Decision
- One wildcard certificate `*.<domain>` from Let's Encrypt, proven with a
  DNS-01 challenge through the Cloudflare API (a token limited to editing
  that zone's DNS).
- Caddy terminates TLS on :443 only and proxies to the Go server on plain
  HTTP. Caddy is a custom build with `caddy-dns/cloudflare`
  (`deploy/Dockerfile`); its Caddyfile reads `DOMAIN`, `UPSTREAM` and the
  token from environment variables.
- Local DNS (UniFi) points each `<tool>.<domain>` at the server's LAN IP;
  nothing is published on the internet, no ports are forwarded.

## Consequences
- Real certificates: nothing to install on family phones.
- Caddy renews automatically (the certificate lives in its data folder;
  keep it, Let's Encrypt rate-limits re-issuing).
- The Go server stays TLS-free and dependency-free.
- Phones must use the home DNS: VPNs and "Private DNS" bypass it.
- A wildcard keeps subdomain names out of public certificate logs.

## Rejected
- Plain HTTP: no Wake Lock on phones.
- A local CA: a root certificate to install on every phone.
- TLS in Go (`autocert`): only HTTP-01/TLS-ALPN, which need public
  reachability; DNS-01 would need a new dependency (`lego`).
