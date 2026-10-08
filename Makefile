.PHONY: run build test vet fmt check

run:
	go run ./cmd/home-tools

build:
	go build -o bin/home-tools ./cmd/home-tools

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

# Run before committing.
check: vet test
	test -z "$$(gofmt -l .)"
