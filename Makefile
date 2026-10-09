.PHONY: run web build test vet fmt check

# Local dev serves one tool on every host (localhost has no subdomain).
TOOL ?= fitness
run:
	go run ./cmd/home-tools -tool $(TOOL)

# Fitness UI dev server (http://localhost:5173); run `make run` alongside.
web:
	pnpm --dir web dev:fitness

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
