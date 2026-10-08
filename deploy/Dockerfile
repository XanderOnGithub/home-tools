# Caddy with the Cloudflare DNS module (the official image has no DNS
# modules), so it can get a wildcard certificate via DNS-01 (decision #8).
# The Caddyfile is baked in; it only reads env vars, so there is no config
# file to place on the server.
#
# Builder runs on the CI machine's own platform and cross-compiles (Go
# needs no emulation), so multi-arch builds stay fast.
FROM --platform=$BUILDPLATFORM caddy:2.10-builder AS builder
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH xcaddy build --with github.com/caddy-dns/cloudflare

FROM caddy:2.10
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY Caddyfile /etc/caddy/Caddyfile
