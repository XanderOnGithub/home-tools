# Caddy with the Cloudflare DNS module (the official image has no DNS
# modules), so it can get a wildcard certificate via DNS-01 (decision #8).
FROM caddy:2.10-builder AS builder
RUN xcaddy build --with github.com/caddy-dns/cloudflare

FROM caddy:2.10
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
