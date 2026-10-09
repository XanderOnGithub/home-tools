.PHONY: run web web-games build test vet fmt check

# Local dev serves one tool on every host (localhost has no subdomain).
TOOL ?= fitness
run:
	go run ./cmd/home-tools -tool $(TOOL)

# UI dev servers; run `make run TOOL=<same tool>` alongside.
web:        # fitness on http://localhost:5173
	pnpm --dir web dev:fitness

web-games:  # games on http://localhost:5174
	pnpm --dir web dev:games

# Production binary with the UI built in.
build:
	pnpm --dir web build
	go build -tags webembed -o bin/home-tools ./cmd/home-tools

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

# Run before committing.
check: vet test
	test -z "$$(gofmt -l .)"
