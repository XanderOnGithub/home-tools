# home-tools: one image with the server (UI built in) and the importer.
#   docker build -t home-tools .
# Data is never in the image: mount a host folder at /data (deploy/).

# Build stages run on the builder's own platform and cross-compile (the UI
# is platform-independent, Go cross-compiles natively), so multi-arch
# builds need no emulation.

# 1. Web UI
FROM --platform=$BUILDPLATFORM node:24-alpine AS web
RUN npm install -g pnpm@10.28.0
WORKDIR /src/web
# Manifests first, so dependencies are cached until they change.
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
COPY web/packages/ui/package.json packages/ui/
COPY web/apps/fitness/package.json apps/fitness/
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# 2. Go binaries (static, so the runtime image needs no libc)
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS go
ARG TARGETOS TARGETARCH
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
COPY --from=web /src/web/apps/fitness/dist web/apps/fitness/dist
RUN go build -tags webembed -trimpath -ldflags="-s -w" -o /out/home-tools ./cmd/home-tools \
 && go build -trimpath -ldflags="-s -w" -o /out/fitness-import ./cmd/fitness-import

# 3. Runtime: no shell, no package manager; CA certs for the importer.
FROM gcr.io/distroless/static-debian12
COPY --from=go /out/ /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["home-tools", "-addr", ":8080", "-data", "/data"]
