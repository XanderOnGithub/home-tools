# home-tools: one image with the server (UI built in) and the importer.
#   docker build -t home-tools .
# Data is never in the image: mount a host folder at /data (deploy/).

# 1. Web UI
FROM node:24-alpine AS web
RUN npm install -g pnpm@10.28.0
WORKDIR /src/web
# Manifests first, so dependencies are cached until they change.
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
COPY web/apps/fitness/package.json apps/fitness/
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# 2. Go binaries (static, so the runtime image needs no libc)
FROM golang:1.27-alpine AS go
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
COPY --from=web /src/web/apps/fitness/dist web/apps/fitness/dist
RUN CGO_ENABLED=0 go build -tags webembed -trimpath -ldflags="-s -w" -o /out/home-tools ./cmd/home-tools \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/fitness-import ./cmd/fitness-import

# 3. Runtime: no shell, no package manager; CA certs for the importer.
FROM gcr.io/distroless/static-debian12
COPY --from=go /out/ /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["home-tools", "-addr", ":8080", "-data", "/data"]
