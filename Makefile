.PHONY: run web build test vet fmt check

run:
	go run ./cmd/home-tools

# Fitness UI dev server (http://localhost:5173); run `make run` alongside.
web:
	pnpm --dir web dev:fitness

build:
	go build -o bin/home-tools ./cmd/home-tools

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

# Run before committing.
check: vet test
	test -z "$$(gofmt -l .)"
